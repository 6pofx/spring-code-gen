// Package pomcheck 解析 Maven pom.xml 并检查本地仓库依赖安装情况
package pomcheck

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Dependency Maven 依赖
type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
	Optional   string `xml:"optional"`
	Type       string `xml:"type"`
}

// DependencyMgmt dependencyManagement 中的依赖
type DependencyMgmt struct {
	Dependencies []Dependency `xml:"dependencies>dependency"`
}

// Parent POM 父节点
type Parent struct {
	GroupID      string `xml:"groupId"`
	ArtifactID   string `xml:"artifactId"`
	Version      string `xml:"version"`
	RelativePath string `xml:"relativePath"`
}

// Plugin 构建插件
type Plugin struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// Project 简化的 POM 结构
type Project struct {
	XMLName              xml.Name       `xml:"project"`
	GroupID              string         `xml:"groupId"`
	ArtifactID           string         `xml:"artifactId"`
	Version              string         `xml:"version"`
	Parent               Parent         `xml:"parent"`
	Dependencies         []Dependency   `xml:"dependencies>dependency"`
	DependencyManagement DependencyMgmt `xml:"dependencyManagement"`
	Plugins              []Plugin       `xml:"build>plugins>plugin"`
}

// CheckResult 单个依赖检查结果
type CheckResult struct {
	GroupID       string `json:"groupId"`
	ArtifactID    string `json:"artifactId"`
	Version       string `json:"version"`
	Scope         string `json:"scope"`
	ResolvedByBOM bool   `json:"resolvedByBOM"`
	Installed     bool   `json:"installed"`
	JarPath       string `json:"jarPath"`
	Comment       string `json:"comment"`
}

// Summary 检查汇总
type Summary struct {
	Total     int    `json:"total"`
	Installed int    `json:"installed"`
	Missing   int    `json:"missing"`
	MavenHome string `json:"mavenHome"`
}

// Report 完整检查报告
type Report struct {
	PomPath  string         `json:"pomPath"`
	Summary  Summary        `json:"summary"`
	Results  []CheckResult  `json:"results"`
	Errors   []string       `json:"errors,omitempty"`
}

// getMavenRepoDir 获取 Maven 本地仓库路径
func getMavenRepoDir() string {
	home, _ := os.UserHomeDir()
	// 检查 MAVEN_OPTS 或 settings.xml 中的自定义路径
	// 默认 ~/.m2/repository
	return filepath.Join(home, ".m2", "repository")
}

// CheckPomFile 解析 pom.xml 文件并检查依赖
func CheckPomFile(pomPath string) (*Report, error) {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("读取 pom.xml 失败: %w", err)
	}
	return CheckPomContent(data, pomPath)
}

// CheckPomContent 解析 pom.xml 内容并检查依赖
func CheckPomContent(data []byte, sourceName string) (*Report, error) {
	report := &Report{
		PomPath: sourceName,
	}

	// 解析 XML
	proj, err := parsePOM(data)
	if err != nil {
		return nil, fmt.Errorf("解析 pom.xml 失败: %w", err)
	}

	// 提取属性
	props := extractProperties(data)

	// 处理 BOM 中的版本
	bomVersions := extractBOMVersions(proj)

	// 检查依赖
	mavenRepo := getMavenRepoDir()
	var results []CheckResult
	installedCount := 0

	for _, dep := range proj.Dependencies {
		// 跳过 type=pom 的（如 BOM import 本身）
		if dep.Type == "pom" {
			continue
		}

		result := CheckResult{
			GroupID:    dep.GroupID,
			ArtifactID: dep.ArtifactID,
			Scope:      dep.Scope,
		}

		// 解析版本（处理属性占位符）
		version := resolveVersion(dep.Version, props)
		if version == "" {
			// 尝试从 BOM 取版本
			if v, ok := bomVersions[dep.GroupID+":"+dep.ArtifactID]; ok {
				version = v
				result.ResolvedByBOM = true
			} else {
				// 从 spring-boot-dependencies BOM 找
				version = resolveFromBOMFile(mavenRepo, props, dep.GroupID, dep.ArtifactID)
				if version != "" {
					result.ResolvedByBOM = true
				}
			}
		}

		result.Version = version

		// 检查本地仓库
		if version == "" || strings.Contains(version, "${") {
			// 版本未解析 — 尝试 fallback：在本地仓库中查找该 artifact 的任何版本
			foundVer, found, jarPath := fallbackCheckLocal(mavenRepo, dep.GroupID, dep.ArtifactID)
			if found {
				result.Version = foundVer
				result.Installed = true
				result.JarPath = jarPath
				result.Comment = "版本号由 BOM 管理"
				result.ResolvedByBOM = true
				installedCount++
			} else {
				result.Comment = "本地仓库中未找到该依赖（可能由 BOM 管理，需要先执行 mvn install）"
			}
		} else {
			jarPath, installed := checkLocalJar(mavenRepo, dep.GroupID, dep.ArtifactID, version)
			result.Installed = installed
			result.JarPath = jarPath
			if installed {
				installedCount++
			} else {
				result.Comment = "本地仓库中未找到该 jar"
			}
		}

		results = append(results, result)
	}

	report.Results = results
	report.Summary = Summary{
		Total:     len(results),
		Installed: installedCount,
		Missing:   len(results) - installedCount,
		MavenHome: mavenRepo,
	}

	return report, nil
}

// parsePOM 解析 XML 为 Project 结构
func parsePOM(data []byte) (*Project, error) {
	var proj Project
	if err := xml.Unmarshal(data, &proj); err != nil {
		return nil, err
	}
	return &proj, nil
}

// ParsePOMContent 解析 pom.xml 内容（导出）
func ParsePOMContent(data []byte) (*Project, error) {
	return parsePOM(data)
}

// ParsePOMFile 读取并解析 pom.xml 文件（导出）
func ParsePOMFile(path string) (*Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	return parsePOM(data)
}

// extractProperties 从 XML 原始内容中提取 <properties>
func extractProperties(data []byte) map[string]string {
	props := make(map[string]string)

	// 先用 regex 提取 <properties> 块
	re := regexp.MustCompile(`<properties>(.*?)</properties>`)
	match := re.FindStringSubmatch(string(data))
	if len(match) < 2 {
		return props
	}

	// 提取每个 <key>value</key>
	propRe := regexp.MustCompile(`<([^/]+?)>([^<]+)</\1>`)
	matches := propRe.FindAllStringSubmatch(match[1], -1)
	for _, m := range matches {
		if len(m) >= 3 {
			props[strings.TrimSpace(m[1])] = strings.TrimSpace(m[2])
		}
	}

	return props
}

// resolveVersion 解析版本值，替换 ${property} 占位符
func resolveVersion(version string, props map[string]string) string {
	if version == "" {
		return ""
	}
	result := version

	// 替换 ${xxx} 占位符
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		key := strings.TrimPrefix(strings.TrimSuffix(match, "}"), "${")
		if val, ok := props[key]; ok {
			return val
		}
		// 尝试 project.* 属性
		if strings.HasPrefix(key, "project.") {
			field := strings.TrimPrefix(key, "project.")
			switch field {
			case "version":
				return "${project.version}" // recursive — return as-is
			}
		}
		return match // 保留原样
	})

	if result == version || strings.Contains(result, "${") {
		return strings.TrimSpace(result)
	}
	return strings.TrimSpace(result)
}

// extractBOMVersions 从 dependencyManagement 提取 BOM 版本映射
func extractBOMVersions(proj *Project) map[string]string {
	versions := make(map[string]string)
	for _, dep := range proj.DependencyManagement.Dependencies {
		if dep.Version != "" && dep.Type != "pom" {
			key := dep.GroupID + ":" + dep.ArtifactID
			versions[key] = dep.Version
		}
	}
	return versions
}

// resolveFromBOMFile 从已安装的 Spring Boot BOM POM 中查找依赖版本
func resolveFromBOMFile(mavenRepo string, props map[string]string, groupID, artifactID string) string {
	// 尝试从 spring-boot-dependencies BOM 解析
	bootVersion := props["spring-boot.version"]
	if bootVersion == "" {
		// 也尝试其他常见属性名
		bootVersion = props["spring-boot-dependencies.version"]
	}
	if bootVersion == "" {
		return ""
	}

	// 检查多个可能的 BOM 路径
	candidates := []struct {
		group string
		artifact string
	}{
		{"org.springframework.boot", "spring-boot-dependencies"},
		{"org.springframework.boot", "spring-boot-parent"},
	}

	for _, c := range candidates {
		bomPath := filepath.Join(mavenRepo,
			strings.ReplaceAll(c.group, ".", string(filepath.Separator)),
			c.artifact, bootVersion,
			fmt.Sprintf("%s-%s.pom", c.artifact, bootVersion))

		if _, err := os.Stat(bomPath); err != nil {
			continue
		}

		data, err := os.ReadFile(bomPath)
		if err != nil {
			continue
		}

		bomProps := extractProperties(data)
		var bomProj Project
		if err := xml.Unmarshal(data, &bomProj); err != nil {
			continue
		}

		for _, dep := range bomProj.DependencyManagement.Dependencies {
			if dep.GroupID == groupID && dep.ArtifactID == artifactID && dep.Version != "" {
				ver := resolveVersion(dep.Version, bomProps)
				if ver != "" && !strings.Contains(ver, "${") {
					return ver
				}
			}
		}
	}

	return ""
}

// fallbackCheckLocal 当版本为空时，尝试在本地仓库中查找该 artifact 的任何版本
func fallbackCheckLocal(repoDir, groupID, artifactID string) (string, bool, string) {
	groupPath := strings.ReplaceAll(groupID, ".", string(filepath.Separator))
	artifactDir := filepath.Join(repoDir, groupPath, artifactID)

	entries, err := os.ReadDir(artifactDir)
	if err != nil {
		return "", false, ""
	}

	// 取最新的版本目录
	var latestVersion string
	for _, e := range entries {
		if e.IsDir() && e.Name() > latestVersion {
			latestVersion = e.Name()
		}
	}

	if latestVersion == "" {
		return "", false, ""
	}

	jarPath, installed := checkLocalJar(repoDir, groupID, artifactID, latestVersion)
	return latestVersion, installed, jarPath
}

// checkLocalJar 检查本地 Maven 仓库中是否存在指定版本的 jar
func checkLocalJar(repoDir, groupID, artifactID, version string) (string, bool) {
	// groupId: com.example → com/example
	groupPath := strings.ReplaceAll(groupID, ".", string(filepath.Separator))
	versionPath := filepath.Join(repoDir, groupPath, artifactID, version)
	jarName := fmt.Sprintf("%s-%s.jar", artifactID, version)
	jarPath := filepath.Join(versionPath, jarName)

	// 检查 jar 文件
	if _, err := os.Stat(jarPath); err == nil {
		return jarPath, true
	}

	// 也检查 pom 文件（对于未打包的依赖）
	pomPath := filepath.Join(versionPath, fmt.Sprintf("%s-%s.pom", artifactID, version))
	if _, err := os.Stat(pomPath); err == nil {
		return pomPath, true
	}

	return jarPath, false
}

// FormatPath 格式化路径（跨平台）
func FormatPath(p string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(p, "/", "\\")
	}
	return p
}

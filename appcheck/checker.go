// Package appcheck 检查 application.yml/properties 配置
package appcheck

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigFormat 配置文件格式
type ConfigFormat string

const (
	FormatYML        ConfigFormat = "yml"
	FormatProperties ConfigFormat = "properties"
)

// RequiredConfig 一个必需的配置项（或配置块）
type RequiredConfig struct {
	Key         string `json:"key"`         // 例如 spring.datasource.url
	YamlPath    string `json:"yamlPath"`    // YAML 层级路径, 如 spring.datasource.url
	PropKey     string `json:"propKey"`     // properties key, 如 spring.datasource.url
	YamlValue   string `json:"yamlValue"`   // YAML 格式的值/块
	PropValue   string `json:"propValue"`   // properties 格式的值
	Reason      string `json:"reason"`      // 为什么需要
	IsBlock     bool   `json:"isBlock"`     // true = 多行 YAML 块, false = 单行键值
}

// ConfigCheckResult 检查结果
type ConfigCheckResult struct {
	Config   RequiredConfig `json:"config"`
	Found    bool           `json:"found"`
	FoundVal string         `json:"foundVal"` // 找到的值（摘要）
}

// CheckReport 检查报告
type CheckReport struct {
	FilePath   string             `json:"filePath"`
	Format     string             `json:"format"` // yml / properties
	Found      []ConfigCheckResult `json:"found"`
	Missing    []ConfigCheckResult `json:"missing"`
	Total      int                `json:"total"`
	FoundCount int                `json:"foundCount"`
}

// requiredConfigs 根据生成配置返回必需的配置项列表
func requiredConfigs(dbType, orm, swagger, springVersion string) []RequiredConfig {
	var configs []RequiredConfig

	// 通用
	configs = append(configs, RequiredConfig{
		Key:       "server.port",
		YamlPath:  "server.port",
		PropKey:   "server.port",
		YamlValue: "8080",
		PropValue: "8080",
		Reason:    "服务器端口",
	})

	// 数据库连接
	switch dbType {
	case "mysql":
		configs = append(configs, RequiredConfig{
			Key:       "spring.datasource.url",
			YamlPath:  "spring.datasource.url",
			PropKey:   "spring.datasource.url",
			YamlValue: "jdbc:mysql://localhost:3306/dbname?useUnicode=true&characterEncoding=utf-8&serverTimezone=Asia/Shanghai",
			PropValue: "jdbc:mysql://localhost:3306/dbname?useUnicode=true&characterEncoding=utf-8&serverTimezone=Asia/Shanghai",
			Reason:    "MySQL 数据库连接 URL",
		})
		configs = append(configs, RequiredConfig{
			Key:       "spring.datasource.username",
			YamlPath:  "spring.datasource.username",
			PropKey:   "spring.datasource.username",
			YamlValue: "root",
			PropValue: "root",
			Reason:    "数据库用户名",
		})
		configs = append(configs, RequiredConfig{
			Key:       "spring.datasource.password",
			YamlPath:  "spring.datasource.password",
			PropKey:   "spring.datasource.password",
			YamlValue: "your_password",
			PropValue: "your_password",
			Reason:    "数据库密码",
		})
		configs = append(configs, RequiredConfig{
			Key:       "spring.datasource.driver-class-name",
			YamlPath:  "spring.datasource.driver-class-name",
			PropKey:   "spring.datasource.driver-class-name",
			YamlValue: "com.mysql.cj.jdbc.Driver",
			PropValue: "com.mysql.cj.jdbc.Driver",
			Reason:    "MySQL JDBC 驱动",
		})
	case "postgresql":
		configs = append(configs, RequiredConfig{
			Key:       "spring.datasource.url",
			YamlPath:  "spring.datasource.url",
			PropKey:   "spring.datasource.url",
			YamlValue: "jdbc:postgresql://localhost:5432/dbname",
			PropValue: "jdbc:postgresql://localhost:5432/dbname",
			Reason:    "PostgreSQL 数据库连接 URL",
		})
		configs = append(configs, RequiredConfig{
			Key:       "spring.datasource.username",
			YamlPath:  "spring.datasource.username",
			PropKey:   "spring.datasource.username",
			YamlValue: "postgres",
			PropValue: "postgres",
			Reason:    "数据库用户名",
		})
		configs = append(configs, RequiredConfig{
			Key:       "spring.datasource.password",
			YamlPath:  "spring.datasource.password",
			PropKey:   "spring.datasource.password",
			YamlValue: "your_password",
			PropValue: "your_password",
			Reason:    "数据库密码",
		})
		configs = append(configs, RequiredConfig{
			Key:       "spring.datasource.driver-class-name",
			YamlPath:  "spring.datasource.driver-class-name",
			PropKey:   "spring.datasource.driver-class-name",
			YamlValue: "org.postgresql.Driver",
			PropValue: "org.postgresql.Driver",
			Reason:    "PostgreSQL JDBC 驱动",
		})
	}

	// ORM 配置
	switch orm {
	case "mybatis-plus":
		configs = append(configs, RequiredConfig{
			Key:      "mybatis-plus.mapper-locations",
			YamlPath: "mybatis-plus.mapper-locations",
			PropKey:  "mybatis-plus.mapper-locations",
			YamlValue: `classpath*:mapper/**/*.xml`,
			PropValue: "classpath*:mapper/**/*.xml",
			Reason:   "MyBatis-Plus XML 映射路径",
		})
		configs = append(configs, RequiredConfig{
			Key:      "mybatis-plus.type-aliases-package",
			YamlPath: "mybatis-plus.type-aliases-package",
			PropKey:  "mybatis-plus.type-aliases-package",
			YamlValue: "${package}.entity",
			PropValue: "${package}.entity",
			Reason:   "MyBatis-Plus 实体别名包",
		})
	case "mybatis":
		configs = append(configs, RequiredConfig{
			Key:      "mybatis.mapper-locations",
			YamlPath: "mybatis.mapper-locations",
			PropKey:  "mybatis.mapper-locations",
			YamlValue: `classpath*:mapper/**/*.xml`,
			PropValue: "classpath*:mapper/**/*.xml",
			Reason:   "MyBatis XML 映射路径",
		})
		configs = append(configs, RequiredConfig{
			Key:      "mybatis.type-aliases-package",
			YamlPath: "mybatis.type-aliases-package",
			PropKey:  "mybatis.type-aliases-package",
			YamlValue: "${package}.entity",
			PropValue: "${package}.entity",
			Reason:   "MyBatis 实体别名包",
		})
	case "jpa":
		configs = append(configs, RequiredConfig{
			Key:      "spring.jpa.hibernate.ddl-auto",
			YamlPath: "spring.jpa.hibernate.ddl-auto",
			PropKey:  "spring.jpa.hibernate.ddl-auto",
			YamlValue: "update",
			PropValue: "update",
			Reason:   "JPA DDL 自动更新",
		})
		configs = append(configs, RequiredConfig{
			Key:      "spring.jpa.show-sql",
			YamlPath: "spring.jpa.show-sql",
			PropKey:  "spring.jpa.show-sql",
			YamlValue: "true",
			PropValue: "true",
			Reason:   "JPA SQL 日志",
		})
	}

	return configs
}

// CheckFile 检查配置文件
func CheckFile(filePath, dbType, orm, swagger, springVersion string) (*CheckReport, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	content := string(data)
	format := detectFormat(filePath)
	required := requiredConfigs(dbType, orm, swagger, springVersion)

	report := &CheckReport{
		FilePath: filePath,
		Format:   string(format),
	}

	for _, cfg := range required {
		result := ConfigCheckResult{Config: cfg}
		val, found := checkConfig(content, format, cfg)
		result.Found = found
		result.FoundVal = val
		if found {
			report.Found = append(report.Found, result)
			report.FoundCount++
		} else {
			report.Missing = append(report.Missing, result)
		}
		report.Total++
	}

	return report, nil
}

// detectFormat 检测配置文件格式
func detectFormat(path string) ConfigFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yml", ".yaml":
		return FormatYML
	case ".properties":
		return FormatProperties
	default:
		return FormatYML
	}
}

// checkConfig 检查配置是否已存在
func checkConfig(content string, format ConfigFormat, cfg RequiredConfig) (string, bool) {
	if format == FormatProperties {
		return checkPropertiesConfig(content, cfg)
	}
	return checkYamlConfig(content, cfg)
}

// checkPropertiesConfig 检查 properties 格式
func checkPropertiesConfig(content string, cfg RequiredConfig) (string, bool) {
	// 按行扫描
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(line, cfg.PropKey+"=") || strings.HasPrefix(line, cfg.PropKey+":") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), true
			}
			parts = strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), true
			}
			return "", true
		}
	}
	return "", false
}

// checkYamlConfig 检查 YAML 格式
func checkYamlConfig(content string, cfg RequiredConfig) (string, bool) {
	// 用 YAML 路径的各级 key 做层级匹配
	// 例如 spring.datasource.url → 检查是否有 spring: → datasource: → url:
	keys := strings.Split(cfg.YamlPath, ".")

	// 简化检查：直接看最后一级 key 是否在文件中的合适层级出现
	// 取最后一级 key 名
	lastKey := keys[len(keys)-1]

	lines := strings.Split(content, "\n")
	depth := 0
	targetDepth := len(keys) - 1 // YAML 中根层级为 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// 计算缩进层级（2空格=1级）
		indent := countIndent(line)
		if indent%2 != 0 {
			indent = indent / 2
		} else {
			indent = indent / 2
		}

		// 更新当前深度
		if indent <= depth && !strings.Contains(trimmed, ":") {
			continue
		}

		// 检查是否匹配最后一级 key
		if indent == targetDepth {
			if strings.HasPrefix(trimmed, lastKey+":") {
				val := strings.TrimSpace(strings.TrimPrefix(trimmed, lastKey+":"))
				return val, true
			}
		}
	}

	return "", false
}

func countIndent(line string) int {
	count := 0
	for _, c := range line {
		if c == ' ' {
			count++
		} else if c == '\t' {
			count += 2
		} else {
			break
		}
	}
	return count
}

// AutoFixConfig 生成缺少的配置内容
type AutoFixConfig struct {
	FilePath    string `json:"filePath"`
	Format      string `json:"format"`
	Snippet     string `json:"snippet"`     // 要插入的内容
	Written     bool   `json:"written"`
	Preview     string `json:"preview"`
}

// AutoFixPreview 预览补全
func AutoFixPreview(filePath, dbType, orm, swagger, springVersion string) (*AutoFixConfig, error) {
	return autoFix(filePath, dbType, orm, swagger, springVersion, false)
}

// AutoFixWrite 写入补全
func AutoFixWrite(filePath, dbType, orm, swagger, springVersion string) (*AutoFixConfig, error) {
	return autoFix(filePath, dbType, orm, swagger, springVersion, true)
}

func autoFix(filePath, dbType, orm, swagger, springVersion string, write bool) (*AutoFixConfig, error) {
	report, err := CheckFile(filePath, dbType, orm, swagger, springVersion)
	if err != nil {
		return nil, err
	}

	if len(report.Missing) == 0 {
		return &AutoFixConfig{
			FilePath: filePath,
			Format:   report.Format,
		}, nil
	}

	format := FormatYML
	if report.Format == "properties" {
		format = FormatProperties
	}

	// 读取原内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	content := string(data)

	// 直接生成完整 YAML 块追加到末尾（Spring 的 SnakeYAML 正确处理重复 key）
	snippet := buildFullYamlSnippet(report.Missing, format)

	newContent := content
	if !strings.HasSuffix(content, "\n") {
		newContent += "\n"
	}
	newContent += snippet

	result := &AutoFixConfig{
		FilePath: filePath,
		Format:   report.Format,
		Snippet:  snippet,
		Preview:  snippet,
	}

	if write {
		os.WriteFile(filePath+".bak", data, 0644)
		if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
			return nil, fmt.Errorf("写入文件失败: %w", err)
		}
		result.Written = true
	}

	return result, nil
}

// mergeYamlContent 将缺失配置合并到 YAML 内容中
// 如果顶层 key 已存在，将子项插入到该块末尾
// 如果顶层 key 不存在，追加到文件末尾
func mergeYamlContent(content string, missing []ConfigCheckResult, format ConfigFormat) (string, string) {
	if len(missing) == 0 {
		return content, ""
	}

	var sb strings.Builder
	sb.WriteString("\n# ===== Spring 代码生成器自动补全 =====\n")

	// properties 格式简单追加
	if format == FormatProperties {
		for _, m := range missing {
			sb.WriteString(fmt.Sprintf("%s=%s\n", m.Config.PropKey, m.Config.PropValue))
		}
		return content + sb.String(), sb.String()
	}

	// YAML: 构建树
	tree := buildYamlTree(missing)
	rootKeys := findExistingRootKeys(content)
	var appendAtEnd strings.Builder

	modified := content

	for _, rootChild := range tree.children {
		if contains(rootKeys, rootChild.name) {
			// 根 key 已存在 → 插入子项到块末尾
			insertion := renderChildrenYAML(rootChild, 1) // 缩进一级
			modified = insertAfterBlock(modified, rootChild.name, insertion)
			sb.WriteString(insertion)
		} else {
			// 根 key 不存在 → 追加到末尾
			block := renderFullYAMLNode(rootChild, 0)
			appendAtEnd.WriteString(block)
			sb.WriteString(block)
		}
	}

	// 追加不存在根 key 的块
	extra := appendAtEnd.String()
	if extra != "" {
		modified += extra
	}

	return modified, sb.String()
}

// renderChildrenYAML 渲染一个节点的所有子项（不含节点本身）
func renderChildrenYAML(node *yamlNode, indent int) string {
	var sb strings.Builder
	for _, child := range node.children {
		sb.WriteString(renderFullYAMLNode(child, indent))
	}
	return sb.String()
}

// renderFullYAMLNode 完整渲染一个 YAML 节点（含节点本身）
func renderFullYAMLNode(node *yamlNode, indent int) string {
	if node == nil {
		return ""
	}
	prefix := strings.Repeat("  ", indent)
	var sb strings.Builder
	if node.isLeaf {
		sb.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, node.name, node.value))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, node.name))
		for _, child := range node.children {
			sb.WriteString(renderFullYAMLNode(child, indent+1))
		}
	}
	return sb.String()
}

// insertAfterBlock 在 YAML 内容中，找到指定根 key 的块结束位置，插入内容
func insertAfterBlock(content, rootKey, insertion string) string {
	lines := strings.Split(content, "\n")
	rootPrefix := rootKey + ":"
	inBlock := false
	blockEnd := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inBlock {
			// 找根 key 所在行
			if countIndent(line) == 0 && strings.HasPrefix(trimmed, rootPrefix) {
				inBlock = true
			}
		} else {
			// 在块内，检查是否遇到了新的顶层 key
			if countIndent(line) == 0 && strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "#") {
				blockEnd = i
				break
			}
		}
	}

	if inBlock && blockEnd == -1 {
		// 根 key 一直持续到文件末尾
		blockEnd = len(lines)
	}

	if blockEnd == -1 {
		// 没找到根 key，追加到末尾
		return content + "\n" + insertion
	}

	// 在 blockEnd 位置插入
	var result []string
	result = append(result, lines[:blockEnd]...)
	result = append(result, insertion)
	result = append(result, lines[blockEnd:]...)
	return strings.Join(result, "\n")
}

// buildFullYamlSnippet 生成完整 YAML 配置块（含根 key），追加到文件末尾
func buildFullYamlSnippet(missing []ConfigCheckResult, format ConfigFormat) string {
	if len(missing) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n# ===== Spring 代码生成器自动补全 =====\n")
	if format == FormatProperties {
		for _, m := range missing {
			sb.WriteString(fmt.Sprintf("%s=%s\n", m.Config.PropKey, m.Config.PropValue))
		}
	} else {
		sb.WriteString(buildYamlBlock(missing, 0))
	}
	return sb.String()
}

// buildYamlBlock 构建 YAML 配置块
func buildYamlBlock(configs []ConfigCheckResult, baseIndent int) string {
	if len(configs) == 0 {
		return ""
	}
	tree := buildYamlTree(configs)
	return renderYamlTree(tree, baseIndent)
}

// yamlNode YAML 树节点
type yamlNode struct {
	name     string
	value    string
	children []*yamlNode
	isLeaf   bool
}

// buildYamlTree 构建 YAML 树
func buildYamlTree(configs []ConfigCheckResult) *yamlNode {
	root := &yamlNode{name: "root"}

	for _, c := range configs {
		keys := strings.Split(c.Config.YamlPath, ".")
		insertYamlPath(root, keys, c.Config)
	}

	return root
}

func insertYamlPath(node *yamlNode, keys []string, cfg RequiredConfig) {
	if len(keys) == 0 {
		return
	}
	key := keys[0]

	// 找子节点
	for _, child := range node.children {
		if child.name == key {
			if len(keys) == 1 {
				child.value = cfg.YamlValue
				child.isLeaf = true
			} else {
				insertYamlPath(child, keys[1:], cfg)
			}
			return
		}
	}

	// 创建新节点
	newNode := &yamlNode{name: key}
	if len(keys) == 1 {
		newNode.value = cfg.YamlValue
		newNode.isLeaf = true
	} else {
		insertYamlPath(newNode, keys[1:], cfg)
	}
	node.children = append(node.children, newNode)
}

// renderYamlTree 渲染 YAML 树为文本
func renderYamlTree(node *yamlNode, indent int) string {
	return doRenderYamlTree(node, indent, false)
}

// renderYamlTreeSkipExisting 渲染但跳过文件中已有的根 key
func renderYamlTreeSkipExisting(node *yamlNode, indent int, existingContent string) string {
	skipKeys := findExistingRootKeys(existingContent)
	return doRenderYamlTree(node, indent, false, skipKeys...)
}

// doRenderYamlTree 渲染 YAML 树
func doRenderYamlTree(node *yamlNode, indent int, isSub bool, skipRootKeys ...string) string {
	if node.name == "root" {
		var sb strings.Builder
		for _, child := range node.children {
			// 跳过已存在的根 key
			if !isSub && contains(skipRootKeys, child.name) {
				// 子节点缩进一级，插入到已有根 key 下
				for _, gc := range child.children {
					sb.WriteString(doRenderYamlTree(gc, indent+1, true))
				}
				continue
			}
			sb.WriteString(doRenderYamlTree(child, indent, false, skipRootKeys...))
		}
		return sb.String()
	}

	prefix := strings.Repeat("  ", indent)
	var sb strings.Builder

	if node.isLeaf {
		sb.WriteString(fmt.Sprintf("%s%s: %s\n", prefix, node.name, node.value))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, node.name))
		for _, child := range node.children {
			sb.WriteString(doRenderYamlTree(child, indent+1, false, skipRootKeys...))
		}
	}

	return sb.String()
}

// findExistingRootKeys 从 YAML 内容中找到所有顶层 key
func findExistingRootKeys(content string) []string {
	var keys []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// 顶层 key: 没有前导空格，以 : 结尾或包含 : 值
		indent := countIndent(line)
		if indent == 0 && strings.Contains(trimmed, ":") {
			key := strings.SplitN(trimmed, ":", 2)[0]
			key = strings.TrimSpace(key)
			if key != "" && !contains(keys, key) {
				keys = append(keys, key)
			}
		}
	}
	return keys
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}



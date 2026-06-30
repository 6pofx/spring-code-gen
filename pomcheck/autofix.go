package pomcheck

import (
	"fmt"
	"os"
	"strings"
)

// AutoFixResult 自动补全结果
type AutoFixResult struct {
	Preview     string `json:"preview"`
	InsertCount int    `json:"insertCount"`
	Written     bool   `json:"written"`
	FilePath    string `json:"filePath"`
}

// AutoFixPomPreview 预览补全结果（不写入文件）
func AutoFixPomPreview(pomPath string, missingDeps []RequiredDep) (*AutoFixResult, error) {
	return autoFixPom(pomPath, missingDeps, false)
}

// AutoFixPomWrite 补全并写入 pom.xml
func AutoFixPomWrite(pomPath string, missingDeps []RequiredDep) (*AutoFixResult, error) {
	return autoFixPom(pomPath, missingDeps, true)
}

func autoFixPom(pomPath string, missingDeps []RequiredDep, write bool) (*AutoFixResult, error) {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("读取 pom.xml 失败: %w", err)
	}

	content := string(data)
	snippet, count := dedupAndBuildSnippet(content, missingDeps)
	if count == 0 {
		return &AutoFixResult{InsertCount: 0, Preview: content, FilePath: pomPath}, nil
	}

	idx := findCloseTag(content)
	if idx == -1 {
		return nil, fmt.Errorf("未找到主 <dependencies> 的闭合标签")
	}

	newContent := content[:idx] + snippet + "\t" + content[idx:]

	result := &AutoFixResult{
		InsertCount: count,
		FilePath:    pomPath,
		Preview:     buildPreview(newContent, idx, len(snippet)),
	}

	if write {
		os.WriteFile(pomPath+".bak", data, 0644)
		if err := os.WriteFile(pomPath, []byte(newContent), 0644); err != nil {
			return nil, fmt.Errorf("写入失败: %w", err)
		}
		result.Written = true
	}

	return result, nil
}

// dedupAndBuildSnippet 去重 + 生成插入片段
func dedupAndBuildSnippet(content string, deps []RequiredDep) (string, int) {
	var b strings.Builder
	n := 0
	for _, d := range deps {
		if strings.Contains(content, "<artifactId>"+d.ArtifactID+"</artifactId>") {
			continue
		}
		b.WriteString("\n")
		b.WriteString(d.toXML("\t\t"))
		n++
	}
	return b.String(), n
}

// findCloseTag 找到主 <dependencies> 的 </dependencies>
// 规则：正向找到第一个 </dependencies>，它前面不能有 <![CDATA[ 且不能是 <!-- ... -->
// 实际直接用极简规则：找第一个不在注释内的 </dependencies>
func findCloseTag(content string) int {
	closeTag := "</dependencies>"
	i := 0
	for i < len(content) {
		// 跳过注释
		if strings.HasPrefix(content[i:], "<!--") {
			end := strings.Index(content[i:], "-->")
			if end == -1 {
				break
			}
			i += end + 3
			continue
		}
		// 检查是否匹配
		if strings.HasPrefix(content[i:], closeTag) {
			return i
		}
		i++
	}
	return -1
}

// InsertDepsIntoContent 对 pom.xml 内容字符串插入缺失依赖
func InsertDepsIntoContent(content string, missingDeps []RequiredDep) (string, int, error) {
	snippet, count := dedupAndBuildSnippet(content, missingDeps)
	if count == 0 {
		return content, 0, nil
	}
	idx := findCloseTag(content)
	if idx == -1 {
		return "", 0, fmt.Errorf("未找到主 <dependencies> 标签")
	}
	return content[:idx] + snippet + "\t" + content[idx:], count, nil
}

// buildPreview 构建修改预览
func buildPreview(newContent string, insertPos, snippetLen int) string {
	before := 200
	after := 200
	start := insertPos - before
	if start < 0 {
		start = 0
	}
	end := insertPos + snippetLen + after
	if end > len(newContent) {
		end = len(newContent)
	}
	preview := newContent[start:end]
	rel := insertPos - start
	if rel >= 0 && rel <= len(preview) {
		preview = preview[:rel] + "\n<<<<<< 新增依赖 >>>>>>\n" + preview[rel:]
	}
	return preview
}

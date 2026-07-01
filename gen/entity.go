package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// EntityTemplateMybatisPlus MyBatis-Plus 实体模板
const EntityTemplateMybatisPlus = `package {{.EntityPackageFull}};

import java.io.Serializable;
import com.baomidou.mybatisplus.annotation.*;
{{range .Imports}}import {{.}};
{{end}}
{{if .Swagger}}
{{SwaggerModelImport .Swagger}}
{{end}}
{{if .UseLombok}}import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.experimental.Accessors;
{{end}}
{{if .UseLombok}}@Data
@EqualsAndHashCode(callSuper = false)
@Accessors(chain = true){{end}}
{{if .Swagger}}{{SwaggerModelAnnotation .Swagger .TableComment}}
{{end}}@TableName("{{.TableName}}")
public class {{.EntityName}} implements Serializable {

    private static final long serialVersionUID = 1L;
{{range .Columns}}
{{ColumnComment .ColComment}}
{{ColumnAnnotation . .IsPrimaryKey .IsAutoInc $.Swagger $.Namespace}}
    private {{.JavaType}} {{.JavaField}};
{{end}}
{{if not .UseLombok}}
{{range .Columns}}
{{GetterSetter .JavaType .JavaField $.EntityName}}
{{end}}{{end}}
}
`

// EntityTemplateMyBatis MyBatis 实体模板（无注解）
const EntityTemplateMyBatis = `package {{.EntityPackageFull}};

{{if .Imports}}import java.io.Serializable;
{{range .Imports}}import {{.}};
{{end}}{{else}}import java.io.Serializable;
{{end}}
{{if .Swagger}}
{{SwaggerModelImport .Swagger}}
{{end}}
{{if .UseLombok}}import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.experimental.Accessors;
{{end}}
{{if .UseLombok}}@Data
@EqualsAndHashCode(callSuper = false)
@Accessors(chain = true){{end}}
{{if .Swagger}}{{SwaggerModelAnnotation .Swagger .TableComment}}
{{end}}public class {{.EntityName}} implements Serializable {

    private static final long serialVersionUID = 1L;
{{range .Columns}}
{{ColumnComment .ColComment}}
    private {{.JavaType}} {{.JavaField}};
{{end}}
{{if not .UseLombok}}
{{range .Columns}}
{{GetterSetter .JavaType .JavaField $.EntityName}}
{{end}}{{end}}
}
`

// EntityTemplateJPA JPA 实体模板
const EntityTemplateJPA = `package {{.EntityPackageFull}};

{{if .Imports}}import java.io.Serializable;
{{range .Imports}}import {{.}};
{{end}}{{else}}import java.io.Serializable;
{{end}}
{{if .Swagger}}
{{SwaggerModelImport .Swagger}}
{{end}}
{{if .UseLombok}}import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.experimental.Accessors;
{{end}}
{{if eq .SpringVersion "3.x"}}import jakarta.persistence.*;
{{end}}{{if eq .SpringVersion "2.x"}}import javax.persistence.*;
{{end}}
{{if .UseLombok}}@Data
@EqualsAndHashCode(callSuper = false)
@Accessors(chain = true){{end}}
{{if .Swagger}}{{SwaggerModelAnnotation .Swagger .TableComment}}
{{end}}@Entity
{{if .HasCompositePK}}@IdClass({{.EntityName}}PK.class)
{{end}}@Table(name = "{{.TableName}}")
public class {{.EntityName}} implements Serializable {

    private static final long serialVersionUID = 1L;
{{range .Columns}}
{{ColumnComment .ColComment}}
{{ColumnAnnotationJPA . .IsPrimaryKey .IsAutoInc $.Swagger $.Namespace}}
    private {{.JavaType}} {{.JavaField}};
{{end}}
{{if not .UseLombok}}
{{range .Columns}}
{{GetterSetter .JavaType .JavaField $.EntityName}}
{{end}}{{end}}
}
`

// FuncMap 模板函数映射
var FuncMap = template.FuncMap{
	"ColumnComment":             columnComment,
	"ColumnAnnotation":          columnAnnotation,
	"ColumnAnnotationJPA":       columnAnnotationJPA,
	"GetterSetter":              getterSetter,
	"SwaggerModelImport":        swaggerModelImport,
	"SwaggerModelAnnotation":    swaggerModelAnnotation,
	"SwaggerFieldAnnotation":    swaggerFieldAnnotation,
	"SwaggerFieldAnnotationJPA": swaggerFieldAnnotationJPA,
	"UpperFirst":                UpperFirst,
	"LowerFirst":                LowerFirst,
}

func getEntityTemplate(orm string) string {
	switch orm {
	case "mybatis-plus":
		return EntityTemplateMybatisPlus
	case "mybatis":
		return EntityTemplateMyBatis
	case "jpa":
		return EntityTemplateJPA
	default:
		return EntityTemplateMybatisPlus
	}
}

func columnComment(comment string) string {
	if comment == "" {
		return ""
	}
	return fmt.Sprintf("    /** %s */", comment)
}

func columnAnnotation(col ColumnInfo, isPK bool, isAutoInc bool, swagger, namespace string) string {
	var annos []string

	// MyBatis-Plus 主键注解：仅第一个主键加 @TableId，后续 PK 不加（复合主键）
	if isPK && col.PKOrder == 1 {
		strategy := GetIdType(isAutoInc)
		annos = append(annos, fmt.Sprintf("    @TableId(value = \"%s\", type = %s)", col.ColName, strategy))
	} else if isPK && col.PKOrder > 1 {
		// 复合主键的后续列，用 @TableField 但不加注解（保留空注释）
	} else if !IsCommonField(col.ColName) {
		// 非主键非通用字段
		annos = append(annos, fmt.Sprintf("    @TableField(\"%s\")", col.ColName))
	}

	// Swagger 字段注解
	if col.ColComment != "" {
		if swagger == "springdoc" {
			annos = append(annos, fmt.Sprintf("    @Schema(description = \"%s\")", col.ColComment))
		} else {
			annos = append(annos, fmt.Sprintf("    @ApiModelProperty(value = \"%s\")", col.ColComment))
		}
	}

	return strings.Join(annos, "\n")
}

func columnAnnotationJPA(col ColumnInfo, isPK bool, isAutoInc bool, swagger, namespace string) string {
	var annos []string

	// isPK 或 PKOrder>0 都视为主键（双重保障）
	if isPK || col.PKOrder > 0 {
		annos = append(annos, "    @Id")
		if isAutoInc {
			annos = append(annos, "    @GeneratedValue(strategy = GenerationType.IDENTITY)")
		}
		annos = append(annos, fmt.Sprintf("    @Column(name = \"%s\")", col.ColName))
	} else {
		annos = append(annos, fmt.Sprintf("    @Column(name = \"%s\"%s)", col.ColName, ifNullable(col.IsNullable)))
	}

	// Swagger 字段注解
	if col.ColComment != "" {
		if swagger == "springdoc" {
			annos = append(annos, fmt.Sprintf("    @Schema(description = \"%s\")", col.ColComment))
		} else {
			annos = append(annos, fmt.Sprintf("    @ApiModelProperty(value = \"%s\")", col.ColComment))
		}
	}

	return strings.Join(annos, "\n")
}

func ifNullable(nullable bool) string {
	if nullable {
		return ""
	}
	return ", nullable = false"
}

func getterSetter(javaType, fieldName, entityName string) string {
	var sb strings.Builder
	// Getter
	sb.WriteString(fmt.Sprintf("\n    public %s get%s() {\n        return this.%s;\n    }\n",
		javaType, UpperFirst(fieldName), fieldName))
	// Setter
	sb.WriteString(fmt.Sprintf("\n    public void set%s(%s %s) {\n        this.%s = %s;\n    }",
		UpperFirst(fieldName), javaType, fieldName, fieldName, fieldName))
	return sb.String()
}

func swaggerModelImport(swagger string) string {
	if swagger == "springdoc" {
		return "import io.swagger.v3.oas.annotations.media.Schema;"
	}
	return "import io.swagger.annotations.ApiModel;\nimport io.swagger.annotations.ApiModelProperty;"
}

func swaggerModelAnnotation(swagger, tableComment string) string {
	if swagger == "springdoc" {
		if tableComment != "" {
			return fmt.Sprintf("@Schema(description = \"%s\")", tableComment)
		}
		return ""
	}
	if tableComment != "" {
		return fmt.Sprintf("@ApiModel(value = \"%s\", description = \"%s\")", tableComment, tableComment)
	}
	return ""
}

func swaggerFieldAnnotation(swagger, comment string) string {
	if comment == "" {
		return ""
	}
	if swagger == "springdoc" {
		return fmt.Sprintf("@Schema(description = \"%s\")", comment)
	}
	return fmt.Sprintf("@ApiModelProperty(value = \"%s\")", comment)
}

func swaggerFieldAnnotationJPA(swagger, comment string) string {
	if comment == "" {
		return ""
	}
	if swagger == "springdoc" {
		return fmt.Sprintf("    @Schema(description = \"%s\")", comment)
	}
	return fmt.Sprintf("    @ApiModelProperty(value = \"%s\")", comment)
}

// PKClassTemplate JPA 复合主键类模板
const PKClassTemplate = `package {{.EntityPackageFull}};

import java.io.Serializable;
{{if .UseLombok}}import lombok.Data;
import lombok.EqualsAndHashCode;
{{end}}
/**
 * {{.TableComment}} 复合主键类
 */
{{if .UseLombok}}@Data
@EqualsAndHashCode{{end}}
public class {{.EntityName}}PK implements Serializable {

    private static final long serialVersionUID = 1L;
{{range .Columns}}{{if .IsPrimaryKey}}
    private {{.JavaType}} {{.JavaField}};{{end}}{{end}}
{{if not .UseLombok}}
{{range .Columns}}{{if .IsPrimaryKey}}
{{GetterSetter .JavaType .JavaField $.EntityName}}{{end}}{{end}}{{end}}
}
`

// GeneratePKClass 生成复合主键类
func GeneratePKClass(td *TemplateData) (string, error) {
	tmpl, err := template.New("pkClass").Funcs(FuncMap).Parse(PKClassTemplate)
	if err != nil {
		return "", fmt.Errorf("解析 PK 类模板失败: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", fmt.Errorf("执行 PK 类模板失败: %w", err)
	}
	return formatJavaCode(buf.String()), nil
}

// WritePKClassFile 写入复合主键类文件
func WritePKClassFile(td *TemplateData, content string, overwritePolicy string) (string, error) {
	dir := td.GetEntityDir()
	fileName := fmt.Sprintf("%sPK.java", td.EntityName)
	filePath := filepath.Join(dir, fileName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}
	return writeFile(filePath, content, overwritePolicy)
}

// GenerateEntity 生成 Entity 文件
func GenerateEntity(td *TemplateData) (string, error) {
	tmplText := getEntityTemplate(td.Orm)
	tmpl, err := template.New("entity").Funcs(FuncMap).Parse(tmplText)
	if err != nil {
		return "", fmt.Errorf("解析 Entity 模板失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, td); err != nil {
		return "", fmt.Errorf("执行 Entity 模板失败: %w", err)
	}

	return formatJavaCode(buf.String()), nil
}

// WriteEntityFile 写入 Entity 文件
func WriteEntityFile(td *TemplateData, content string, overwritePolicy string) (string, error) {
	dir := td.GetEntityDir()
	fileName := fmt.Sprintf("%s.java", td.EntityName)
	filePath := filepath.Join(dir, fileName)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}

	return writeFile(filePath, content, overwritePolicy)
}

// formatJavaCode 简单格式化 Java 代码（去多余空行）
func formatJavaCode(code string) string {
	// 去除多余的连续空行（最多保留一个空行）
	lines := strings.Split(code, "\n")
	var result []string
	emptyCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyCount++
			if emptyCount <= 1 {
				result = append(result, "")
			}
		} else {
			emptyCount = 0
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// writeFile 根据覆盖策略写入文件
func writeFile(filePath, content, overwritePolicy string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); err == nil {
		switch overwritePolicy {
		case "skip":
			return fmt.Sprintf("跳过已存在文件: %s", filePath), nil
		case "overwrite":
			// 继续写入
		default:
			// ask 模式下也覆盖（前端已确认）
		}
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入文件失败 %s: %w", filePath, err)
	}
	return fmt.Sprintf("✓ 已生成: %s", filePath), nil
}

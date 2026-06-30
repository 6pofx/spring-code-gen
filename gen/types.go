package gen

import (
	"fmt"
	"strings"
	"unicode"
)

// JavaTypeMap MySQL/PostgreSQL 数据库类型 → Java 类型映射
var JavaTypeMap = map[string]string{
	// MySQL
	"int":       "Integer",
	"integer":   "Integer",
	"bigint":    "Long",
	"smallint":  "Integer",
	"tinyint":   "Integer",
	"varchar":   "String",
	"char":      "String",
	"text":      "String",
	"mediumtext": "String",
	"longtext":  "String",
	"boolean":   "Boolean",
	"bool":      "Boolean",
	"date":      "java.time.LocalDate",
	"datetime":  "java.time.LocalDateTime",
	"timestamp": "java.time.LocalDateTime",
	"decimal":   "java.math.BigDecimal",
	"numeric":   "java.math.BigDecimal",
	"float":     "Float",
	"double":    "Double",
	"blob":      "byte[]",
	"longblob":  "byte[]",
	"mediumblob": "byte[]",
	"json":      "String",
	// PostgreSQL
	"int4":          "Integer",
	"int8":          "Long",
	"int2":          "Integer",
	"serial":        "Integer",
	"bigserial":     "Long",
	"smallserial":   "Integer",
	"character varying": "String",
	"character":     "String",
	"timestamptz":   "java.time.LocalDateTime",
	"real":          "Float",
	"float4":        "Float",
	"float8":        "Double",
	"double precision": "Double",
	"bytea":         "byte[]",
	"jsonb":         "String",
}

// MapToJavaType 将数据库类型映射为 Java 类型
func MapToJavaType(dbType string) string {
	lower := strings.ToLower(dbType)
	// 提取类型名（去掉括号内的精度）
	if idx := strings.Index(lower, "("); idx > 0 {
		lower = lower[:idx]
	}
	lower = strings.TrimSpace(lower)
	if jt, ok := JavaTypeMap[lower]; ok {
		return jt
	}
	return "String"
}

// IsJavaTypePrimitive 判断 Java 类型是否为基本类型包装类（无需 import）
func IsJavaTypePrimitive(javaType string) bool {
	switch javaType {
	case "Integer", "Long", "Float", "Double", "Boolean", "String", "byte[]":
		return true
	}
	return false
}

// NeedImport 从 Java 类型中提取需要 import 的完整类名
func NeedImport(javaType string) string {
	// java.time.LocalDate, java.time.LocalDateTime, java.math.BigDecimal
	// 如果是 byte[] 不需要 import
	if javaType == "byte[]" {
		return ""
	}
	if strings.HasPrefix(javaType, "java.") {
		return javaType
	}
	return ""
}

// SnakeToPascal snake_case → PascalCase
func SnakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		parts[i] = UpperFirst(p)
	}
	return strings.Join(parts, "")
}

// SnakeToCamel snake_case → camelCase
func SnakeToCamel(s string) string {
	pascal := SnakeToPascal(s)
	return LowerFirst(pascal)
}

// SnakeToKebab snake_case → kebab-case
func SnakeToKebab(s string) string {
	return strings.ReplaceAll(s, "_", "-")
}

// UpperFirst 首字母大写
func UpperFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}

// LowerFirst 首字母小写
func LowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	return string(unicode.ToLower(r[0])) + string(r[1:])
}

// GetIdType 根据自增情况返回 MyBatis-Plus IdType 策略
func GetIdType(isAutoInc bool) string {
	if isAutoInc {
		return "IdType.AUTO"
	}
	return "IdType.INPUT"
}

// IsCommonField 判断是否为通用字段（通常由 MyBatis-Plus 自动填充，不写入 insert/update）
var commonFields = map[string]bool{
	"create_time":   true,
	"update_time":   true,
	"created_time":  true,
	"updated_time":  true,
	"deleted":       true,
	"is_deleted":    true,
	"create_by":     true,
	"update_by":     true,
	"created_by":    true,
	"updated_by":    true,
	"version":       true,
}

// GetEntityDir 获取实体类输出目录
func (td *TemplateData) GetEntityDir() string {
	return fmt.Sprintf("%s/%s", td.OutputDir, strings.ReplaceAll(td.EntityPackageFull, ".", "/"))
}

// GetMapperDir 获取 Mapper 输出目录
func (td *TemplateData) GetMapperDir() string {
	return fmt.Sprintf("%s/%s", td.OutputDir, strings.ReplaceAll(td.MapperPackageFull, ".", "/"))
}

// GetMapperXMLDir 获取 Mapper XML 输出目录 (resources/mapper)
func (td *TemplateData) GetMapperXMLDir() string {
	return fmt.Sprintf("%s/src/main/resources/mapper", td.OutputDir)
}

// GetServiceDir 获取 Service 输出目录
func (td *TemplateData) GetServiceDir() string {
	return fmt.Sprintf("%s/%s", td.OutputDir, strings.ReplaceAll(td.ServicePackageFull, ".", "/"))
}

// GetServiceImplDir 获取 ServiceImpl 输出目录
func (td *TemplateData) GetServiceImplDir() string {
	return fmt.Sprintf("%s/%s", td.OutputDir, strings.ReplaceAll(td.ServiceImplPkgFull, ".", "/"))
}

// GetControllerDir 获取 Controller 输出目录
func (td *TemplateData) GetControllerDir() string {
	return fmt.Sprintf("%s/%s", td.OutputDir, strings.ReplaceAll(td.ControllerPkgFull, ".", "/"))
}

// BuildImports 根据列和配置构建 import 列表
func (td *TemplateData) BuildImports() []string {
	importSet := make(map[string]bool)

	// 列类型需要的 import
	for _, col := range td.Columns {
		imp := NeedImport(col.JavaType)
		if imp != "" {
			importSet[imp] = true
		}
	}

	// 响应类 import (如果响应类中有泛型 T，不需要额外 import)
	if td.ResponseClass != "" && td.ResponseClass != "void" {
		// ResponseEntity 在 spring-web 中，不需要显式 import
	}

	// 分页 import 由 Controller/Service 模板自行处理，不在此处统一加

	// 列表
	var result []string
	for imp := range importSet {
		result = append(result, imp)
	}
	sortStrings(result)
	return result
}

func sortStrings(s []string) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// NonPrimaryColumns 返回非主键列
func (td *TemplateData) NonPrimaryColumns() []ColumnInfo {
	var result []ColumnInfo
	for _, c := range td.Columns {
		if !c.IsPrimaryKey {
			result = append(result, c)
		}
	}
	return result
}

// NonCommonColumns 返回非通用字段列
func (td *TemplateData) NonCommonColumns() []ColumnInfo {
	var result []ColumnInfo
	for _, c := range td.Columns {
		if !c.IsCommonField {
			result = append(result, c)
		}
	}
	return result
}

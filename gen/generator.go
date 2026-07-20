package gen

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// BuildTemplateData 根据配置和表信息构建模板数据
func BuildTemplateData(cfg *Config, tableName, tableComment string, columns []ColumnInfo) *TemplateData {
	entityName := SnakeToPascal(tableName)
	entityVar := SnakeToCamel(tableName)
	entityPath := SnakeToKebab(tableName)

	// 确定命名空间
	namespace := "jakarta"
	if cfg.SpringVersion == "2.x" {
		namespace = "javax"
	}

	// 确定注入注解
	var injectAnnotation string
	if cfg.InjectType == "Resource" {
		injectAnnotation = "@Resource"
	} else {
		injectAnnotation = "@Autowired"
	}

	// 包名
	basePkg := cfg.BasePackage
	if basePkg == "" {
		basePkg = "com.example"
	}

	entityPkg := cfg.EntityPackage
	if entityPkg == "" {
		entityPkg = "entity"
	}
	mapperPkg := cfg.MapperPackage
	if mapperPkg == "" {
		mapperPkg = "mapper"
	}
	servicePkg := cfg.ServicePackage
	if servicePkg == "" {
		servicePkg = "service"
	}
	serviceImplPkg := cfg.ServiceImplPkg
	if serviceImplPkg == "" {
		serviceImplPkg = "service.impl"
	}
	controllerPkg := cfg.ControllerPkg
	if controllerPkg == "" {
		controllerPkg = "controller"
	}

	// 响应类
	respClass := cfg.ResponseClass
	if respClass == "" {
		respClass = "R"
	}

	// 标记主键次序，判断复合主键
	pkCount := 0
	for i := range columns {
		if columns[i].IsPrimaryKey {
			pkCount++
			columns[i].PKOrder = pkCount
		}
	}
	hasCompositePK := pkCount > 1

	// 找主键（取第一个）
	var pk ColumnInfo
	for _, c := range columns {
		if c.IsPrimaryKey {
			pk = c
			break
		}
	}
	if pk.JavaType == "" && len(columns) > 0 {
		// 无主键检测到时：JPA 必须至少有一个 @Id，将第一列升为主键
		columns[0].IsPrimaryKey = true
		columns[0].PKOrder = 1
		columns[0].MPStrategy = "IdType.INPUT"
		pk = columns[0]
	} else if pk.JavaType == "" {
		pk = ColumnInfo{
			ColName:    "id",
			JavaType:   "Long",
			JavaField:  "id",
			MPStrategy: "IdType.AUTO",
		}
	}

	td := &TemplateData{
		Config: *cfg,

		TableName:    tableName,
		TableComment: tableComment,
		EntityName:   entityName,
		EntityVar:    entityVar,
		EntityPath:   entityPath,
		PrimaryKey:     pk,
		HasCompositePK: hasCompositePK,
		Columns:      columns,

		EntityPackageFull:  basePkg + "." + entityPkg,
		MapperPackageFull:  basePkg + "." + mapperPkg,
		ServicePackageFull: basePkg + "." + servicePkg,
		ServiceImplPkgFull: basePkg + "." + serviceImplPkg,
		ControllerPkgFull:  basePkg + "." + controllerPkg,
		ConfigPackageFull:   basePkg + ".config",

		Namespace:        namespace,
		InjectAnnotation: injectAnnotation,
	}

	if cfg.ResponseClass != "" {
		td.ResponseClass = cfg.ResponseClass
	} else {
		td.ResponseClass = "R"
	}

	td.Imports = td.BuildImports()

	return td
}

// Result 生成结果
type Result struct {
	Type    string // entity / mapper / mapper-xml / service / service-impl / controller
	Table   string
	File    string
	Success bool
	Message string
}

// GenerateAll 为指定表生成所有代码，通过 channel 返回结果
func GenerateAll(td *TemplateData, cfg *Config, tableName string, results chan<- Result) {
	// Entity
	if cfg.GenEntity {
		content, err := GenerateEntity(td)
		if err != nil {
			results <- Result{Type: "entity", Table: tableName, Success: false, Message: err.Error()}
		} else {
			msg, err := WriteEntityFile(td, content, cfg.OverwritePolicy)
			if err != nil {
				results <- Result{Type: "entity", Table: tableName, Success: false, Message: err.Error()}
			} else {
				results <- Result{Type: "entity", Table: tableName, Success: true, Message: formatResultMsg(msg, td.EntityName+".java")}
			}
		}
		// JPA 复合主键 → 生成 PK 类
		if cfg.Orm == "jpa" && td.HasCompositePK {
			pkContent, pkErr := GeneratePKClass(td)
			if pkErr != nil {
				results <- Result{Type: "entity", Table: tableName, Success: false, Message: pkErr.Error()}
			} else {
				pkMsg, pkErr := WritePKClassFile(td, pkContent, cfg.OverwritePolicy)
				if pkErr != nil {
					results <- Result{Type: "entity", Table: tableName, Success: false, Message: pkErr.Error()}
				} else {
					results <- Result{Type: "entity", Table: tableName, Success: true, Message: formatResultMsg(pkMsg, td.EntityName+"PK.java")}
				}
			}
		}
	}

	// Mapper
	if cfg.GenMapper {
		content, err := GenerateMapper(td)
		if err != nil {
			results <- Result{Type: "mapper", Table: tableName, Success: false, Message: err.Error()}
		} else {
			msg, err := WriteMapperFile(td, content, cfg.OverwritePolicy)
			if err != nil {
				results <- Result{Type: "mapper", Table: tableName, Success: false, Message: err.Error()}
			} else {
				suffix := "Mapper.java"
				if cfg.Orm == "jpa" {
					suffix = "Repository.java"
				}
				results <- Result{Type: "mapper", Table: tableName, Success: true, Message: formatResultMsg(msg, td.EntityName+suffix)}
			}
		}

		// MyBatis XML
		if cfg.Orm == "mybatis" {
			xmlContent, err := GenerateMapperXML(td)
			if err != nil {
				results <- Result{Type: "mapper-xml", Table: tableName, Success: false, Message: err.Error()}
			} else {
				msg, err := WriteMapperXMLFile(td, xmlContent, cfg.OverwritePolicy)
				if err != nil {
					results <- Result{Type: "mapper-xml", Table: tableName, Success: false, Message: err.Error()}
				} else {
					results <- Result{Type: "mapper-xml", Table: tableName, Success: true, Message: formatResultMsg(msg, td.EntityName+"Mapper.xml")}
				}
			}
		}
	}

	// Service + ServiceImpl
	if cfg.GenService {
		svcContent, err := GenerateService(td)
		if err != nil {
			results <- Result{Type: "service", Table: tableName, Success: false, Message: err.Error()}
		} else {
			msg, err := WriteServiceFile(td, svcContent, cfg.OverwritePolicy)
			if err != nil {
				results <- Result{Type: "service", Table: tableName, Success: false, Message: err.Error()}
			} else {
				results <- Result{Type: "service", Table: tableName, Success: true, Message: formatResultMsg(msg, td.EntityName+"Service.java")}
			}
		}

		implContent, err := GenerateServiceImpl(td)
		if err != nil {
			results <- Result{Type: "service-impl", Table: tableName, Success: false, Message: err.Error()}
		} else {
			msg, err := WriteServiceImplFile(td, implContent, cfg.OverwritePolicy)
			if err != nil {
				results <- Result{Type: "service-impl", Table: tableName, Success: false, Message: err.Error()}
			} else {
				results <- Result{Type: "service-impl", Table: tableName, Success: true, Message: formatResultMsg(msg, td.EntityName+"ServiceImpl.java")}
			}
		}
	}

	// Redis 配置类（仅第一个表会执行，后续表跳过）
	if cfg.EnableRedis && td.ConfigPackageFull != "" {
		content, err := GenerateRedisConfig(td)
		if err != nil {
			results <- Result{Type: "redis-config", Table: tableName, Success: false, Message: err.Error()}
		} else {
			msg, err := WriteRedisConfigFile(td, content, cfg.OverwritePolicy)
			if err != nil {
				results <- Result{Type: "redis-config", Table: tableName, Success: false, Message: err.Error()}
			} else {
				results <- Result{Type: "redis-config", Table: tableName, Success: true, Message: formatResultMsg(msg, "RedisConfig.java")}
			}
		}
	}

	// Controller
	if cfg.GenController {
		content, err := GenerateController(td)
		if err != nil {
			results <- Result{Type: "controller", Table: tableName, Success: false, Message: err.Error()}
		} else {
			msg, err := WriteControllerFile(td, content, cfg.OverwritePolicy)
			if err != nil {
				results <- Result{Type: "controller", Table: tableName, Success: false, Message: err.Error()}
			} else {
				results <- Result{Type: "controller", Table: tableName, Success: true, Message: formatResultMsg(msg, td.EntityName+"Controller.java")}
			}
		}
	}
}

func formatResultMsg(path, filename string) string {
	return fmt.Sprintf("✓ 已生成: %s", path)
}

// GenerateResponseClassCode 生成统一响应类 Java 代码
func GenerateResponseClassCode(basePackage, className string) (string, error) {
	if className == "" {
		className = "R"
	}
	tmpl, err := template.New("response").Parse(ResponseClassTemplate)
	if err != nil {
		return "", err
	}
	data := struct {
		BasePackage  string
		ResponseClass string
	}{
		BasePackage:  basePackage,
		ResponseClass: className,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return formatJavaCode(buf.String()), nil
}

// WriteResponseClassFile 写入统一响应类文件
func WriteResponseClassFile(outputDir, basePackage, className, overwritePolicy string) (string, error) {
	content, err := GenerateResponseClassCode(basePackage, className)
	if err != nil {
		return "", err
	}

	dir := fmt.Sprintf("%s/%s", outputDir, strings.ReplaceAll(basePackage, ".", "/"))
	filePath := fmt.Sprintf("%s/%s.java", dir, className)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}

	return writeFile(filePath, content, overwritePolicy)
}

// ResponseClassTemplate 统一响应类模板
const ResponseClassTemplate = `package {{.BasePackage}};

import java.io.Serializable;

/**
 * 统一响应结果
 */
public class {{.ResponseClass}}<T> implements Serializable {

    private static final long serialVersionUID = 1L;

    private int code;
    private String message;
    private T data;

    public {{.ResponseClass}}() {}

    public {{.ResponseClass}}(int code, String message, T data) {
        this.code = code;
        this.message = message;
        this.data = data;
    }

    public static <T> {{.ResponseClass}}<T> ok() {
        return new {{.ResponseClass}}<>(200, "success", null);
    }

    public static <T> {{.ResponseClass}}<T> ok(T data) {
        return new {{.ResponseClass}}<>(200, "success", data);
    }

    public static <T> {{.ResponseClass}}<T> error(String message) {
        return new {{.ResponseClass}}<>(500, message, null);
    }

    public static <T> {{.ResponseClass}}<T> error(int code, String message) {
        return new {{.ResponseClass}}<>(code, message, null);
    }

    public int getCode() { return code; }
    public void setCode(int code) { this.code = code; }
    public String getMessage() { return message; }
    public void setMessage(String message) { this.message = message; }
    public T getData() { return data; }
    public void setData(T data) { this.data = data; }
}
`

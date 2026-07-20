package gen

// Config 代码生成配置
type Config struct {
	// 数据库连接
	DBType     string `json:"dbType"`     // mysql / postgresql
	DBHost     string `json:"dbHost"`
	DBPort     int    `json:"dbPort"`
	DBName     string `json:"dbName"`
	DBUser     string `json:"dbUser"`
	DBPassword string `json:"dbPassword"`

	// 表选择
	TablePrefix   string   `json:"tablePrefix"`
	ExcludePrefix string   `json:"excludePrefix"`
	TableNames    []string `json:"tableNames"`

	// ORM / 框架
	Orm           string `json:"orm"`           // mybatis-plus / mybatis / jpa
	SpringVersion string `json:"springVersion"` // 2.x / 3.x
	Swagger       string `json:"swagger"`       // springdoc / swagger2
	InjectType    string `json:"injectType"`    // Resource / Autowired

	// 响应类
	ResponseClass string `json:"responseClass"` // 默认 "R"

	// Lombok
	UseLombok bool `json:"useLombok"`

	// 包名
	BasePackage      string `json:"basePackage"`      // com.example
	EntityPackage    string `json:"entityPackage"`    // entity
	MapperPackage    string `json:"mapperPackage"`    // mapper
	ServicePackage   string `json:"servicePackage"`   // service
	ServiceImplPkg   string `json:"serviceImplPkg"`   // service.impl
	ControllerPkg    string `json:"controllerPkg"`    // controller

	// 生成开关
	GenEntity     bool `json:"genEntity"`
	GenMapper     bool `json:"genMapper"`
	GenService    bool `json:"genService"`
	GenController bool `json:"genController"`

	// 输出
	OutputDir string `json:"outputDir"`

	// 覆盖策略: overwrite / skip / ask
	OverwritePolicy string `json:"overwritePolicy"`

	// Redis 缓存
	EnableRedis bool   `json:"enableRedis"`
	RedisClient string `json:"redisClient"` // lettuce / jedis
	JsonLib     string `json:"jsonLib"`     // jackson / fastjson2 / gson
}

// TemplateData 传递给模板的数据
type TemplateData struct {
	Config

	// 当前表
	TableName    string // 数据库中原始表名
	TableComment string // 表注释
	EntityName   string // Java 类名 (PascalCase)
	EntityVar    string // 变量名 (camelCase)
	EntityPath   string // URL 路径名 (kebab-case)
	PrimaryKey   ColumnInfo

	// 包路径
	EntityPackageFull    string // com.example.entity
	MapperPackageFull    string
	ServicePackageFull   string
	ServiceImplPkgFull   string
	ControllerPkgFull    string
	ConfigPackageFull    string // com.example.config

	// 列
	Columns []ColumnInfo

	// 导入
	Imports []string

	// 命名空间 (javax / jakarta)
	Namespace string

	// 注入注解全名
	InjectAnnotation string

	// 分页类
	PageImport string

	// 复合主键
	HasCompositePK bool // 是否有复合主键

	// XML
	MapperXML string
}

// ColumnInfo 列信息
type ColumnInfo struct {
	ColName      string // 数据库列名
	ColComment   string // 列注释
	JavaType     string // Java 类型
	JavaField    string // Java 字段名 (camelCase)
	IsNullable   bool
	IsAutoInc    bool
	IsPrimaryKey bool
	PKOrder      int    // 主键次序: 0=非主键, 1=第一个PK, 2=后续PK(复合)
	IsCommonField bool // 是否为通用字段 (create_time, update_time, deleted 等)

	// MyBatis-Plus 相关
	MPStrategy string // IdType 策略
}

// ColumnInfoList 用于模板排序
type ColumnInfoList []ColumnInfo

func (l ColumnInfoList) NonPrimary() []ColumnInfo {
	var result []ColumnInfo
	for _, c := range l {
		if !c.IsPrimaryKey {
			result = append(result, c)
		}
	}
	return result
}

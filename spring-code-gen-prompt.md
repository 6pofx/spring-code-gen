# Spring 代码生成器 — 开发需求

## 概述
用 Go 实现一个 Spring 代码生成器（GUI 版），连接数据库读取表结构，生成 Entity / Mapper / Service / ServiceImpl / Controller 的 Java 源代码文件。

## 技术要求

### 语言和框架
- 语言：Go 1.22+
- GUI：web界面 + 本地 HTTP 服务器（内置 embed 前端页面，单文件 HTML/CSS/JS，不依赖 Electron）
- 数据库驱动：lib/pq（PostgreSQL）、go-sql-driver/mysql（MySQL）
- 模板引擎：Go 标准库 text/template
- 打包：go build 生成独立二进制

### 生成的文件
生成到用户指定的输出目录，每个文件独立：
- `${entityDir}/${entityName}.java` — Entity
- `${mapperDir}/${entityName}Mapper.java` — Mapper
- `${serviceDir}/${entityName}Service.java` — Service 接口
- `${serviceImplDir}/${entityName}ServiceImpl.java` — ServiceImpl
- `${controllerDir}/${entityName}Controller.java` — Controller

### 支持的数据库
- MySQL 5.7+ / 8.x
- PostgreSQL 10+

### 数据库元数据读取
通过 JDBC-like SQL 查询 information_schema / pg_catalog 获取：
- 表名、表注释
- 列名、列类型、列注释、是否可空、是否自增
- 主键信息
- 类型映射：数据库类型 → Java 类型

### Java 类型映射表
| 数据库类型 | Java 类型 |
|-----------|----------|
| INT / INTEGER / INT4 | Integer |
| BIGINT / INT8 / BIGSERIAL | Long |
| SMALLINT / INT2 | Integer |
| TINYINT | Integer |
| VARCHAR / CHAR / TEXT | String |
| BOOLEAN / BOOL | Boolean |
| DATE | java.time.LocalDate |
| DATETIME / TIMESTAMP / TIMESTAMPTZ | java.time.LocalDateTime |
| DECIMAL / NUMERIC | java.math.BigDecimal |
| FLOAT / FLOAT4 | Float |
| DOUBLE / FLOAT8 | Double |
| BLOB / BYTEA | byte[] |
| JSON / JSONB | String |
| 其他 | String |

### 可配置选项（GUI 界面中设置）

#### 1. 数据库连接
- 数据库类型：MySQL / PostgreSQL（下拉选择）
- 主机、端口（默认 MySQL 3306 / PostgreSQL 5432）
- 数据库名、用户名、密码
- 测试连接按钮

#### 2. 表选择
- 展示所有表名列表（多选）
- 表前缀过滤输入框（如 `sys_`，匹配的表会被列出）
- 排除前缀输入框（如 `flyway_,QRTZ_`，匹配的表不显示）

#### 3. ORM 选择
- MyBatis-Plus（默认）
- MyBatis
- JPA / Hibernate

#### 4. Spring Boot 版本
- 2.x（javax 命名空间）
- 3.x（jakarta 命名空间，默认）

#### 5. Swagger / API 文档
- SpringDoc OpenAPI 3（`@Schema` `@Operation` `@Tag`，默认）
- Swagger 2（`@ApiModel` `@ApiModelProperty` `@Api`）

#### 6. 统一响应类
- 可选生成的响应类名，默认 `R`
- 包含三个字段：`code int`, `message string`, `data T`
- 静态方法：`ok()`, `ok(data)`, `error(message)`（泛型）

#### 7. 注入方式
- `@Resource`（jakarta 或 javax，取决于 Spring Boot 版本，默认）
- `@Autowired`

#### 8. Lombok
- 启用（默认）：Entity 加 `@Data @EqualsAndHashCode @Accessors`
- 禁用：生成 getter/setter

#### 9. 包名配置
- 父包名（如 `com.example`）
- 子包名：entity / mapper / service / service.impl / controller（可自定义）

#### 10. 输出目录
- 输入框指定源代码输出路径

#### 11. 生成开关（复选框）
- [x] 生成 Entity
- [x] 生成 Mapper
- [x] 生成 Service / ServiceImpl
- [x] 生成 Controller

#### 12. 文件覆盖策略
- 覆盖已有文件
- 跳过已有文件
- 询问确认

### 控制器和生成器接口参考

#### Controller 基本实现
参考以下模式，支持分页查询、全部列表、按ID查询、新增、修改、删除、批量删除：

```java
@RestController
@RequestMapping("/{entityPath}")
@Tag(name = "{tableComment}")
public class {entityName}Controller {

    @Resource
    private {entityName}Service {entityVar}Service;

    @Operation(summary = "分页查询")
    @GetMapping("/page")
    public R<Page<{entityName}>> page(...) { }

    @Operation(summary = "全部列表")
    @GetMapping("/list")
    public R<List<{entityName}>> list() { }

    @Operation(summary = "根据ID查询")
    @GetMapping("/{id}")
    public R<{entityName}> get(@PathVariable Long id) { }

    @Operation(summary = "新增")
    @PostMapping
    public R<Void> add(@RequestBody {entityName} entity) { }

    @Operation(summary = "修改")
    @PutMapping
    public R<Void> update(@RequestBody {entityName} entity) { }

    @Operation(summary = "删除")
    @DeleteMapping("/{id}")
    public R<Void> delete(@PathVariable Long id) { }

    @Operation(summary = "批量删除")
    @DeleteMapping("/batch")
    public R<Void> deleteBatch(@RequestParam String ids) { }
}
```

#### Entity 基本实现
- MyBatis-Plus：`@TableName @TableId @TableField @Schema @Data`
- MyBatis：无注解，有对应 XML
- JPA：`@Entity @Table @Id @GeneratedValue @Column @Schema @Data`

#### Service 接口
- MyBatis-Plus：继承 `IService<Entity>`
- MyBatis / JPA：自定义方法签名

#### ServiceImpl
- MyBatis-Plus：继承 `ServiceImpl<Mapper, Entity>`
- MyBatis：继承/实现自定义 base
- JPA：注入 Repository

#### Mapper
- MyBatis-Plus：继承 `BaseMapper<Entity>` 加 `@Mapper`
- MyBatis：接口 + XML 映射文件
- JPA：继承 `JpaRepository<Entity, Long>` 或 `JpaSpecificationExecutor<Entity>`

### 模板实现
用 Go 的 `text/template`，每个文件类型一个模板字符串（硬编码在 Go 源码中，不依赖外部模板文件）。

模板需要根据配置条件化输出：
```go
{{if .UseLombok}}@Data
@EqualsAndHashCode(callSuper = false)
@Accessors(chain = true){{end}}
{{if eq .Orm "mybatis-plus"}}@TableName("{{.TableName}}"){{end}}
```

### GUI 界面设计
一个干净的 HTML 页面，分区域：
1. 顶部：标题 + 描述
2. 数据库连接区：类型下拉 + 主机/端口/库名/用户名/密码 + 测试连接按钮
3. 表选择区：前缀过滤输入框 + 排除前缀输入框 + 表多选列表
4. 代码配置区：用卡片分两列或手风琴：
   - ORM / Spring Boot 版本 / Swagger / 注入方式
   - 响应类名 / Lombok开关
   - 包名配置
   - 生成开关 / 覆盖策略
5. 输出目录 + 生成按钮 + 进度/日志输出区域

整体风格简洁，中文界面，适配移动端（响应式）。

### 程序入口
```go
func main() {
    // 启动 HTTP 服务器
    // 内置嵌入 index.html
    // 端口 9527（可配置）
    http.ListenAndServe(":9527", nil)
}
```

### 目录结构建议
```
spring-code-gen/
├── main.go              # 入口 + HTTP 路由
├── web/
│   └── index.html       # 嵌入的 GUI 页面
├── db/
│   ├── mysql.go         # MySQL 元数据读取
│   └── postgres.go      # PostgreSQL 元数据读取
├── gen/
│   ├── entity.go        # Entity 模板 + 生成
│   ├── mapper.go        # Mapper 模板 + 生成
│   ├── service.go       # Service 模板 + 生成
│   ├── service_impl.go  # ServiceImpl 模板 + 生成
│   ├── controller.go    # Controller 模板 + 生成
│   ├── config.go        # 配置结构体
│   └── types.go         # 类型映射 + 工具函数
└── go.mod
```

### 非功能需求
- 数据库连接失败给出友好中文提示
- 生成过程显示实时日志（每生成一个文件打印一条）
- 支持 CORS（可选，便于开发时前后端分离调试）
- 生成的代码格式化（至少缩进和空行正确）

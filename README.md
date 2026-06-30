# Spring 代码生成器

一个 Go 语言编写的 Web GUI 工具，通过连接数据库读取表结构，一键生成 Spring Boot 项目的 **Entity / Mapper / Service / ServiceImpl / Controller** 全套 Java 源代码。

同时提供 **pom.xml 依赖检查** 和 **application.yml 配置检查** 功能，确保生成代码的项目环境就绪。

## 快速开始

### 下载运行

直接运行二进制文件（无需安装 Go 环境）：

```bash
# Windows
spring-code-gen.exe

# 浏览器打开
http://localhost:9527
```

### 从源码构建

```bash
# 需要 Go 1.22+
go build -o spring-code-gen.exe .
./spring-code-gen.exe
```

默认端口 **9527**，浏览器访问 `http://localhost:9527` 即可打开 GUI。

---

## 功能总览

### 1. 代码生成（核心）

连接数据库 → 选择表 → 配置生成选项 → 一键生成

**支持的数据库：**
- MySQL 5.7+ / 8.x
- PostgreSQL 10+

**支持的 ORM 框架：**
- MyBatis-Plus（默认）
- MyBatis（含 XML 映射文件）
- JPA / Hibernate

**生成的代码：**
| 文件 | 说明 |
|------|------|
| `{entity}.java` | 实体类（含 @TableName、@TableId、Swagger 注解等） |
| `{entity}Mapper.java` | Mapper 接口（或 Repository） |
| `{entity}Service.java` | Service 接口 |
| `{entity}ServiceImpl.java` | Service 实现类 |
| `{entity}Controller.java` | REST Controller（CRUD + 分页 + 批量删除） |
| `R.java` | 统一响应类（可选） |

**可配置选项：**
- Spring Boot 2.x (javax) / 3.x (jakarta)
- SpringDoc OpenAPI 3 / Swagger 2
- @Resource / @Autowired 注入
- Lombok 启用/禁用
- 包名自定义
- 文件覆盖策略

### 2. pom.xml 依赖检查

根据代码生成配置，自动判断需要哪些 Maven 依赖，并与项目的 pom.xml 对比：

- `spring-boot-starter-web` — Web 支持
- `mysql-connector-j` / `postgresql` — 数据库驱动
- `mybatis-plus-spring-boot3-starter` / `mybatis-spring-boot-starter` / `spring-boot-starter-data-jpa` — ORM 框架
- `springdoc-openapi-starter-webmvc-ui` — API 文档
- `lombok` — 代码简化

**支持两种输入方式：**
- 指定 pom.xml 文件路径
- 粘贴 pom.xml 内容

**操作流程：**
1. 在页面上方配置好生成选项（ORM / 数据库 / Swagger 等）
2. 在"pom.xml 依赖检查"区域填入 pom.xml 路径或粘贴内容
3. 点击「匹配检查」— 对比所需依赖和现有依赖
4. 缺失的依赖会显示建议的 XML 片段
5. 点击「预览补全」查看改动，点击「写入 pom.xml」自动写入（原文件备份为 .bak）

### 3. application.yml/properties 配置检查

根据生成配置，检查项目的配置文件是否缺少必要的 Spring Boot 配置项：

- `server.port` — 服务端口
- `spring.datasource.url/username/password/driver-class-name` — 数据库连接
- `mybatis-plus.mapper-locations/type-aliases-package` — MyBatis-Plus 配置
- `spring.jpa.hibernate.ddl-auto/show-sql` — JPA 配置

**支持格式：**
- YAML（`application.yml` / `application.yaml`）
- Properties（`application.properties`）

**操作流程：**
1. 配置好生成选项
2. 填写配置文件路径（如 `src/main/resources/application.yml`）
3. 点击「检查」— 对比所需配置和现有配置
4. 缺失的配置项会高亮显示
5. 点击「预览补全」→「写入」自动追加配置

---

## GUI 界面卡片顺序

```
1. 数据库连接     → 连接数据库，测试连通性
2. 表选择         → 过滤、选择要生成的表
3. 代码配置       → ORM / Swagger / Lombok / Spring Boot 版本
4. 包名配置       → 包名路径
5. 生成选项       → 输出目录、覆盖策略、生成开关
6. pom.xml 依赖检查  → 检查并补全缺失的 Maven 依赖
7. 配置文件检查   → 检查并补全缺失的 application 配置
8. 生成代码       → 执行生成，实时日志
```

---

## 目录结构

```
spring-code-gen/
├── main.go              # 入口 + HTTP 路由 + embed 前端
├── web/
│   └── index.html       # 单页 GUI（HTML/CSS/JS，响应式）
├── db/
│   ├── db.go            # 数据库公共接口
│   ├── mysql.go         # MySQL 元数据读取
│   └── postgres.go      # PostgreSQL 元数据读取
├── gen/
│   ├── config.go        # 配置结构体
│   ├── types.go         # 类型映射 + 命名工具
│   ├── entity.go        # Entity 模板 + 生成器
│   ├── mapper.go        # Mapper 模板 + XML 模板
│   ├── service.go       # Service 接口模板
│   ├── service_impl.go  # ServiceImpl 模板
│   ├── controller.go    # Controller 模板
│   └── generator.go     # 生成编排 + 响应类模板
├── pomcheck/
│   ├── checker.go       # pom.xml 解析 + 本地仓库检查
│   ├── required.go      # 配置→依赖映射 + 匹配检查
│   └── autofix.go       # pom.xml 自动补全
└── appcheck/
    └── checker.go       # application.yml/properties 检查与补全
```

---

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/test-connection` | 测试数据库连接 |
| POST | `/api/tables` | 获取表列表 |
| POST | `/api/generate-response` | 生成统一响应类 |
| POST | `/api/generate-table` | 生成代码（SSE 流式） |
| POST | `/api/pom-check` | pom.xml 依赖安装检查 |
| POST | `/api/pom-check-content` | 粘贴内容依赖检查 |
| POST | `/api/pom-check-deps` | 生成配置 ↔ pom.xml 匹配检查 |
| POST | `/api/pom-auto-fix` | 自动补全 pom.xml |
| POST | `/api/app-check` | application 配置检查 |
| POST | `/api/app-auto-fix` | 自动补全 application 配置 |

---

## 构建

```bash
# 构建独立二进制
go build -o spring-code-gen.exe .

# 交叉编译（Linux）
GOOS=linux GOARCH=amd64 go build -o spring-code-gen-linux .

# 交叉编译（macOS）
GOOS=darwin GOARCH=amd64 go build -o spring-code-gen-macos .
```

### 依赖

- Go 1.22+
- 仅运行时依赖：无（前端已 embed 到二进制中）
- 构建依赖：`github.com/go-sql-driver/mysql`、`github.com/lib/pq`

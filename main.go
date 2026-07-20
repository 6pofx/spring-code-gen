package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"

	"spring-code-gen/appcheck"
	"spring-code-gen/db"
	"spring-code-gen/gen"
	"spring-code-gen/pomcheck"
)

//go:embed web/index.html
var webFS embed.FS

func main() {
	// 提取 index.html
	indexData, err := webFS.ReadFile("web/index.html")
	if err != nil {
		log.Fatalf("读取嵌入页面失败: %v", err)
	}

	// 路由
	mux := http.NewServeMux()

	// 静态页面
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexData)
	})

	// API
	mux.HandleFunc("/api/test-connection", handleTestConnection)
	mux.HandleFunc("/api/tables", handleTables)
	mux.HandleFunc("/api/generate-response", handleGenerateResponse)
	mux.HandleFunc("/api/generate-table", handleGenerateTable)
	mux.HandleFunc("/api/pom-check", handlePomCheck)
	mux.HandleFunc("/api/pom-check-content", handlePomCheckContent)
	mux.HandleFunc("/api/pom-check-deps", handlePomCheckDeps)
	mux.HandleFunc("/api/pom-auto-fix", handlePomAutoFix)
	mux.HandleFunc("/api/app-check", handleAppCheck)
	mux.HandleFunc("/api/app-auto-fix", handleAppAutoFix)

	// CORS 包装
	handler := corsMiddleware(mux)

	port := 9527
	log.Printf("=====================================")
	log.Printf("  Spring 代码生成器")
	log.Printf("  启动服务器: http://localhost:%d", port)
	log.Printf("=====================================")
	if err := http.ListenAndServe(":"+strconv.Itoa(port), handler); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// corsMiddleware 添加 CORS 支持
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// jsonResp 发送 JSON 响应
func jsonResp(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

// dbConnReq 数据库连接请求
type dbConnReq struct {
	DBType     string `json:"dbType"`
	DBHost     string `json:"dbHost"`
	DBPort     int    `json:"dbPort"`
	DBName     string `json:"dbName"`
	DBUser     string `json:"dbUser"`
	DBPassword string `json:"dbPassword"`
}

// tablesReq 表查询请求
type tablesReq struct {
	dbConnReq
	TablePrefix   string `json:"tablePrefix"`
	ExcludePrefix string `json:"excludePrefix"`
}

// 处理测试连接
func handleTestConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req dbConnReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 端口默认值
	if req.DBPort == 0 {
		if req.DBType == "postgresql" {
			req.DBPort = 5432
		} else {
			req.DBPort = 3306
		}
	}

	if err := db.TestConnection(req.DBType, req.DBHost, req.DBPort, req.DBName, req.DBUser, req.DBPassword); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "连接失败: " + err.Error()})
		return
	}

	jsonResp(w, map[string]interface{}{"success": true, "message": "连接成功"})
}

// 处理获取表列表
func handleTables(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req tablesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误"})
		return
	}

	if req.DBPort == 0 {
		if req.DBType == "postgresql" {
			req.DBPort = 5432
		} else {
			req.DBPort = 3306
		}
	}

	tables, err := db.GetTables(req.DBType, req.DBHost, req.DBPort, req.DBName, req.DBUser, req.DBPassword, req.TablePrefix, req.ExcludePrefix)
	if err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "获取表列表失败: " + err.Error()})
		return
	}

	// 转换为前端格式
	type tableItem struct {
		TableName    string `json:"tableName"`
		TableComment string `json:"tableComment"`
	}
	var items []tableItem
	for _, t := range tables {
		items = append(items, tableItem{TableName: t.TableName, TableComment: t.TableComment})
	}

	if items == nil {
		items = []tableItem{}
	}

	jsonResp(w, map[string]interface{}{"success": true, "tables": items})
}

// 生成响应类请求
type responseClassReq struct {
	OutputDir       string `json:"outputDir"`
	BasePackage     string `json:"basePackage"`
	ClassName       string `json:"className"`
	OverwritePolicy string `json:"overwritePolicy"`
}

// 处理生成响应类
func handleGenerateResponse(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req responseClassReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误"})
		return
	}

	msg, err := gen.WriteResponseClassFile(req.OutputDir, req.BasePackage, req.ClassName, req.OverwritePolicy)
	if err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	jsonResp(w, map[string]interface{}{"success": true, "message": msg})
}

// 生成单个表请求
type generateTableReq struct {
	Config    *gen.Config `json:"config"`
	TableName string      `json:"tableName"`
}

// 处理生成单个表（SSE 流式）
func handleGenerateTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req generateTableReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		fmt.Fprintf(w, "data: %s\n\n", jsonError("请求参数错误: "+err.Error()))
		return
	}

	cfg := req.Config
	tableName := req.TableName

	// 端口默认值
	if cfg.DBPort == 0 {
		if cfg.DBType == "postgresql" {
			cfg.DBPort = 5432
		} else {
			cfg.DBPort = 3306
		}
	}

	// 设置 SSE 头
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Fprintf(w, "data: %s\n\n", jsonError("不支持 SSE"))
		return
	}

	// 读取表列信息
	columns, err := db.GetColumns(cfg.DBType, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword, tableName)
	if err != nil {
		fmt.Fprintf(w, "data: %s\n\n", jsonError("读取表 "+tableName+" 列信息失败: "+err.Error()))
		flusher.Flush()
		return
	}

	// 获取表注释（通过列查询后间接获取）
	// 我们需要先获取表元数据
	tableComment := ""
	tables, err := db.GetTables(cfg.DBType, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPassword, "", "")
	if err == nil {
		for _, t := range tables {
			if t.TableName == tableName {
				tableComment = t.TableComment
				break
			}
		}
	}

	// 构建模板数据
	td := gen.BuildTemplateData(cfg, tableName, tableComment, columns)

	// 通过 channel 接收生成结果并 SSE 推送
	results := make(chan gen.Result, 10)
	go func() {
		gen.GenerateAll(td, cfg, tableName, results)
		close(results)
	}()

	for r := range results {
		data, _ := json.Marshal(r)
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}
}

func jsonError(msg string) string {
	data, _ := json.Marshal(gen.Result{Type: "error", Success: false, Message: msg})
	return string(data)
}

// pomCheckReq 依赖检查请求（文件路径）
type pomCheckReq struct {
	PomPath string `json:"pomPath"`
}

// pomCheckContentReq 依赖检查请求（粘贴内容）
type pomCheckContentReq struct {
	Content string `json:"content"`
}

// 处理 pom.xml 依赖检查（文件路径）
func handlePomCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req pomCheckReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误: " + err.Error()})
		return
	}

	if req.PomPath == "" {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请提供 pom.xml 文件路径"})
		return
	}

	report, err := pomcheck.CheckPomFile(req.PomPath)
	if err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	jsonResp(w, map[string]interface{}{"success": true, "report": report})
}

// 处理 pom.xml 依赖检查（粘贴内容）
func handlePomCheckContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req pomCheckContentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误: " + err.Error()})
		return
	}

	if req.Content == "" {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请粘贴 pom.xml 内容"})
		return
	}

	report, err := pomcheck.CheckPomContent([]byte(req.Content), "pasted pom.xml")
	if err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	jsonResp(w, map[string]interface{}{"success": true, "report": report})
}

// pomCheckDepsReq 依赖匹配检查请求
type pomCheckDepsReq struct {
	PomPath       string `json:"pomPath"`
	Content       string `json:"content"`
	DBType        string `json:"dbType"`
	Orm           string `json:"orm"`
	Swagger       string `json:"swagger"`
	SpringVersion string `json:"springVersion"`
	UseLombok     bool   `json:"useLombok"`
	EnableRedis   bool   `json:"enableRedis"`
	RedisClient   string `json:"redisClient"`
	JsonLib       string `json:"jsonLib"`
}

// 处理依赖匹配检查
func handlePomCheckDeps(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req pomCheckDepsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误: " + err.Error()})
		return
	}

	// 读取 pom.xml 依赖
	var existingDeps []pomcheck.Dependency

	if req.Content != "" {
		proj, err := pomcheck.ParsePOMContent([]byte(req.Content))
		if err != nil {
			jsonResp(w, map[string]interface{}{"success": false, "message": "解析 pom.xml 内容失败: " + err.Error()})
			return
		}
		existingDeps = proj.Dependencies
	} else if req.PomPath != "" {
		proj, err := pomcheck.ParsePOMFile(req.PomPath)
		if err != nil {
			jsonResp(w, map[string]interface{}{"success": false, "message": "读取 pom.xml 失败: " + err.Error()})
			return
		}
		existingDeps = proj.Dependencies
	} else {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请提供 pom.xml 文件路径或粘贴内容"})
		return
	}

	found, missing := pomcheck.CheckRequiredDeps(req.DBType, req.Orm, req.Swagger, req.SpringVersion, req.UseLombok, existingDeps, req.EnableRedis, req.RedisClient, req.JsonLib)

	jsonResp(w, map[string]interface{}{
		"success": true,
		"found":   found,
		"missing": missing,
	})
}

// autoFixReq 自动补全请求
type autoFixReq struct {
	PomPath    string              `json:"pomPath"`
	Content    string              `json:"content"`
	MissingDeps []pomcheck.RequiredDep `json:"missingDeps"`
	PreviewOnly bool               `json:"previewOnly"` // true=仅预览 false=写入
}

// 处理自动补全 pom.xml
func handlePomAutoFix(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req autoFixReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误: " + err.Error()})
		return
	}

	if len(req.MissingDeps) == 0 {
		jsonResp(w, map[string]interface{}{"success": false, "message": "没有需要补全的依赖"})
		return
	}

	// 文件路径模式
	if req.PomPath != "" {
		var result *pomcheck.AutoFixResult
		var err error

		if req.PreviewOnly {
			result, err = pomcheck.AutoFixPomPreview(req.PomPath, req.MissingDeps)
		} else {
			result, err = pomcheck.AutoFixPomWrite(req.PomPath, req.MissingDeps)
		}

		if err != nil {
			jsonResp(w, map[string]interface{}{"success": false, "message": err.Error()})
			return
		}

		jsonResp(w, map[string]interface{}{"success": true, "result": result})
		return
	}

	// 粘贴内容模式（仅预览，不写回文件）
	if req.Content != "" {
		newContent, count, err := pomcheck.InsertDepsIntoContent(req.Content, req.MissingDeps)
		if err != nil {
			jsonResp(w, map[string]interface{}{"success": false, "message": err.Error()})
			return
		}
		jsonResp(w, map[string]interface{}{
			"success": true,
			"result": &pomcheck.AutoFixResult{
				Preview:     newContent,
				InsertCount: count,
				Written:     false,
			},
		})
		return
	}

	jsonResp(w, map[string]interface{}{"success": false, "message": "请提供 pom.xml 文件路径或粘贴内容"})
}

// appCheckReq 配置文件检查请求
type appCheckReq struct {
	FilePath       string `json:"filePath"`
	DBType         string `json:"dbType"`
	Orm            string `json:"orm"`
	Swagger        string `json:"swagger"`
	SpringVersion  string `json:"springVersion"`
	EnableRedis    bool   `json:"enableRedis"`
}

// 处理配置文件检查
func handleAppCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req appCheckReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误"})
		return
	}

	if req.FilePath == "" {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请提供配置文件路径"})
		return
	}

	report, err := appcheck.CheckFile(req.FilePath, req.DBType, req.Orm, req.Swagger, req.SpringVersion, req.EnableRedis)
	if err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	jsonResp(w, map[string]interface{}{"success": true, "report": report})
}

// appAutoFixReq 配置文件补全请求
type appAutoFixReq struct {
	FilePath       string `json:"filePath"`
	DBType         string `json:"dbType"`
	Orm            string `json:"orm"`
	Swagger        string `json:"swagger"`
	SpringVersion  string `json:"springVersion"`
	PreviewOnly    bool   `json:"previewOnly"`
	EnableRedis    bool   `json:"enableRedis"`
}

// 处理配置文件补全
func handleAppAutoFix(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req appAutoFixReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请求参数错误"})
		return
	}

	if req.FilePath == "" {
		jsonResp(w, map[string]interface{}{"success": false, "message": "请提供配置文件路径"})
		return
	}

	var result *appcheck.AutoFixConfig
	var err error

	if req.PreviewOnly {
		result, err = appcheck.AutoFixPreview(req.FilePath, req.DBType, req.Orm, req.Swagger, req.SpringVersion, req.EnableRedis)
	} else {
		result, err = appcheck.AutoFixWrite(req.FilePath, req.DBType, req.Orm, req.Swagger, req.SpringVersion, req.EnableRedis)
	}

	if err != nil {
		jsonResp(w, map[string]interface{}{"success": false, "message": err.Error()})
		return
	}

	jsonResp(w, map[string]interface{}{"success": true, "result": result})
}

// 确保嵌入文件被使用（避免编译错误）
var _ = fs.FS(webFS)

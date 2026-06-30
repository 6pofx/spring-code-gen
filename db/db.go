package db

import (
	"database/sql"
	"fmt"
	"strings"

	"spring-code-gen/gen"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// ColumnMeta 列元数据
type ColumnMeta struct {
	ColName    string // 列名
	ColType    string // 数据库类型（含精度如 varchar(255)）
	ColComment string // 列注释
	IsNullable string // YES / NO
	IsAutoInc  string // YES / NO 或 true/false
	ColKey     string // PRI / MUL / ""
}

// TableMeta 表元数据
type TableMeta struct {
	TableName    string // 表名
	TableComment string // 表注释
}

// Connect 建立数据库连接
func Connect(dbType, host string, port int, dbName, user, password string) (*sql.DB, error) {
	var dsn string
	switch dbType {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			user, password, host, port, dbName)
	case "postgresql":
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbName)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
	return sql.Open(dbType, dsn)
}

// TestConnection 测试数据库连接
func TestConnection(dbType, host string, port int, dbName, user, password string) error {
	dbConn, err := Connect(dbType, host, port, dbName, user, password)
	if err != nil {
		return err
	}
	defer dbConn.Close()
	return dbConn.Ping()
}

// GetTables 获取表列表
func GetTables(dbType, host string, port int, dbName, user, password, prefix, excludePrefix string) ([]TableMeta, error) {
	dbConn, err := Connect(dbType, host, port, dbName, user, password)
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	var tables []TableMeta

	switch dbType {
	case "mysql":
		tables, err = getMySQLTables(dbConn, dbName)
	case "postgresql":
		tables, err = getPostgresTables(dbConn, dbName)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
	if err != nil {
		return nil, err
	}

	// 过滤
	var filtered []TableMeta
	excludes := parseExcludes(excludePrefix)

	for _, t := range tables {
		// 排除前缀检查
		skip := false
		for _, ex := range excludes {
			if strings.HasPrefix(strings.ToLower(t.TableName), strings.ToLower(ex)) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		// 包含前缀检查
		if prefix != "" && !strings.HasPrefix(strings.ToLower(t.TableName), strings.ToLower(prefix)) {
			continue
		}
		filtered = append(filtered, t)
	}

	return filtered, nil
}

// GetColumns 获取表列信息
func GetColumns(dbType, host string, port int, dbName, user, password, tableName string) ([]gen.ColumnInfo, error) {
	dbConn, err := Connect(dbType, host, port, dbName, user, password)
	if err != nil {
		return nil, err
	}
	defer dbConn.Close()

	var cols []ColumnMeta
	switch dbType {
	case "mysql":
		cols, err = getMySQLColumns(dbConn, dbName, tableName)
	case "postgresql":
		cols, err = getPostgresColumns(dbConn, dbName, tableName)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
	if err != nil {
		return nil, err
	}

	// 转换为 ColumnInfo
	var result []gen.ColumnInfo
	for _, c := range cols {
		javaType := gen.MapToJavaType(c.ColType)
		isPK := strings.EqualFold(c.ColKey, "PRI") || strings.EqualFold(c.ColKey, "P") || strings.Contains(strings.ToUpper(c.ColKey), "PRI")
		info := gen.ColumnInfo{
			ColName:       c.ColName,
			ColComment:    c.ColComment,
			JavaType:      javaType,
			JavaField:     gen.SnakeToCamel(c.ColName),
			IsPrimaryKey:  isPK,
			IsAutoInc:     c.IsAutoInc == "YES" || c.IsAutoInc == "true" || c.IsAutoInc == "1",
			IsNullable:    c.IsNullable == "YES" || c.IsNullable == "true" || c.IsNullable == "1",
			IsCommonField: gen.IsCommonField(c.ColName),
			MPStrategy:    gen.GetIdType(c.IsAutoInc == "YES" || c.IsAutoInc == "true" || c.IsAutoInc == "1"),
		}
		result = append(result, info)
	}

	// 设置 PKOrder（确保复合主键信息不丢失）
	pkIdx := 0
	for i := range result {
		if result[i].IsPrimaryKey {
			pkIdx++
			result[i].PKOrder = pkIdx
		}
	}

	return result, nil
}

// parseExcludes 解析排除前缀（逗号分隔）
func parseExcludes(excludePrefix string) []string {
	if excludePrefix == "" {
		return nil
	}
	parts := strings.Split(excludePrefix, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

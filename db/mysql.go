package db

import (
	"database/sql"
)

// getMySQLTables 获取 MySQL 表列表
func getMySQLTables(dbConn *sql.DB, dbName string) ([]TableMeta, error) {
	query := `SELECT TABLE_NAME, IFNULL(TABLE_COMMENT, '') TABLE_COMMENT
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME`
	rows, err := dbConn.Query(query, dbName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableMeta
	for rows.Next() {
		var t TableMeta
		if err := rows.Scan(&t.TableName, &t.TableComment); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

// getMySQLColumns 获取 MySQL 表列信息
func getMySQLColumns(dbConn *sql.DB, dbName, tableName string) ([]ColumnMeta, error) {
	query := `SELECT COLUMN_NAME, COLUMN_TYPE, IFNULL(COLUMN_COMMENT, '') COLUMN_COMMENT,
		IS_NULLABLE, IF(EXTRA LIKE '%auto_increment%','YES','NO') IS_AUTOINC, COLUMN_KEY
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`
	rows, err := dbConn.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []ColumnMeta
	for rows.Next() {
		var c ColumnMeta
		if err := rows.Scan(&c.ColName, &c.ColType, &c.ColComment, &c.IsNullable, &c.IsAutoInc, &c.ColKey); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

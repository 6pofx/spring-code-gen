package db

import (
	"database/sql"
	"fmt"
)

// getPostgresTables 获取 PostgreSQL 表列表
func getPostgresTables(dbConn *sql.DB, dbName string) ([]TableMeta, error) {
	query := `SELECT c.relname::VARCHAR, COALESCE(d.description, '')::VARCHAR AS table_comment
		FROM pg_class c
		LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
		LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = 0
		WHERE c.relkind = 'r' AND n.nspname = 'public'
		ORDER BY c.relname`
	rows, err := dbConn.Query(query)
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

// getPostgresColumns 获取 PostgreSQL 表列信息
func getPostgresColumns(dbConn *sql.DB, dbName, tableName string) ([]ColumnMeta, error) {
	query := fmt.Sprintf(`
		SELECT
			a.attname::VARCHAR AS column_name,
			COALESCE(tp.typname, '')::VARCHAR || 
				CASE 
					WHEN tp.typname IN ('varchar', 'char', 'character', 'character varying') 
						THEN '(' || a.atttypmod - 4 || ')'
					WHEN tp.typname = 'numeric' AND a.atttypmod > 0
						THEN '(' || ((a.atttypmod - 4) >> 16) || ',' || ((a.atttypmod - 4) & 65535) || ')'
					ELSE ''
				END AS column_type,
			COALESCE(d.description, '')::VARCHAR AS column_comment,
			CASE WHEN a.attnotnull THEN 'NO' ELSE 'YES' END AS is_nullable,
			CASE WHEN cs.relname IS NOT NULL THEN 'YES' ELSE 'NO' END AS is_autoinc,
			CASE WHEN con.contype = 'p' THEN 'PRI' ELSE '' END AS column_key
		FROM pg_attribute a
		JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_type tp ON tp.oid = a.atttypid
		LEFT JOIN pg_description d ON d.objoid = a.attrelid AND d.objsubid = a.attnum
		LEFT JOIN pg_constraint con ON con.conrelid = a.attrelid 
			AND con.contype = 'p' 
			AND a.attnum = ANY(con.conkey)
		LEFT JOIN pg_class cs ON cs.oid = (SELECT pg_get_serial_sequence('%s', a.attname)::regclass::oid 
			WHERE pg_get_serial_sequence('%s', a.attname) IS NOT NULL)
		WHERE a.attnum > 0 
			AND NOT a.attisdropped
			AND c.relname = '%s'
			AND c.relkind = 'r'
			AND c.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
		ORDER BY a.attnum
	`, tableName, tableName, tableName)

	rows, err := dbConn.Query(query)
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

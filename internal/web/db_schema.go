package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// forbiddenSQLRegex matches dangerous SQL keywords (word boundaries, case insensitive)
var forbiddenSQLRegex = regexp.MustCompile(`(?i)\b(DROP|TRUNCATE|ALTER|CREATE|DETACH|ATTACH|VACUUM|REINDEX)\b`)

// TableColumn represents a column in a database table
type TableColumn struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	NotNull      bool   `json:"not_null"`
	DefaultValue string `json:"default_value,omitempty"`
	PrimaryKey   bool   `json:"primary_key"`
}

// ForeignKey represents a foreign key relationship
type ForeignKey struct {
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
	OnDelete   string `json:"on_delete,omitempty"`
}

// TableInfo represents information about a database table
type TableInfo struct {
	Name       string       `json:"name"`
	Columns    []TableColumn `json:"columns"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
}

// SchemaResponse represents the complete database schema
type SchemaResponse struct {
	Tables        []TableInfo `json:"tables"`
	DBQueryAccess bool        `json:"db_query_access"`
}

// handleDBSchema returns the database schema information
func (r *Router) handleDBSchema(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schema, err := r.getDBSchema()
	if err != nil {
		r.logger.Error("Failed to get database schema", zap.Error(err))
		http.Error(w, "Failed to get database schema", http.StatusInternalServerError)
		return
	}
	schema.DBQueryAccess = r.config.Admin.DBQueryAccess

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schema); err != nil {
		r.logger.Error("Failed to encode schema response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getDBSchema extracts schema information from the database.
func (r *Router) getDBSchema() (*SchemaResponse, error) {
	tables, err := r.getTableNames()
	if err != nil {
		return nil, err
	}

	schema := &SchemaResponse{
		Tables: make([]TableInfo, 0, len(tables)),
	}

	for _, tableName := range tables {
		columns, err := r.getTableColumns(tableName)
		if err != nil {
			return nil, err
		}

		foreignKeys, err := r.getForeignKeys(tableName)
		if err != nil {
			return nil, err
		}

		schema.Tables = append(schema.Tables, TableInfo{
			Name:        tableName,
			Columns:     columns,
			ForeignKeys: foreignKeys,
		})
	}

	return schema, nil
}

// getTableNames returns all user table names.
func (r *Router) getTableNames() ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var name string
		_ = rows.Scan(&name)
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

// getTableColumns returns column information for a table.
func (r *Router) getTableColumns(tableName string) ([]TableColumn, error) {
	if !isValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	query := `
		SELECT
			c.column_name,
			c.data_type,
			CASE WHEN c.is_nullable = 'NO' THEN 1 ELSE 0 END as not_null,
			COALESCE(c.column_default, '') as column_default,
			CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN 1 ELSE 0 END as is_pk
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu
			ON c.table_schema = kcu.table_schema
			AND c.table_name = kcu.table_name
			AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc
			ON kcu.constraint_schema = tc.constraint_schema
			AND kcu.constraint_name = tc.constraint_name
			AND tc.constraint_type = 'PRIMARY KEY'
		WHERE c.table_schema = 'public' AND c.table_name = $1
		ORDER BY c.ordinal_position
	`

	rows, err := r.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]TableColumn, 0)
	for rows.Next() {
		var col TableColumn
		var notNull, pk int
		_ = rows.Scan(&col.Name, &col.Type, &notNull, &col.DefaultValue, &pk)
		col.NotNull = notNull == 1
		col.PrimaryKey = pk == 1
		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// getForeignKeys returns foreign key relationships for a table.
func (r *Router) getForeignKeys(tableName string) ([]ForeignKey, error) {
	if !isValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	query := `
		SELECT
			tc.table_name AS from_table,
			kcu.column_name AS from_column,
			ccu.table_name AS to_table,
			ccu.column_name AS to_column,
			rc.delete_rule AS on_delete
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_schema = kcu.constraint_schema
			AND tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_schema = tc.constraint_schema
			AND ccu.constraint_name = tc.constraint_name
		JOIN information_schema.referential_constraints rc
			ON rc.constraint_schema = tc.constraint_schema
			AND rc.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = 'public'
			AND tc.table_name = $1
		ORDER BY kcu.ordinal_position
	`

	rows, err := r.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	foreignKeys := make([]ForeignKey, 0)
	for rows.Next() {
		var fk ForeignKey
		_ = rows.Scan(&fk.FromTable, &fk.FromColumn, &fk.ToTable, &fk.ToColumn, &fk.OnDelete)
		foreignKeys = append(foreignKeys, fk)
	}

	return foreignKeys, rows.Err()
}

// isValidTableName validates that a table name is safe to use in SQL.
var tableNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func isValidTableName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	return tableNameRegex.MatchString(name)
}

// dbQueryRequest is the JSON body for POST /api/admin/db-query
type dbQueryRequest struct {
	Query string `json:"query"`
}

// dbQueryResponse is the JSON response for POST /api/admin/db-query
type dbQueryResponse struct {
	Rows         []map[string]interface{} `json:"rows,omitempty"`
	RowsAffected int64                    `json:"rows_affected,omitempty"`
	LastInsertID int64                    `json:"last_insert_id,omitempty"`
	Message      string                   `json:"message,omitempty"`
}

// handleDBQuery executes a safe SQL query (admin-only, DB_QUERY_ACCESS must be true).
// Blocks: DROP, TRUNCATE, ALTER, CREATE, DETACH, ATTACH, VACUUM, REINDEX.
func (r *Router) handleDBQuery(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.config.Admin.DBQueryAccess {
		http.Error(w, "DB query access is disabled", http.StatusForbidden)
		return
	}

	var body dbQueryRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	query := strings.TrimSpace(body.Query)
	if query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}
	const maxQueryLen = 50000
	if len(query) > maxQueryLen {
		http.Error(w, "Query too long", http.StatusBadRequest)
		return
	}

	// Reject multiple statements
	parts := strings.Split(query, ";")
	var nonEmpty int
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			nonEmpty++
		}
	}
	if nonEmpty > 1 {
		http.Error(w, "Only a single SQL statement is allowed", http.StatusBadRequest)
		return
	}

	// Block dangerous keywords
	if forbiddenSQLRegex.MatchString(query) {
		http.Error(w, "Forbidden SQL: DROP, TRUNCATE, ALTER, CREATE, DETACH, ATTACH, VACUUM, REINDEX are not allowed", http.StatusBadRequest)
		return
	}

	upper := strings.ToUpper(strings.TrimSpace(query))
	isSelect := strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "WITH")

	var resp dbQueryResponse
	if isSelect {
		rows, err := r.db.Query(query)
		if err != nil {
			r.logger.Warn("DB query error", zap.Error(err), zap.String("query", query))
			http.Error(w, "Query failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer rows.Close()

		cols, _ := rows.Columns()

		result := make([]map[string]interface{}, 0)
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		for rows.Next() {
			_ = rows.Scan(valPtrs...)
			row := make(map[string]interface{}, len(cols))
			for i, c := range cols {
				v := vals[i]
				if v == nil {
					row[c] = nil
				} else if b, ok := v.([]byte); ok {
					row[c] = string(b)
				} else {
					row[c] = v
				}
			}
			result = append(result, row)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Rows error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		resp.Rows = result
	} else {
		res, err := r.db.Exec(query)
		if err != nil {
			r.logger.Warn("DB exec error", zap.Error(err), zap.String("query", query))
			http.Error(w, "Query failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		resp.RowsAffected, _ = res.RowsAffected()
		resp.LastInsertID, _ = res.LastInsertId()
		resp.Message = "OK"
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		r.logger.Error("Failed to encode db-query response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

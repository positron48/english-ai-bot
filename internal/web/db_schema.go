package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

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
	Tables []TableInfo `json:"tables"`
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(schema); err != nil {
		r.logger.Error("Failed to encode schema response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getDBSchema extracts schema information from SQLite database
func (r *Router) getDBSchema() (*SchemaResponse, error) {
	// Get all table names
	tables, err := r.getTableNames()
	if err != nil {
		return nil, err
	}

	schema := &SchemaResponse{
		Tables: make([]TableInfo, 0, len(tables)),
	}

	for _, tableName := range tables {
		// Get columns for this table
		columns, err := r.getTableColumns(tableName)
		if err != nil {
			return nil, err
		}

		// Get foreign keys for this table
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

// getTableNames returns all table names in the database
func (r *Router) getTableNames() ([]string, error) {
	query := `
		SELECT name 
		FROM sqlite_master 
		WHERE type='table' 
		AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}

	return tables, rows.Err()
}

// getTableColumns returns column information for a table
func (r *Router) getTableColumns(tableName string) ([]TableColumn, error) {
	// Validate table name to prevent SQL injection
	if !isValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}
	query := `PRAGMA table_info(` + tableName + `)`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]TableColumn, 0)
	for rows.Next() {
		var col TableColumn
		var cid int
		var notNull, pk int
		var dfltValue sql.NullString

		if err := rows.Scan(&cid, &col.Name, &col.Type, &notNull, &dfltValue, &pk); err != nil {
			return nil, err
		}

		col.NotNull = notNull == 1
		col.PrimaryKey = pk == 1
		if dfltValue.Valid {
			col.DefaultValue = dfltValue.String
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// getForeignKeys returns foreign key relationships for a table
func (r *Router) getForeignKeys(tableName string) ([]ForeignKey, error) {
	// Validate table name to prevent SQL injection
	if !isValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}
	query := `PRAGMA foreign_key_list(` + tableName + `)`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	foreignKeys := make([]ForeignKey, 0)
	for rows.Next() {
		var fk ForeignKey
		var id, seq int
		var onUpdate sql.NullString
		var match sql.NullString

		if err := rows.Scan(&id, &seq, &fk.ToTable, &fk.FromColumn, &fk.ToColumn, &onUpdate, &fk.OnDelete, &match); err != nil {
			return nil, err
		}

		fk.FromTable = tableName

		// Only add each foreign key once (seq 0 is the first column)
		if seq == 0 {
			foreignKeys = append(foreignKeys, fk)
		}
	}

	return foreignKeys, rows.Err()
}

// isValidTableName validates that a table name is safe to use in SQL
// SQLite identifiers can contain letters, digits, underscores, and must not start with a digit
var tableNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func isValidTableName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	return tableNameRegex.MatchString(name)
}

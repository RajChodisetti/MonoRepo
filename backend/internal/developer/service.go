package developer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxRows      = 200
	defaultQueryTimeout = 8 * time.Second
	maxQueryLength      = 20_000
)

var (
	ErrDatabaseUnavailable = errors.New("developer database unavailable")
	ErrQueryRequired       = errors.New("sql query is required")
	ErrQueryTooLong        = errors.New("sql query is too long")
	ErrReadOnlyRequired    = errors.New("only read-only SQL queries are allowed")
)

type Service struct {
	pool    *pgxpool.Pool
	maxRows int
	timeout time.Duration
}

type SchemaTable struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Columns []SchemaColumn `json:"columns"`
}

type SchemaColumn struct {
	Name            string `json:"name"`
	DataType        string `json:"data_type"`
	DatabaseType    string `json:"database_type"`
	IsNullable      bool   `json:"is_nullable"`
	ColumnDefault   string `json:"column_default,omitempty"`
	OrdinalPosition int    `json:"ordinal_position"`
}

type QueryColumn struct {
	Name string `json:"name"`
}

type QueryResult struct {
	Columns   []QueryColumn `json:"columns"`
	Rows      [][]any       `json:"rows"`
	RowCount  int           `json:"row_count"`
	Truncated bool          `json:"truncated"`
	ElapsedMS int64         `json:"elapsed_ms"`
	MaxRows   int           `json:"max_rows"`
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:    pool,
		maxRows: defaultMaxRows,
		timeout: defaultQueryTimeout,
	}
}

func (service *Service) Schema(ctx context.Context) ([]SchemaTable, error) {
	if service == nil || service.pool == nil {
		return nil, ErrDatabaseUnavailable
	}

	const query = `
		SELECT
			c.table_name,
			t.table_type,
			c.column_name,
			c.data_type,
			c.udt_name,
			c.is_nullable = 'YES' AS is_nullable,
			COALESCE(c.column_default, '') AS column_default,
			c.ordinal_position
		FROM information_schema.columns AS c
		JOIN information_schema.tables AS t
			ON t.table_schema = c.table_schema
			AND t.table_name = c.table_name
		WHERE c.table_schema = 'public'
			AND t.table_type IN ('BASE TABLE', 'VIEW')
		ORDER BY c.table_name, c.ordinal_position`

	ctx, cancel := context.WithTimeout(ctx, service.timeout)
	defer cancel()

	rows, err := service.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query schema: %w", err)
	}
	defer rows.Close()

	tableIndexByName := map[string]int{}
	ordered := make([]SchemaTable, 0)
	for rows.Next() {
		var (
			tableName       string
			tableType       string
			column          SchemaColumn
			isNullable      bool
			columnDefault   string
			ordinalPosition int
		)
		if err := rows.Scan(
			&tableName,
			&tableType,
			&column.Name,
			&column.DataType,
			&column.DatabaseType,
			&isNullable,
			&columnDefault,
			&ordinalPosition,
		); err != nil {
			return nil, fmt.Errorf("scan schema: %w", err)
		}
		tableIndex, ok := tableIndexByName[tableName]
		if !ok {
			ordered = append(ordered, SchemaTable{Name: tableName, Type: tableType})
			tableIndex = len(ordered) - 1
			tableIndexByName[tableName] = tableIndex
		}
		column.IsNullable = isNullable
		column.ColumnDefault = columnDefault
		column.OrdinalPosition = ordinalPosition
		ordered[tableIndex].Columns = append(ordered[tableIndex].Columns, column)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}
	return ordered, nil
}

func (service *Service) Execute(ctx context.Context, rawQuery string) (QueryResult, error) {
	result := QueryResult{MaxRows: defaultMaxRows}
	query, err := NormalizeReadOnlyQuery(rawQuery)
	if err != nil {
		return result, err
	}
	if service == nil || service.pool == nil {
		return result, ErrDatabaseUnavailable
	}
	result.MaxRows = service.maxRows

	ctx, cancel := context.WithTimeout(ctx, service.timeout)
	defer cancel()

	started := time.Now()
	tx, err := service.pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, fmt.Errorf("begin read-only transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(context.Background())
	}()

	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '8000ms'"); err != nil {
		return result, fmt.Errorf("set statement timeout: %w", err)
	}

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return result, fmt.Errorf("execute query: %w", err)
	}

	fields := rows.FieldDescriptions()
	result.Columns = make([]QueryColumn, 0, len(fields))
	for _, field := range fields {
		result.Columns = append(result.Columns, QueryColumn{Name: string(field.Name)})
	}

	for rows.Next() {
		if len(result.Rows) >= service.maxRows {
			result.Truncated = true
			break
		}
		values, err := rows.Values()
		if err != nil {
			rows.Close()
			return result, fmt.Errorf("read row values: %w", err)
		}
		row := make([]any, len(values))
		for i, value := range values {
			row[i] = jsonSafeValue(value)
		}
		result.Rows = append(result.Rows, row)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("read query rows: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return result, fmt.Errorf("commit read-only transaction: %w", err)
	}

	result.RowCount = len(result.Rows)
	result.ElapsedMS = time.Since(started).Milliseconds()
	return result, nil
}

func NormalizeReadOnlyQuery(rawQuery string) (string, error) {
	query := strings.TrimSpace(rawQuery)
	if query == "" {
		return "", ErrQueryRequired
	}
	if len(query) > maxQueryLength {
		return "", ErrQueryTooLong
	}

	for strings.HasSuffix(query, ";") {
		query = strings.TrimSpace(strings.TrimSuffix(query, ";"))
	}
	if strings.Contains(query, ";") {
		return "", ErrReadOnlyRequired
	}

	keyword := firstSQLKeyword(query)
	switch keyword {
	case "select", "with", "show", "explain":
		return query, nil
	default:
		return "", ErrReadOnlyRequired
	}
}

func firstSQLKeyword(query string) string {
	rest := strings.TrimSpace(query)
	for {
		switch {
		case strings.HasPrefix(rest, "--"):
			if i := strings.IndexByte(rest, '\n'); i >= 0 {
				rest = strings.TrimSpace(rest[i+1:])
				continue
			}
			return ""
		case strings.HasPrefix(rest, "/*"):
			if i := strings.Index(rest, "*/"); i >= 0 {
				rest = strings.TrimSpace(rest[i+2:])
				continue
			}
			return ""
		default:
			for i, r := range rest {
				if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
					return strings.ToLower(rest[:i])
				}
			}
			return strings.ToLower(rest)
		}
	}
}

func jsonSafeValue(value any) any {
	switch v := value.(type) {
	case nil:
		return nil
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case []byte:
		return string(v)
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		if _, err := json.Marshal(v); err == nil {
			return v
		}
		return fmt.Sprint(v)
	}
}

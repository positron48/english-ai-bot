package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/stdlib"
)

var registerPostgresCompatOnce sync.Once

func registerPostgresCompatDriver() {
	registerPostgresCompatOnce.Do(func() {
		sql.Register("postgres_compat", &compatDriver{base: &stdlib.Driver{}})
	})
}

type compatDriver struct{ base driver.Driver }

func (d *compatDriver) Open(name string) (driver.Conn, error) {
	c, err := d.base.Open(name)
	if err != nil {
		return nil, err
	}
	return &compatConn{Conn: c}, nil
}

type compatConn struct{ driver.Conn }

func (c *compatConn) Prepare(query string) (driver.Stmt, error) {
	return c.Conn.Prepare(rebindQuestionToDollar(query))
}

func (c *compatConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if pc, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return pc.PrepareContext(ctx, rebindQuestionToDollar(query))
	}
	return c.Prepare(query)
}

func (c *compatConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	q := rebindQuestionToDollar(query)
	if ex, ok := c.Conn.(driver.ExecerContext); ok {
		return ex.ExecContext(ctx, q, args)
	}
	return nil, driver.ErrSkip
}

func (c *compatConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	q := rebindQuestionToDollar(query)
	if qr, ok := c.Conn.(driver.QueryerContext); ok {
		return qr.QueryContext(ctx, q, args)
	}
	return nil, driver.ErrSkip
}

func rebindQuestionToDollar(query string) string {
	if query == "" || !strings.Contains(query, "?") {
		return query
	}

	var b strings.Builder
	b.Grow(len(query) + 8)
	idx := 1
	inSingle := false
	inDouble := false

	for i := 0; i < len(query); i++ {
		r, sz := utf8.DecodeRuneInString(query[i:])
		if r == utf8.RuneError && sz == 1 {
			r = rune(query[i])
			sz = 1
		}

		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
			b.WriteRune(r)
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
			b.WriteRune(r)
		case '?':
			if inSingle || inDouble {
				b.WriteRune(r)
			} else {
				b.WriteByte('$')
				b.WriteString(strconv.Itoa(idx))
				idx++
			}
		default:
			b.WriteRune(r)
		}
		i += sz - 1
	}

	return b.String()
}

func openPostgresDB(dsn string) (*sql.DB, error) {
	registerPostgresCompatDriver()
	conn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres database: %w", err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}
	return conn, nil
}

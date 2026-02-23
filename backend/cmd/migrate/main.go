package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := strings.TrimSpace(os.Args[1])
	switch cmd {
	case "up":
		fs := flag.NewFlagSet("up", flag.ExitOnError)
		dir := fs.String("dir", "./migrations", "migrations directory containing *.up.sql")
		_ = fs.Parse(os.Args[2:])

		databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
		if databaseURL == "" {
			fatalf("DATABASE_URL is required")
		}

		conn, err := pgx.Connect(ctx, databaseURL)
		if err != nil {
			fatalf("connect: %v", err)
		}
		defer conn.Close(ctx)

		if err := ensureSchemaMigrations(ctx, conn); err != nil {
			fatalf("ensure schema_migrations: %v", err)
		}

		files, err := listMigrationFiles(*dir, ".up.sql")
		if err != nil {
			fatalf("list migrations: %v", err)
		}

		appliedCount := 0
		for _, path := range files {
			filename := filepath.Base(path)
			applied, err := isApplied(ctx, conn, filename)
			if err != nil {
				fatalf("check applied %s: %v", filename, err)
			}
			if applied {
				continue
			}

			sqlBytes, err := os.ReadFile(path)
			if err != nil {
				fatalf("read %s: %v", filename, err)
			}

			if err := applyMigration(ctx, conn, filename, string(sqlBytes)); err != nil {
				fatalf("apply %s: %v", filename, err)
			}

			fmt.Printf("applied %s\n", filename)
			appliedCount++
		}

		fmt.Printf("done (%d applied)\n", appliedCount)
	case "status":
		fs := flag.NewFlagSet("status", flag.ExitOnError)
		dir := fs.String("dir", "./migrations", "migrations directory containing *.up.sql")
		_ = fs.Parse(os.Args[2:])

		databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
		if databaseURL == "" {
			fatalf("DATABASE_URL is required")
		}

		conn, err := pgx.Connect(ctx, databaseURL)
		if err != nil {
			fatalf("connect: %v", err)
		}
		defer conn.Close(ctx)

		if err := ensureSchemaMigrations(ctx, conn); err != nil {
			fatalf("ensure schema_migrations: %v", err)
		}

		files, err := listMigrationFiles(*dir, ".up.sql")
		if err != nil {
			fatalf("list migrations: %v", err)
		}

		applied, err := listApplied(ctx, conn)
		if err != nil {
			fatalf("list applied: %v", err)
		}

		appliedNames := mapKeys(applied)
		fmt.Printf("applied migrations (%d):\n", len(appliedNames))
		for _, name := range appliedNames {
			fmt.Printf("  - %s\n", name)
		}

		pendingCount := 0
		for _, path := range files {
			name := filepath.Base(path)
			if applied[name] {
				continue
			}
			pendingCount++
		}

		fmt.Printf("pending migrations (%d):\n", pendingCount)
		for _, path := range files {
			name := filepath.Base(path)
			if applied[name] {
				continue
			}
			fmt.Printf("  - %s\n", name)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/migrate up [-dir ./migrations]")
	fmt.Fprintln(os.Stderr, "  go run ./cmd/migrate status [-dir ./migrations]")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func ensureSchemaMigrations(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
		  filename TEXT PRIMARY KEY,
		  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	return err
}

func listMigrationFiles(dir string, suffix string) ([]string, error) {
	glob := filepath.Join(dir, "*"+suffix)
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return matches, nil
}

func isApplied(ctx context.Context, conn *pgx.Conn, filename string) (bool, error) {
	var exists bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = $1)`, filename).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func listApplied(ctx context.Context, conn *pgx.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, `SELECT filename FROM schema_migrations ORDER BY filename ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out[name] = true
	}
	return out, rows.Err()
}

func applyMigration(ctx context.Context, conn *pgx.Conn, filename string, sql string) error {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return fmt.Errorf("empty migration")
	}

	wrapped := strings.Builder{}
	wrapped.WriteString("BEGIN;\n")
	wrapped.WriteString(sql)
	if !strings.HasSuffix(sql, ";") {
		wrapped.WriteString(";\n")
	} else {
		wrapped.WriteString("\n")
	}
	wrapped.WriteString("INSERT INTO schema_migrations (filename, applied_at) VALUES ('")
	wrapped.WriteString(strings.ReplaceAll(filename, "'", "''"))
	wrapped.WriteString("', now()) ON CONFLICT (filename) DO NOTHING;\n")
	wrapped.WriteString("COMMIT;\n")

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := execMulti(ctx, conn, wrapped.String())
	if err == nil {
		return nil
	}

	_ = execMulti(context.Background(), conn, "ROLLBACK;")
	return err
}

func execMulti(ctx context.Context, conn *pgx.Conn, sql string) error {
	results, err := conn.PgConn().Exec(ctx, sql).ReadAll()
	if err != nil {
		return err
	}
	for _, res := range results {
		if res.Err != nil {
			return res.Err
		}
	}
	return nil
}

func mapKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

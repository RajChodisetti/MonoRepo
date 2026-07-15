package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up|down]")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(ctx)

	dir := "migrations"
	if envDir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); envDir != "" {
		dir = envDir
	}

	switch os.Args[1] {
	case "up":
		if err := runMigrations(ctx, conn, dir, "up"); err != nil {
			log.Fatal(err)
		}
		fmt.Println("migrations applied")
	case "down":
		if err := runMigrations(ctx, conn, dir, "down"); err != nil {
			log.Fatal(err)
		}
		fmt.Println("migrations rolled back")
	default:
		log.Fatal("usage: migrate [up|down]")
	}
}

func runMigrations(ctx context.Context, conn *pgx.Conn, dir, direction string) error {
	suffix := "." + direction + ".sql"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), suffix) {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if direction == "down" {
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	}

	for _, name := range files {
		path := filepath.Join(dir, name)
		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if _, err := conn.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("exec %s: %w", name, err)
		}
		log.Printf("applied %s", name)
	}
	return nil
}

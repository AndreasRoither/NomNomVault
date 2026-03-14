package main

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/AndreasRoither/NomNomVault/backend/internal/ent"
	"github.com/AndreasRoither/NomNomVault/backend/internal/ent/generated/migrate"
)

func main() {
	var (
		dsn          = flag.String("dsn", "", "PostgreSQL DSN")
		requireClean = flag.Bool("require-clean", false, "exit non-zero when schema drift exists")
	)

	flag.Parse()

	if *dsn == "" {
		log.Fatal("dsn is required")
	}

	db, err := sql.Open("pgx", *dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	driver := entsql.OpenDB("postgres", db)
	client := ent.NewClient(ent.Driver(driver))
	defer client.Close()

	var buffer bytes.Buffer
	if err := client.Schema.WriteTo(
		context.Background(),
		&buffer,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		log.Fatalf("diff schema: %v", err)
	}

	diff := strings.TrimSpace(buffer.String())
	if diff == "" {
		return
	}

	fmt.Fprintln(os.Stdout, diff)

	if *requireClean {
		os.Exit(1)
	}
}

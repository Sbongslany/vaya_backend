package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/yourorg/ehailing/backend/internal/config"
	"github.com/yourorg/ehailing/backend/internal/database"
)

func main() {
	_ = godotenv.Load()
	cfg, _ := config.Load()
	pool, err := database.NewPostgres(context.Background(), cfg.Postgres)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	rows, err := pool.Query(context.Background(), `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'payouts'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Columns in the 'payouts' table:")
	for rows.Next() {
		var name, dtype string
		rows.Scan(&name, &dtype)
		fmt.Printf("- %s (%s)\n", name, dtype)
	}
}

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
	if err != nil { log.Fatal("DB Error:", err) }
	defer pool.Close()

	fmt.Println("--- TESTING OVERVIEW QUERY ---")
	var tu, td, tp, tt, at int
	var tr float64
	err = pool.QueryRow(context.Background(), `
		SELECT 
			(SELECT COUNT(*) FROM auth.users),
			(SELECT COUNT(DISTINCT ur.user_id) FROM auth.user_roles ur JOIN auth.roles r ON ur.role_id = r.id WHERE r.name = 'DRIVER'),
			(SELECT COUNT(DISTINCT ur.user_id) FROM auth.user_roles ur JOIN auth.roles r ON ur.role_id = r.id WHERE r.name = 'PASSENGER'),
			(SELECT COUNT(*) FROM trips),
			(SELECT COUNT(*) FROM trips WHERE status IN ('REQUESTED', 'DRIVER_ASSIGNED', 'DRIVER_EN_ROUTE', 'ARRIVED_AT_PICKUP', 'IN_PROGRESS')),
			COALESCE((SELECT SUM(final_fare) FROM trips WHERE status IN ('PAYMENT_COMPLETED', 'TRIP_COMPLETED')), 0)
	`).Scan(&tu, &td, &tp, &tt, &at, &tr)
	if err != nil {
		fmt.Println("❌ OVERVIEW ERROR:", err)
	} else {
		fmt.Println("✅ OVERVIEW SUCCESS")
	}

	fmt.Println("\n--- TESTING USERS QUERY ---")
	rows, err := pool.Query(context.Background(), `
		SELECT 
			u.id, 
			u.email, 
			COALESCE((SELECT r.name FROM auth.user_roles ur JOIN auth.roles r ON ur.role_id = r.id WHERE ur.user_id = u.id LIMIT 1), 'UNKNOWN') as role, 
			u.status, 
			u.created_at 
		FROM auth.users u LIMIT 5
	`)
	if err != nil {
		fmt.Println("❌ USERS QUERY ERROR:", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id, email, role, status, created string
			err := rows.Scan(&id, &email, &role, &status, &created)
			if err != nil {
				fmt.Println("❌ USERS SCAN ERROR:", err)
			} else {
				fmt.Println("✅ USERS ROW:", id, email, role, status)
			}
		}
	}

	fmt.Println("\n--- TESTING TRIPS QUERY ---")
	rows2, err := pool.Query(context.Background(), `SELECT id, passenger_id, driver_id, status, trip_type, estimated_fare, final_fare, created_at FROM trips LIMIT 5`)
	if err != nil {
		fmt.Println("❌ TRIPS QUERY ERROR:", err)
	} else {
		defer rows2.Close()
		for rows2.Next() {
			var id, pid, did, status, ttype, created string
			var est, final float64
			err := rows2.Scan(&id, &pid, &did, &status, &ttype, &est, &final, &created)
			if err != nil {
				fmt.Println("❌ TRIPS SCAN ERROR:", err)
			} else {
				fmt.Println("✅ TRIPS ROW:", id, status)
			}
		}
	}
}
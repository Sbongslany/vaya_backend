package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"

	"github.com/yourorg/ehailing/backend/internal/auth/infrastructure/security"
	"github.com/yourorg/ehailing/backend/internal/config"
	"github.com/yourorg/ehailing/backend/internal/database"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := database.NewPostgres(context.Background(), cfg.Postgres)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	email := "admin@vaya.co.za"
	password := "Admin@123"

	fmt.Println("--- FIXING PASSWORD WITH ARGON2 ---")

	// 1. Generate Argon2 hash using YOUR exact password service
	passwordSvc := security.NewPasswordService()
	hash, err := passwordSvc.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	fmt.Println("✅ Argon2 Hash Generated Successfully")

	// 2. Find User
	var userID uuid.UUID
	err = pool.QueryRow(context.Background(), "SELECT id FROM auth.users WHERE email = $1", email).Scan(&userID)

	if err == pgx.ErrNoRows {
		fmt.Println("Creating user...")
		userID = uuid.New()
		_, err = pool.Exec(context.Background(),
			"INSERT INTO auth.users (id, first_name, last_name, email, phone, password_hash, status, created_at, updated_at) VALUES ($1, $2, $3, $4, NULL, $5, 'ACTIVE', NOW(), NOW())",
			userID, "Super", "Admin", email, hash)
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		fmt.Println("✅ User created!")
	} else if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("User found. Updating hash...")
		_, err = pool.Exec(context.Background(), "UPDATE auth.users SET password_hash = $1, status = 'ACTIVE' WHERE id = $2", hash, userID)
		if err != nil {
			log.Fatalf("Failed to update: %v", err)
		}
		fmt.Println("✅ Password updated with Argon2 hash!")
	}

	// 3. Ensure Role
	var roleID int
	err = pool.QueryRow(context.Background(), "SELECT id FROM auth.roles WHERE name = 'SUPER_ADMIN'").Scan(&roleID)
	if err == pgx.ErrNoRows {
		err = pool.QueryRow(context.Background(), "INSERT INTO auth.roles (name, description) VALUES ('SUPER_ADMIN', 'Full access') RETURNING id").Scan(&roleID)
		if err != nil {
			log.Fatal(err)
		}
	}

	var exists bool
	pool.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM auth.user_roles WHERE user_id = $1 AND role_id = $2)", userID, roleID).Scan(&exists)
	if !exists {
		pool.Exec(context.Background(), "INSERT INTO auth.user_roles (user_id, role_id) VALUES ($1, $2)", userID, roleID)
		fmt.Println("✅ SUPER_ADMIN role assigned.")
	}

	fmt.Println("\n🎉 DONE. Restart your backend and try logging in.")
}

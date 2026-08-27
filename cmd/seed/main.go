package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/ehailing/backend/internal/config"
	"github.com/yourorg/ehailing/backend/internal/database"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env not found, using defaults")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Connect to Postgres
	pgPool, err := database.NewPostgres(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer pgPool.Close()

	// Admin credentials
	adminEmail := "admin@vaya.co.za"
	adminPassword := "Admin@123"
	adminFirstName := "Super"
	adminLastName := "Admin"

	// Check if admin already exists
	var exists bool
	err = pgPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)", adminEmail).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check user: %v", err)
	}
	if exists {
		fmt.Println("✅ Admin already exists!")
		fmt.Printf("   Email: %s\n", adminEmail)
		fmt.Printf("   Password: %s\n", adminPassword)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create user
	userID := uuid.New()
	userQuery := `
		INSERT INTO auth.users (id, first_name, last_name, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'ACTIVE', NOW(), NOW())
	`
	_, err = pgPool.Exec(ctx, userQuery, userID, adminFirstName, adminLastName, adminEmail, string(hashedPassword))
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Println("✅ User created successfully!")

	// Assign SUPER_ADMIN role
	roleQuery := `
		INSERT INTO auth.user_roles (user_id, role_id)
		SELECT $1, id FROM auth.roles WHERE name = 'SUPER_ADMIN'
	`
	_, err = pgPool.Exec(ctx, roleQuery, userID)
	if err != nil {
		log.Fatalf("Failed to assign role: %v", err)
	}
	fmt.Println("✅ SUPER_ADMIN role assigned!")

	fmt.Println("\n========================================")
	fmt.Println("   🎉 Your Admin Account is Ready!")
	fmt.Println("========================================")
	fmt.Printf("   Email:    %s\n", adminEmail)
	fmt.Printf("   Password: %s\n", adminPassword)
	fmt.Println("========================================")
	fmt.Println("\nYou can now log in at http://localhost:3000/login")
}

package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"github.com/yourorg/ehailing/backend/internal/auth/domain"
	"github.com/yourorg/ehailing/backend/internal/auth/domain/entities"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO auth.users (id, first_name, last_name, email, phone, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.FirstName, user.LastName, user.Email, user.Phone,
		user.PasswordHash, user.Status, user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return domain.ErrUserAlreadyExists
		}
		return err
	}

	// Assign PASSENGER role by default
	roleQuery := `
		INSERT INTO auth.user_roles (user_id, role_id)
		SELECT $1, id FROM auth.roles WHERE name = 'PASSENGER'
	`
	_, err = r.pool.Exec(ctx, roleQuery, user.ID)
	return err
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	user := &entities.User{}
	query := `
		SELECT id, first_name, last_name, email, phone, password_hash, status, 
		       email_verified_at, phone_verified_at, created_at, updated_at
		FROM auth.users WHERE id = $1
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
		&user.PasswordHash, &user.Status, &user.EmailVerifiedAt, &user.PhoneVerifiedAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	// Load roles
	roles, err := r.GetRolesByUserID(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, roleName := range roles {
		user.Roles = append(user.Roles, domain.Role(roleName))
	}

	return user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	user := &entities.User{}
	query := `
		SELECT id, first_name, last_name, email, phone, password_hash, status, 
		       email_verified_at, phone_verified_at, created_at, updated_at
		FROM auth.users WHERE email = $1
	`
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
		&user.PasswordHash, &user.Status, &user.EmailVerifiedAt, &user.PhoneVerifiedAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	roles, err := r.GetRolesByUserID(ctx, user.ID)
	if err != nil { return nil, err }
	for _, roleName := range roles {
		user.Roles = append(user.Roles, domain.Role(roleName))
	}

	return user, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*entities.User, error) {
	user := &entities.User{}
	query := `
		SELECT id, first_name, last_name, email, phone, password_hash, status, 
		       email_verified_at, phone_verified_at, created_at, updated_at
		FROM auth.users WHERE phone = $1
	`
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Phone,
		&user.PasswordHash, &user.Status, &user.EmailVerifiedAt, &user.PhoneVerifiedAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	roles, err := r.GetRolesByUserID(ctx, user.ID)
	if err != nil { return nil, err }
	for _, roleName := range roles {
		user.Roles = append(user.Roles, domain.Role(roleName))
	}

	return user, nil
}

func (r *UserRepository) GetRolesByUserID(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT r.name FROM auth.roles r
		INNER JOIN auth.user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}
	return roles, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.users SET password_hash = $1, updated_at = $2 WHERE id = $3", passwordHash, time.Now(), id)
	return err
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.users SET status = $1, updated_at = $2 WHERE id = $3", status, time.Now(), id)
	return err
}

func (r *UserRepository) UpdateEmailVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.users SET email_verified_at = $1, updated_at = $1 WHERE id = $2", time.Now(), id)
	return err
}

func (r *UserRepository) UpdatePhoneVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "UPDATE auth.users SET phone_verified_at = $1, updated_at = $1 WHERE id = $2", time.Now(), id)
	return err
}

func (r *UserRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID int) error {
	query := `
		INSERT INTO auth.user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userID, roleID)
	return err
}
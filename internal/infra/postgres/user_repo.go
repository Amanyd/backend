package postgres

import (
	"context"

	"github.com/eulerbutcooler/wingman-backend/internal/domain"
	"github.com/eulerbutcooler/wingman-backend/internal/infra/postgres/gen"
	"github.com/eulerbutcooler/wingman-backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Wraps sqlc's generated queries and satisfies port.UserRepository
type userRepo struct {
	q *gen.Queries
}

// Constructor. Takes pgxpool and returns the interface
func NewUserRepo(pool *pgxpool.Pool) port.UserRepository {
	return &userRepo{q: gen.New(pool)}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	row, err := r.q.CreateUser(ctx, gen.CreateUserParams{
		Name:         user.Name,
		EnrollmentID: user.EnrollmentID,
		Rank:         user.Rank,
		Batch:        user.Batch,
		Role:         string(user.Role),
		PasswordHash: user.PasswordHash,
	})
	if err != nil {
		return err
	}
	user.ID = row.ID
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainUser(row), nil
}

func (r *userRepo) GetByEnrollmentID(ctx context.Context, enrollmentID string) (*domain.User, error) {
	row, err := r.q.GetUserByEnrollmentID(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}
	return toDomainUser(row), nil
}

// Converts a sqlc-generated gen.User into our domain.User
// Lives in infra layer
func toDomainUser(user gen.User) *domain.User {
	return &domain.User{
		ID:           user.ID,
		Name:         user.Name,
		EnrollmentID: user.EnrollmentID,
		Rank:         user.Rank,
		Batch:        user.Batch,
		Role:         domain.Role(user.Role),
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

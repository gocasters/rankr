package contributor

import (
    "context"
)

// Repository defines the interface for contributor data operations
type Repository interface {
    Create(ctx context.Context, contributor *Contributor) error
    GetByID(ctx context.Context, id string) (*Contributor, error)
    GetByUsername(ctx context.Context, username string) (*Contributor, error)
    GetByEmail(ctx context.Context, email string) (*Contributor, error)
    Update(ctx context.Context, contributor *Contributor) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]*Contributor, error)
    Exists(ctx context.Context, id string) (bool, error)
    ExistsByUsername(ctx context.Context, username string) (bool, error)
    ExistsByEmail(ctx context.Context, email string) (bool, error)
    Count(ctx context.Context) (int, error)
}

// CacheRepository defines the interface for contributor cache operations
type CacheRepository interface {
    GetByID(ctx context.Context, id string) (*Contributor, error)
    SetByID(ctx context.Context, contributor *Contributor) error
    DeleteByID(ctx context.Context, id string) error
    Clear(ctx context.Context) error
}

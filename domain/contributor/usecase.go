package contributor

import (
    "context"
    "errors"
    "fmt"
)

var (
    ErrContributorNotFound = errors.New("contributor not found")
    ErrInvalidInput       = errors.New("invalid input")
    ErrContributorExists  = errors.New("contributor already exists")
)

// UseCase struct for contributor business logic
type UseCase struct {
    repo  Repository
    cache CacheRepository
}

func NewUseCase(repo Repository, cache CacheRepository) *UseCase {
    return &UseCase{
        repo:  repo,
        cache: cache,
    }
}

func (uc *UseCase) CreateContributor(ctx context.Context, username, email, displayName string) (*Contributor, error) {
    if username == "" || email == "" || displayName == "" {
        return nil, ErrInvalidInput
    }

    // Check if contributor already exists
    exists, err := uc.repo.ExistsByUsername(ctx, username)
    if err != nil {
        return nil, fmt.Errorf("failed to check username existence: %w", err)
    }
    if exists {
        return nil, ErrContributorExists
    }

    exists, err = uc.repo.ExistsByEmail(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("failed to check email existence: %w", err)
    }
    if exists {
        return nil, ErrContributorExists
    }

    contributor := NewContributor(username, email, displayName)

    err = uc.repo.Create(ctx, contributor)
    if err != nil {
        return nil, fmt.Errorf("failed to create contributor: %w", err)
    }

    // Cache the contributor
    if uc.cache != nil {
        _ = uc.cache.SetByID(ctx, contributor)
    }

    return contributor, nil
}

func (uc *UseCase) GetContributor(ctx context.Context, id string) (*Contributor, error) {
    if id == "" {
        return nil, ErrInvalidInput
    }

    // Try to get from cache first
    if uc.cache != nil {
        cached, err := uc.cache.GetByID(ctx, id)
        if err == nil {
            return cached, nil
        }
    }

    contributor, err := uc.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, ErrContributorNotFound) {
            return nil, ErrContributorNotFound
        }
        return nil, fmt.Errorf("failed to get contributor: %w", err)
    }

    // Cache the contributor
    if uc.cache != nil {
        _ = uc.cache.SetByID(ctx, contributor)
    }

    return contributor, nil
}

func (uc *UseCase) UpdateContributor(ctx context.Context, id string, update *ContributorUpdate) (*Contributor, error) {
    if id == "" || update == nil {
        return nil, ErrInvalidInput
    }

    contributor, err := uc.GetContributor(ctx, id)
    if err != nil {
        return nil, err
    }

    // Update fields
    if update.Username != "" {
        contributor.Username = update.Username
    }
    if update.Email != "" {
        contributor.Email = update.Email
    }
    if update.DisplayName != "" {
        contributor.DisplayName = update.DisplayName
    }
    if update.AvatarURL != "" {
        contributor.AvatarURL = update.AvatarURL
    }
    if update.GitHubID != "" {
        contributor.GitHubID = update.GitHubID
    }
    if update.IsActive != nil {
        contributor.IsActive = *update.IsActive
    }

    err = uc.repo.Update(ctx, contributor)
    if err != nil {
        return nil, fmt.Errorf("failed to update contributor: %w", err)
    }

    // Update cache
    if uc.cache != nil {
        _ = uc.cache.SetByID(ctx, contributor)
    }

    return contributor, nil
}

func (uc *UseCase) DeleteContributor(ctx context.Context, id string) error {
    if id == "" {
        return ErrInvalidInput
    }

    // Check if contributor exists
    _, err := uc.GetContributor(ctx, id)
    if err != nil {
        return err
    }

    err = uc.repo.Delete(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to delete contributor: %w", err)
    }

    // Remove from cache
    if uc.cache != nil {
        _ = uc.cache.DeleteByID(ctx, id)
    }

    return nil
}

func (uc *UseCase) ListContributors(ctx context.Context, limit, offset int) ([]*Contributor, error) {
    if limit <= 0 {
        limit = 10
    }
    if offset < 0 {
        offset = 0
    }

    contributors, err := uc.repo.List(ctx, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to list contributors: %w", err)
    }

    return contributors, nil
}

func (uc *UseCase) GetContributorByUsername(ctx context.Context, username string) (*Contributor, error) {
    if username == "" {
        return nil, ErrInvalidInput
    }

    contributor, err := uc.repo.GetByUsername(ctx, username)
    if err != nil {
        if errors.Is(err, ErrContributorNotFound) {
            return nil, ErrContributorNotFound
        }
        return nil, fmt.Errorf("failed to get contributor by username: %w", err)
    }

    return contributor, nil
}

func (uc *UseCase) GetContributorByEmail(ctx context.Context, email string) (*Contributor, error) {
    if email == "" {
        return nil, ErrInvalidInput
    }

    contributor, err := uc.repo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, ErrContributorNotFound) {
            return nil, ErrContributorNotFound
        }
        return nil, fmt.Errorf("failed to get contributor by email: %w", err)
    }

    return contributor, nil
}

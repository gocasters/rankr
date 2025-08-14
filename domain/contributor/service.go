package contributor

import (
    "context"
)

// Service provides contributor operations
type Service struct {
    useCase *UseCase
}

// NewService creates a new contributor service
func NewService(useCase *UseCase) *Service {
    return &Service{useCase: useCase}
}

// CreateContributor creates a new contributor
func (s *Service) CreateContributor(ctx context.Context, username, email, displayName string) (*Contributor, error) {
    return s.useCase.CreateContributor(ctx, username, email, displayName)
}

// GetContributor retrieves a contributor by ID
func (s *Service) GetContributor(ctx context.Context, id string) (*Contributor, error) {
    return s.useCase.GetContributor(ctx, id)
}

// UpdateContributor updates an existing contributor
func (s *Service) UpdateContributor(ctx context.Context, id string, update *ContributorUpdate) (*Contributor, error) {
    return s.useCase.UpdateContributor(ctx, id, update)
}

// DeleteContributor deletes a contributor by ID
func (s *Service) DeleteContributor(ctx context.Context, id string) error {
    return s.useCase.DeleteContributor(ctx, id)
}

// ListContributors returns a list of contributors
func (s *Service) ListContributors(ctx context.Context, limit, offset int) ([]*Contributor, error) {
    return s.useCase.ListContributors(ctx, limit, offset)
}

// GetContributorByUsername retrieves a contributor by username
func (s *Service) GetContributorByUsername(ctx context.Context, username string) (*Contributor, error) {
    return s.useCase.GetContributorByUsername(ctx, username)
}

// GetContributorByEmail retrieves a contributor by email
func (s *Service) GetContributorByEmail(ctx context.Context, email string) (*Contributor, error) {
    return s.useCase.GetContributorByEmail(ctx, email)
}

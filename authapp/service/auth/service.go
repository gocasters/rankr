package auth

import (
	"context"
	"errors"
	"time"

	"github.com/gocasters/rankr/pkg/cachemanager"
	types "github.com/gocasters/rankr/type"
)

var ErrGrantNotFound = errors.New("grant not found")

type Repository interface {
	Create(ctx context.Context, g Grant) (types.ID, error)
	Update(ctx context.Context, g Grant) error
	Get(ctx context.Context, id types.ID) (Grant, error)
	Delete(ctx context.Context, id types.ID) error
	List(ctx context.Context) ([]Grant, error)
}

type Validator interface {
	ValidateCreate(req CreateGrantRequest) error
	ValidateUpdate(req UpdateGrantRequest) error
	ValidateID(id types.ID) error
}

type Service struct {
	repository   Repository
	validator    Validator
	cacheManager cachemanager.CacheManager
}

func NewService(repo Repository, validator Validator, cache cachemanager.CacheManager, _ interface{}) Service {
	return Service{
		repository:   repo,
		validator:    validator,
		cacheManager: cache,
	}
}

func (s Service) CreateGrant(ctx context.Context, req CreateGrantRequest) (Grant, error) {
	if err := s.validator.ValidateCreate(req); err != nil {
		return Grant{}, err
	}

	now := time.Now().UTC()
	grant := Grant{
		Subject:   req.Subject,
		Object:    req.Object,
		Action:    req.Action,
		Field:     req.Field,
		CreatedAt: now,
		UpdatedAt: now,
	}

	id, err := s.repository.Create(ctx, grant)
	if err != nil {
		return Grant{}, err
	}

	grant.ID = id
	return grant, nil
}

func (s Service) UpdateGrant(ctx context.Context, req UpdateGrantRequest) (Grant, error) {
	if err := s.validator.ValidateUpdate(req); err != nil {
		return Grant{}, err
	}

	grant, err := s.repository.Get(ctx, types.ID(req.ID))
	if err != nil {
		return Grant{}, err
	}

	if req.Subject != "" {
		grant.Subject = req.Subject
	}
	if req.Object != "" {
		grant.Object = req.Object
	}
	if req.Action != "" {
		grant.Action = req.Action
	}
	if req.Field != nil {
		grant.Field = req.Field
	}
	grant.UpdatedAt = time.Now().UTC()

	if err := s.repository.Update(ctx, grant); err != nil {
		return Grant{}, err
	}

	return grant, nil
}

func (s Service) GetGrant(ctx context.Context, id types.ID) (Grant, error) {
	if err := s.validator.ValidateID(id); err != nil {
		return Grant{}, err
	}

	return s.repository.Get(ctx, id)
}

func (s Service) DeleteGrant(ctx context.Context, id types.ID) error {
	if err := s.validator.ValidateID(id); err != nil {
		return err
	}

	return s.repository.Delete(ctx, id)
}

func (s Service) ListGrants(ctx context.Context) ([]Grant, error) {
	return s.repository.List(ctx)
}

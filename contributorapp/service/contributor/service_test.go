package contributor

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/cachemanager"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/role"
	"github.com/gocasters/rankr/pkg/statuscode"
	types "github.com/gocasters/rankr/type"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// service_test.go — Issue #274 (TD-23), scenarios B1–B16 from test-plan-274.html.
//
// Covers the profile / create / update acceptance criteria with a hand-written
// testify mock of the Repository interface (no new dependencies). Every test
// locks the CURRENT behaviour: these are characterization tests, so a future
// refactor that changes what the service sends to the repository, what it
// returns to the HTTP handler, or how errors are typed must break here first.

func TestMain(m *testing.M) {
	// logger.L() panics when the logger is not initialised (V10) — the service
	// logs on every error path, so init once for the whole package.
	if err := logger.Init(logger.Config{Level: "error"}); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// --- mock Repository (6 methods, interface defined in service.go) ---

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetContributorByID(ctx context.Context, id types.ID) (*Contributor, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Contributor), args.Error(1)
}

func (m *mockRepository) GetContributorByGitHubUsername(ctx context.Context, username string) (*Contributor, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Contributor), args.Error(1)
}

func (m *mockRepository) CreateContributor(ctx context.Context, c Contributor) (*Contributor, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Contributor), args.Error(1)
}

func (m *mockRepository) UpdateProfileContributor(ctx context.Context, c Contributor) (*Contributor, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Contributor), args.Error(1)
}

func (m *mockRepository) UpdatePassword(ctx context.Context, id types.ID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *mockRepository) FindByVCSUsernames(ctx context.Context, provider VcsProvider, usernames []string) ([]*Contributor, error) {
	args := m.Called(ctx, provider, usernames)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Contributor), args.Error(1)
}

// newTestService wires the mock repo with a zero-value cache (the service
// never touches CacheManager in the tested paths) and the real validator.
func newTestService(repo *mockRepository) Service {
	return NewService(repo, cachemanager.CacheManager{}, NewValidator())
}

// --- profile (AC: "Tests for profile") ---

func TestGetProfile_Success(t *testing.T) {
	repo := new(mockRepository)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	contrib := &Contributor{
		ID:             7,
		GitHubID:       42,
		GitHubUsername: "alice01",
		DisplayName:    "Alice",
		ProfileImage:   "https://img.example/a.png",
		Bio:            "hello there",
		PrivacyMode:    PrivacyModeReal,
		CreatedAt:      now,
		Password:       "HASHED_SECRET", // must never leak into the response
		TwoFactor:      true,
		Role:           string(role.User),
	}
	repo.On("GetContributorByID", mock.Anything, types.ID(7)).Return(contrib, nil).Once()

	svc := newTestService(repo)
	resp, err := svc.GetProfile(context.Background(), 7)
	require.NoError(t, err)

	assert.Equal(t, contrib.ID, resp.ID)
	assert.Equal(t, contrib.GitHubID, resp.GitHubID)
	assert.Equal(t, contrib.GitHubUsername, resp.GitHubUsername)
	assert.Equal(t, contrib.DisplayName, resp.DisplayName)
	assert.Equal(t, contrib.ProfileImage, resp.ProfileImage)
	assert.Equal(t, contrib.Bio, resp.Bio)
	assert.Equal(t, contrib.PrivacyMode, resp.PrivacyMode)
	assert.Equal(t, contrib.CreatedAt, resp.CreatedAt)

	repo.AssertNumberOfCalls(t, "GetContributorByID", 1)
	repo.AssertExpectations(t)
}

// V8 lock: the HTTP handler type-asserts errmsg.ErrorResponse WITHOUT the ok
// form, so GetProfile must never return any other error type on failure.
func TestGetProfile_NotFound_ReturnsErrorResponse(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(999)).Return(nil, ErrNotFoundID).Once()

	svc := newTestService(repo)
	_, err := svc.GetProfile(context.Background(), 999)
	require.Error(t, err)

	eResp, ok := err.(errmsg.ErrorResponse)
	require.True(t, ok, "GetProfile must always return errmsg.ErrorResponse, got %T", err)
	assert.Equal(t, ErrNotFoundID.Error(), eResp.Message)
	require.Contains(t, eResp.Errors, "contributor_get_profile")
	repo.AssertExpectations(t)
}

// V9 lock: the profile response must never expose password or two-factor data.
func TestGetProfile_NeverExposesSecrets(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(1)).Return(&Contributor{
		ID:             1,
		GitHubUsername: "alice01",
		Password:       "SECRET_HASH",
		TwoFactor:      true,
		Role:           "admin",
		Email:          "alice@example.com",
	}, nil).Once()

	svc := newTestService(repo)
	resp, err := svc.GetProfile(context.Background(), 1)
	require.NoError(t, err)

	raw, marshalErr := json.Marshal(resp)
	require.NoError(t, marshalErr)
	assert.NotContains(t, string(raw), "password")
	assert.NotContains(t, string(raw), "two_factor")
	assert.NotContains(t, string(raw), "email")
	assert.NotContains(t, string(raw), "role")
}

func TestGetProfile_OptionalFieldsEmpty(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(2)).Return(&Contributor{
		ID:             2,
		GitHubUsername: "bob001",
		CreatedAt:      time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}, nil).Once()

	svc := newTestService(repo)
	resp, err := svc.GetProfile(context.Background(), 2)
	require.NoError(t, err)
	assert.Empty(t, resp.DisplayName)
	assert.Empty(t, resp.Bio)
	assert.Empty(t, resp.ProfileImage)

	raw, marshalErr := json.Marshal(resp)
	require.NoError(t, marshalErr)
	// omitempty: absent optionals must not appear in the JSON at all
	assert.NotContains(t, string(raw), "display_name")
	assert.NotContains(t, string(raw), "bio")
	assert.NotContains(t, string(raw), "profile_image")
}

// --- create (AC: "create/update") ---

func TestCreate_Success_HashesPassword(t *testing.T) {
	repo := new(mockRepository)
	var captured Contributor
	repo.On("CreateContributor", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { captured = args.Get(1).(Contributor) }).
		Return(&Contributor{ID: 42}, nil).Once()

	svc := newTestService(repo)
	req := CreateContributorRequest{
		GitHubID:       99,
		GitHubUsername: "carol01",
		Password:       "plain-password",
		DisplayName:    "Carol",
		ProfileImage:   "img.png",
		Bio:            "hi",
		PrivacyMode:    PrivacyModeAnonymous,
	}
	resp, err := svc.CreateContributor(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, types.ID(42), resp.ID)

	// the plain password must never reach the repository
	assert.NotEqual(t, req.Password, captured.Password)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(captured.Password), []byte(req.Password)),
		"stored password must be a bcrypt hash of the plain one")

	assert.Equal(t, string(role.User), captured.Role)
	assert.Equal(t, PrivacyModeAnonymous, captured.PrivacyMode, "privacy mode must pass through unchanged")
	assert.Equal(t, req.GitHubID, captured.GitHubID)
	assert.Equal(t, req.GitHubUsername, captured.GitHubUsername)
	assert.Equal(t, req.DisplayName, captured.DisplayName)
	assert.Equal(t, req.ProfileImage, captured.ProfileImage)
	assert.Equal(t, req.Bio, captured.Bio)
	assert.False(t, captured.CreatedAt.IsZero())
	assert.Equal(t, captured.CreatedAt, captured.UpdatedAt, "create stamps both timestamps with the same value")
	repo.AssertExpectations(t)
}

// V1 lock: the "default privacy to real" line in service.go runs AFTER
// validation, so an empty privacy mode is rejected before any repository call
// and that default line is dead code. Reordering them changes this test.
func TestCreate_EmptyPrivacy_RejectedBeforeRepo(t *testing.T) {
	repo := new(mockRepository)
	svc := newTestService(repo)

	req := validCreateReq()
	req.PrivacyMode = ""

	_, err := svc.CreateContributor(context.Background(), req)
	requireErrContains(t, err, ErrValidationRequired)
	repo.AssertNotCalled(t, "CreateContributor", mock.Anything, mock.Anything)
}

func TestCreate_InvalidInput_NoRepoCall(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CreateContributorRequest)
	}{
		{"short username", func(r *CreateContributorRequest) { r.GitHubUsername = "ab" }},
		{"short password", func(r *CreateContributorRequest) { r.Password = "12345" }},
		{"invalid privacy", func(r *CreateContributorRequest) { r.PrivacyMode = "hidden" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			svc := newTestService(repo)

			req := validCreateReq()
			tt.mutate(&req)

			_, err := svc.CreateContributor(context.Background(), req)
			require.Error(t, err)
			// validation fails before hashing and before any repository access
			repo.AssertNotCalled(t, "CreateContributor", mock.Anything, mock.Anything)
		})
	}
}

func TestCreate_RepoError_Propagates(t *testing.T) {
	dbErr := errors.New(`duplicate key value violates unique constraint "contributors_github_username_key"`)
	repo := new(mockRepository)
	repo.On("CreateContributor", mock.Anything, mock.Anything).Return(nil, dbErr).Once()

	svc := newTestService(repo)
	_, err := svc.CreateContributor(context.Background(), validCreateReq())
	require.Error(t, err)
	assert.ErrorIs(t, err, dbErr, "repository error must pass through unwrapped")

	// current contract: create propagates the RAW repository error (the HTTP
	// handler maps unknown errors to 500 by itself)
	_, isResp := err.(errmsg.ErrorResponse)
	assert.False(t, isResp, "create currently returns the raw repository error")
	repo.AssertExpectations(t)
}

func TestCreate_BcryptDifferentEachTime(t *testing.T) {
	var first, second Contributor
	repo := new(mockRepository)
	repo.On("CreateContributor", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			c := args.Get(1).(Contributor)
			if first.Password == "" {
				first = c
			} else {
				second = c
			}
		}).
		Return(&Contributor{ID: 1}, nil)

	svc := newTestService(repo)
	req := validCreateReq()

	_, err := svc.CreateContributor(context.Background(), req)
	require.NoError(t, err)
	_, err = svc.CreateContributor(context.Background(), req)
	require.NoError(t, err)

	require.NotEmpty(t, first.Password)
	require.NotEmpty(t, second.Password)
	assert.NotEqual(t, first.Password, second.Password, "bcrypt must salt each hash differently")
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(first.Password), []byte(req.Password)))
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(second.Password), []byte(req.Password)))
}

// --- update (AC: "create/update") ---

func TestUpdateProfile_Success(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(5)).
		Return(&Contributor{ID: 5, GitHubUsername: "alice01"}, nil).Once()

	created := time.Date(2025, 5, 5, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	stored := &Contributor{
		ID:             5,
		GitHubID:       77,
		GitHubUsername: "alice02",
		DisplayName:    "Alice 2",
		ProfileImage:   "new.png",
		Bio:            "new bio",
		PrivacyMode:    PrivacyModeAnonymous,
		CreatedAt:      created,
		UpdatedAt:      updated,
	}
	repo.On("UpdateProfileContributor", mock.Anything, mock.Anything).Return(stored, nil).Once()

	svc := newTestService(repo)
	resp, err := svc.UpdateProfile(context.Background(), UpdateProfileRequest{
		ID:             5,
		GitHubID:       77,
		GitHubUsername: "alice02",
		DisplayName:    "Alice 2",
		ProfileImage:   "new.png",
		Bio:            "new bio",
		PrivacyMode:    PrivacyModeAnonymous,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(5), resp.ID)
	assert.Equal(t, int64(77), resp.GitHubID)
	assert.Equal(t, "alice02", resp.GitHubUsername)
	assert.Equal(t, "Alice 2", resp.DisplayName)
	assert.Equal(t, PrivacyModeAnonymous, resp.PrivacyMode)
	assert.Equal(t, created, resp.CreatedAt)
	assert.Equal(t, updated, resp.UpdatedAt)

	// existence check first, then exactly one write
	repo.AssertNumberOfCalls(t, "GetContributorByID", 1)
	repo.AssertNumberOfCalls(t, "UpdateProfileContributor", 1)
	repo.AssertExpectations(t)
}

func TestUpdateProfile_NotFound_NoUpdate(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(404)).Return(nil, ErrNotFoundID).Once()

	svc := newTestService(repo)
	_, err := svc.UpdateProfile(context.Background(), UpdateProfileRequest{
		ID:             404,
		GitHubUsername: "ghost01",
		PrivacyMode:    PrivacyModeReal,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNotFoundID)
	repo.AssertNotCalled(t, "UpdateProfileContributor", mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestUpdateProfile_Invalid_NoRepoCall(t *testing.T) {
	repo := new(mockRepository)
	svc := newTestService(repo)

	req := validUpdateReq()
	req.GitHubUsername = "" // invalid

	_, err := svc.UpdateProfile(context.Background(), req)
	requireErrContains(t, err, ErrValidationRequired)
	repo.AssertNotCalled(t, "GetContributorByID", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "UpdateProfileContributor", mock.Anything, mock.Anything)
}

// V8 lock: unlike GetProfile, UpdateProfile propagates RAW repository errors.
// If someone later wraps them into errmsg.ErrorResponse, this test (and the
// HTTP handler's error mapping) must be revisited consciously.
func TestUpdateProfile_ErrorType_Raw(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(1)).Return(&Contributor{ID: 1}, nil).Once()
	dbErr := errors.New("connection refused")
	repo.On("UpdateProfileContributor", mock.Anything, mock.Anything).Return(nil, dbErr).Once()

	svc := newTestService(repo)
	_, err := svc.UpdateProfile(context.Background(), validUpdateReq())
	require.Error(t, err)
	assert.ErrorIs(t, err, dbErr)

	_, isResp := err.(errmsg.ErrorResponse)
	assert.False(t, isResp, "current contract: raw error, NOT errmsg.ErrorResponse")
}

func TestUpdateProfile_OnlyFieldsChanged(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(1)).Return(&Contributor{ID: 1}, nil).Once()

	var sent Contributor
	repo.On("UpdateProfileContributor", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { sent = args.Get(1).(Contributor) }).
		Return(&Contributor{ID: 1}, nil).Once()

	svc := newTestService(repo)
	_, err := svc.UpdateProfile(context.Background(), validUpdateReq())
	require.NoError(t, err)

	// requested profile fields are populated
	assert.Equal(t, int64(1), sent.ID)
	assert.Equal(t, "alice01", sent.GitHubUsername)
	assert.Equal(t, PrivacyModeReal, sent.PrivacyMode)

	// security-sensitive fields stay zero: they are NOT part of a profile
	// update. NB: the repository still writes email=$7 with this zero value —
	// see the F2 finding in the PR disclosure (repository layer, out of scope).
	assert.Empty(t, sent.Password)
	assert.Empty(t, sent.Role)
	assert.False(t, sent.IsVerified)
	assert.Empty(t, sent.Email)
}

// --- upsert (the bridge used by the CSV import path — AC: "CSV job parsing") ---

func TestUpsert_New_vs_Existing(t *testing.T) {
	t.Run("missing username creates with IsNew=true", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByGitHubUsername", mock.Anything, "ghost01").
			Return(nil, ErrNotFoundGithubUsername).Once()
		repo.On("CreateContributor", mock.Anything, mock.Anything).
			Return(&Contributor{ID: 11}, nil).Once()

		svc := newTestService(repo)
		resp, err := svc.Upsert(context.Background(), UpsertContributorRequest{
			GitHubID:       1,
			GitHubUsername: "ghost01",
			PrivacyMode:    PrivacyModeReal,
		})
		require.NoError(t, err)
		assert.Equal(t, types.ID(11), resp.ID)
		assert.True(t, resp.IsNew)
		repo.AssertNotCalled(t, "UpdateProfileContributor", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})

	t.Run("existing username updates and keeps its ID", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByGitHubUsername", mock.Anything, "alice01").
			Return(&Contributor{ID: 22, GitHubUsername: "alice01"}, nil).Once()
		repo.On("UpdateProfileContributor", mock.Anything, mock.Anything).
			Return(&Contributor{ID: 22}, nil).Once()

		svc := newTestService(repo)
		resp, err := svc.Upsert(context.Background(), UpsertContributorRequest{
			GitHubID:       1,
			GitHubUsername: "alice01",
			PrivacyMode:    PrivacyModeReal,
		})
		require.NoError(t, err)
		assert.Equal(t, types.ID(22), resp.ID)
		assert.False(t, resp.IsNew)
		repo.AssertNotCalled(t, "CreateContributor", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})
}

func TestUpsert_RepoErrorMapping(t *testing.T) {
	t.Run("get failure maps to unexpected ErrorResponse", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByGitHubUsername", mock.Anything, "alice01").
			Return(nil, errors.New("db down")).Once()

		svc := newTestService(repo)
		_, err := svc.Upsert(context.Background(), UpsertContributorRequest{GitHubUsername: "alice01"})
		require.Error(t, err)

		eResp, ok := err.(errmsg.ErrorResponse)
		require.True(t, ok, "upsert must return errmsg.ErrorResponse, got %T", err)
		assert.Equal(t, "failed to get contributor", eResp.Message)
		assert.Equal(t, statuscode.IntCodeUnExpected, eResp.InternalErrCode)
	})

	t.Run("create failure keeps db error in message", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByGitHubUsername", mock.Anything, "alice01").
			Return(nil, ErrNotFoundGithubUsername).Once()
		repo.On("CreateContributor", mock.Anything, mock.Anything).
			Return(nil, errors.New("boom")).Once()

		svc := newTestService(repo)
		_, err := svc.Upsert(context.Background(), UpsertContributorRequest{GitHubUsername: "alice01"})
		require.Error(t, err)

		eResp, ok := err.(errmsg.ErrorResponse)
		require.True(t, ok)
		assert.Contains(t, eResp.Message, "failed to create contributor")
		assert.Equal(t, statuscode.IntCodeUnExpected, eResp.InternalErrCode)
	})

	t.Run("update failure maps to update message", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByGitHubUsername", mock.Anything, "alice01").
			Return(&Contributor{ID: 3, GitHubUsername: "alice01"}, nil).Once()
		repo.On("UpdateProfileContributor", mock.Anything, mock.Anything).
			Return(nil, errors.New("boom")).Once()

		svc := newTestService(repo)
		_, err := svc.Upsert(context.Background(), UpsertContributorRequest{GitHubUsername: "alice01"})
		require.Error(t, err)

		eResp, ok := err.(errmsg.ErrorResponse)
		require.True(t, ok)
		assert.Equal(t, "failed to update contributor", eResp.Message)
		assert.Equal(t, statuscode.IntCodeUnExpected, eResp.InternalErrCode)
	})
}

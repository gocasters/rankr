package contributor

import (
	"context"
	"errors"
	"testing"

	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/statuscode"
	types "github.com/gocasters/rankr/type"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// password_test.go — UpdatePassword, VerifyPassword and passwordMatches.
//
// These paths are security-critical: UpdatePassword is called by the HTTP API
// and VerifyPassword by authapp over gRPC (login). The tests protect:
//   - the old password is checked before anything is written;
//   - the new password is stored only as a bcrypt hash, never plain;
//   - a wrong password is a normal "Valid=false" answer, not an error;
//   - lookup by ID wins over lookup by username.
// Test passwords below are obviously fake and are never logged.

// testHash returns a bcrypt hash at MinCost (fast; CompareHashAndPassword
// accepts any cost, so this does not weaken what the tests prove).
func testHash(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	require.NoError(t, err)
	return string(h)
}

// --- UpdatePassword ---

func TestUpdatePassword_Success_StoresBcryptOfNewPassword(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(5)).
		Return(&Contributor{ID: 5, Password: testHash(t, "old-fake-pw")}, nil).Once()

	var storedHash string
	repo.On("UpdatePassword", mock.Anything, types.ID(5), mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) { storedHash = args.String(2) }).
		Return(nil).Once()

	resp, err := newTestService(repo).UpdatePassword(context.Background(), UpdatePasswordRequest{
		ID: 5, OldPassword: "old-fake-pw", NewPassword: "new-fake-pw",
	})
	require.NoError(t, err)
	assert.True(t, resp.Success)

	assert.NotEqual(t, "new-fake-pw", storedHash, "SECURITY: the new password must never be stored in plain text")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("new-fake-pw")),
		"the stored value must be a bcrypt hash of the NEW password")
	assert.Error(t, bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("old-fake-pw")),
		"the stored value must not still match the OLD password")
	repo.AssertExpectations(t)
}

func TestUpdatePassword_WrongOldPassword_Unauthorized_NoWrite(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(5)).
		Return(&Contributor{ID: 5, Password: testHash(t, "real-fake-pw")}, nil).Once()

	resp, err := newTestService(repo).UpdatePassword(context.Background(), UpdatePasswordRequest{
		ID: 5, OldPassword: "guessed-pw", NewPassword: "new-fake-pw",
	})
	require.Error(t, err)
	assert.False(t, resp.Success)

	eResp, ok := err.(errmsg.ErrorResponse)
	require.True(t, ok, "wrong old password must be errmsg.ErrorResponse (handler maps it to 401), got %T", err)
	assert.Equal(t, "old password is incorrect", eResp.Message)
	assert.Equal(t, statuscode.IntCodeNotAuthorize, eResp.InternalErrCode)

	repo.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	repo.AssertExpectations(t)
}

func TestUpdatePassword_Invalid_NoRepoCall(t *testing.T) {
	tests := []struct {
		name string
		req  UpdatePasswordRequest
	}{
		{"zero id", UpdatePasswordRequest{ID: 0, OldPassword: "old-fake-pw", NewPassword: "new-fake-pw"}},
		{"short new password", UpdatePasswordRequest{ID: 5, OldPassword: "old-fake-pw", NewPassword: "12345"}},
		{"empty old password", UpdatePasswordRequest{ID: 5, OldPassword: "", NewPassword: "new-fake-pw"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			_, err := newTestService(repo).UpdatePassword(context.Background(), tt.req)
			require.Error(t, err)
			repo.AssertNotCalled(t, "GetContributorByID", mock.Anything, mock.Anything)
			repo.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestUpdatePassword_RepoErrors_Propagate(t *testing.T) {
	t.Run("lookup failure: no write", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByID", mock.Anything, types.ID(5)).Return(nil, ErrNotFoundID).Once()

		resp, err := newTestService(repo).UpdatePassword(context.Background(), UpdatePasswordRequest{
			ID: 5, OldPassword: "old-fake-pw", NewPassword: "new-fake-pw",
		})
		assert.ErrorIs(t, err, ErrNotFoundID)
		assert.False(t, resp.Success)
		repo.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("write failure: Success=false", func(t *testing.T) {
		dbErr := errors.New("connection refused")
		repo := new(mockRepository)
		repo.On("GetContributorByID", mock.Anything, types.ID(5)).
			Return(&Contributor{ID: 5, Password: testHash(t, "old-fake-pw")}, nil).Once()
		repo.On("UpdatePassword", mock.Anything, types.ID(5), mock.Anything).Return(dbErr).Once()

		resp, err := newTestService(repo).UpdatePassword(context.Background(), UpdatePasswordRequest{
			ID: 5, OldPassword: "old-fake-pw", NewPassword: "new-fake-pw",
		})
		assert.ErrorIs(t, err, dbErr)
		assert.False(t, resp.Success, "a failed write must never report success")
		repo.AssertExpectations(t)
	})
}

// --- VerifyPassword (gRPC contract with authapp) ---

func TestVerifyPassword_RequiredInputs_NoRepoCall(t *testing.T) {
	tests := []struct {
		name    string
		req     VerifyPasswordRequest
		wantErr string
	}{
		{"empty password", VerifyPasswordRequest{ID: 1, Password: ""}, "password is required"},
		{"no id and no username", VerifyPasswordRequest{Password: "any-fake-pw"}, "id or github_username is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockRepository)
			_, err := newTestService(repo).VerifyPassword(context.Background(), tt.req)
			requireErrContains(t, err, tt.wantErr)
			repo.AssertNotCalled(t, "GetContributorByID", mock.Anything, mock.Anything)
			repo.AssertNotCalled(t, "GetContributorByGitHubUsername", mock.Anything, mock.Anything)
		})
	}
}

func TestVerifyPassword_Lookup(t *testing.T) {
	hash := testHash(t, "login-fake-pw")

	t.Run("by id", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByID", mock.Anything, types.ID(9)).
			Return(&Contributor{ID: 9, Role: "admin", Password: hash}, nil).Once()

		resp, err := newTestService(repo).VerifyPassword(context.Background(),
			VerifyPasswordRequest{ID: 9, Password: "login-fake-pw"})
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, types.ID(9), resp.ID)
		assert.Equal(t, "admin", resp.Role, "authapp builds the token role from this field")
		repo.AssertNotCalled(t, "GetContributorByGitHubUsername", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})

	t.Run("by username when id is zero", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByGitHubUsername", mock.Anything, "alice01").
			Return(&Contributor{ID: 3, Role: "user", Password: hash}, nil).Once()

		resp, err := newTestService(repo).VerifyPassword(context.Background(),
			VerifyPasswordRequest{GitHubUsername: "alice01", Password: "login-fake-pw"})
		require.NoError(t, err)
		assert.True(t, resp.Valid)
		assert.Equal(t, types.ID(3), resp.ID)
		repo.AssertNotCalled(t, "GetContributorByID", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})

	t.Run("id wins when both are set", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("GetContributorByID", mock.Anything, types.ID(9)).
			Return(&Contributor{ID: 9, Password: hash}, nil).Once()

		_, err := newTestService(repo).VerifyPassword(context.Background(),
			VerifyPasswordRequest{ID: 9, GitHubUsername: "someone-else", Password: "login-fake-pw"})
		require.NoError(t, err)
		repo.AssertNotCalled(t, "GetContributorByGitHubUsername", mock.Anything, mock.Anything)
		repo.AssertExpectations(t)
	})
}

func TestVerifyPassword_WrongPassword_ValidFalseNoError(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(9)).
		Return(&Contributor{ID: 9, Password: testHash(t, "login-fake-pw")}, nil).Once()

	resp, err := newTestService(repo).VerifyPassword(context.Background(),
		VerifyPasswordRequest{ID: 9, Password: "wrong-fake-pw"})
	require.NoError(t, err, "a wrong password is a normal answer, not an error")
	assert.False(t, resp.Valid, "SECURITY: a wrong password must never verify")
}

func TestVerifyPassword_RepoError_Propagates(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByGitHubUsername", mock.Anything, "ghost01").
		Return(nil, ErrNotFoundGithubUsername).Once()

	resp, err := newTestService(repo).VerifyPassword(context.Background(),
		VerifyPasswordRequest{GitHubUsername: "ghost01", Password: "any-fake-pw"})
	assert.ErrorIs(t, err, ErrNotFoundGithubUsername)
	assert.False(t, resp.Valid)
}

// --- passwordMatches ---

func TestPasswordMatches_Bcrypt(t *testing.T) {
	hash := testHash(t, "match-fake-pw")

	tests := []struct {
		name     string
		stored   string
		provided string
		want     bool
	}{
		{"correct password matches its hash", hash, "match-fake-pw", true},
		{"wrong password does not match", hash, "other-fake-pw", false},
		{"empty input does not match a hash", hash, "", false},
		{"case differs: no match", hash, "MATCH-FAKE-PW", false},
		{"empty stored value (CSV-imported user, no password) never matches", "", "any-fake-pw", false},
		{"hash string itself is not accepted as the password", hash, hash, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, passwordMatches(tt.stored, tt.provided))
		})
	}
}

// KNOWN BUG #346 (bugPlaintextFallback): when the stored value is not a bcrypt
// hash, passwordMatches falls back to `stored == provided`, so a plaintext
// (legacy or corrupted) password column logs the user in. Expected: only a
// bcrypt hash can ever match. If product decides legacy plaintext rows must
// keep working, this becomes a characterization test with that decision
// recorded in the issue.
func TestKnownBug_PasswordMatches_PlaintextFallback(t *testing.T) {
	repo := new(mockRepository)
	repo.On("GetContributorByID", mock.Anything, types.ID(9)).
		Return(&Contributor{ID: 9, Password: "plain-fake-pw"}, nil).Once()

	resp, err := newTestService(repo).VerifyPassword(context.Background(),
		VerifyPasswordRequest{ID: 9, Password: "plain-fake-pw"})
	require.NoError(t, err)

	knownBug(t, bugPlaintextFallback, func(a *assert.Assertions) {
		a.False(passwordMatches("plain-fake-pw", "plain-fake-pw"),
			"SECURITY: a non-bcrypt stored value must never match")
		a.False(resp.Valid,
			"SECURITY: VerifyPassword (authapp login) must reject a plaintext stored password")
	})
}

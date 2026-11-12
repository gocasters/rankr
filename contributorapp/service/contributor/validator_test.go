package contributor

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validator_test.go — Issue #274 (TD-23), scenarios A1–A14 from test-plan-274.html.
//
// The validator is stateless, so these tests exercise it directly with raw
// inputs (no mocks). They lock the current validation contract of
// ValidateCreateContributorRequest / ValidateUpdateProfileRequest /
// ValidateUpdatePasswordRequest / ValidateUpsertContributorRequest so a future
// refactor cannot silently widen or narrow what the API accepts.

func validCreateReq() CreateContributorRequest {
	return CreateContributorRequest{
		GitHubID:       1,
		GitHubUsername: "alice01",
		Password:       "secret123",
		PrivacyMode:    PrivacyModeReal,
	}
}

func validUpdateReq() UpdateProfileRequest {
	return UpdateProfileRequest{
		ID:             1,
		GitHubID:       1,
		GitHubUsername: "alice01",
		PrivacyMode:    PrivacyModeReal,
	}
}

func validPasswordReq() UpdatePasswordRequest {
	return UpdatePasswordRequest{
		ID:          1,
		OldPassword: "oldpass123",
		NewPassword: "newpass123",
	}
}

func requireErrContains(t *testing.T, err error, want string) {
	t.Helper()
	require.Error(t, err, "expected validation error containing %q", want)
	assert.Contains(t, err.Error(), want)
}

// A1 (P0): privacy mode on create — only exact "real"/"anonymous" pass.
func TestValidateCreate_PrivacyMode(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name    string
		privacy PrivacyMode
		wantErr string // "" means valid
	}{
		{"real accepted", PrivacyModeReal, ""},
		{"anonymous accepted", PrivacyModeAnonymous, ""},
		{"empty rejected as required", "", ErrValidationRequired},
		{"uppercase REAL rejected (case-sensitive)", "REAL", ErrValidationEnumPrivacy},
		{"arbitrary yes rejected", "yes", ErrValidationEnumPrivacy},
		{"numeric 1 rejected", "1", ErrValidationEnumPrivacy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validCreateReq()
			req.PrivacyMode = tt.privacy

			err := v.ValidateCreateContributorRequest(ctx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			requireErrContains(t, err, tt.wantErr)
		})
	}
}

// A2 (P0): privacy mode on update — must be equivalent to create (A1).
func TestValidateUpdate_PrivacyMode(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name    string
		privacy PrivacyMode
		wantErr string
	}{
		{"real accepted", PrivacyModeReal, ""},
		{"anonymous accepted", PrivacyModeAnonymous, ""},
		{"empty rejected as required", "", ErrValidationRequired},
		{"uppercase REAL rejected", "REAL", ErrValidationEnumPrivacy},
		{"arbitrary yes rejected", "yes", ErrValidationEnumPrivacy},
		{"numeric 1 rejected", "1", ErrValidationEnumPrivacy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validUpdateReq()
			req.PrivacyMode = tt.privacy

			err := v.ValidateUpdateProfileRequest(ctx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			requireErrContains(t, err, tt.wantErr)
		})
	}
}

// A3 (P1): username length boundaries 3..100 on create.
func TestValidateCreate_UsernameBoundary(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name     string
		username string
		wantErr  string
	}{
		{"2 chars rejected", "ab", ErrValidationLength3To100},
		{"3 chars accepted", "abc", ""},
		{"100 chars accepted", strings.Repeat("a", 100), ""},
		{"101 chars rejected", strings.Repeat("a", 101), ErrValidationLength3To100},
		{"empty rejected as required", "", ErrValidationRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validCreateReq()
			req.GitHubUsername = tt.username

			err := v.ValidateCreateContributorRequest(ctx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			requireErrContains(t, err, tt.wantErr)
		})
	}
}

// A4 (P1): password length boundaries 6..72 (72 = bcrypt limit).
func TestValidateCreate_PasswordBoundary(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name     string
		password string
		wantErr  string
	}{
		{"5 chars rejected", "12345", ErrValidationLength6To72},
		{"6 chars accepted", "123456", ""},
		{"72 chars accepted", strings.Repeat("p", 72), ""},
		{"73 chars rejected", strings.Repeat("p", 73), ErrValidationLength6To72},
		{"empty rejected as required", "", ErrValidationRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validCreateReq()
			req.Password = tt.password

			err := v.ValidateCreateContributorRequest(ctx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			requireErrContains(t, err, tt.wantErr)
		})
	}
}

// A5 (P1): GitHubID — 0 means "unknown" (rule skipped, valid);
// any non-zero value must be >= 1.
func TestValidateCreate_GitHubID(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name    string
		gid     int64
		wantErr string
	}{
		{"zero means unknown and is accepted", 0, ""},
		{"1 accepted", 1, ""},
		{"negative rejected", -5, ErrValidationPositive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validCreateReq()
			req.GitHubID = tt.gid

			err := v.ValidateCreateContributorRequest(ctx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			requireErrContains(t, err, tt.wantErr)
		})
	}
}

// A6 (P1): optional text fields — bounded when set, unbounded-free when empty.
func TestValidateCreate_OptionalLengths(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name    string
		mutate  func(*CreateContributorRequest)
		wantErr string // "" = valid; otherwise substring of error
	}{
		{"display_name 100 accepted", func(r *CreateContributorRequest) {
			r.DisplayName = strings.Repeat("d", 100)
		}, ""},
		{"display_name 101 rejected with shared length message", func(r *CreateContributorRequest) {
			r.DisplayName = strings.Repeat("d", 101)
		}, ErrValidationLength3To100},
		{"bio 500 accepted", func(r *CreateContributorRequest) {
			r.Bio = strings.Repeat("b", 500)
		}, ""},
		{"bio 501 rejected", func(r *CreateContributorRequest) {
			r.Bio = strings.Repeat("b", 501)
		}, "bio"},
		{"profile_image 255 accepted", func(r *CreateContributorRequest) {
			r.ProfileImage = strings.Repeat("i", 255)
		}, ""},
		{"profile_image 256 rejected", func(r *CreateContributorRequest) {
			r.ProfileImage = strings.Repeat("i", 256)
		}, "profile_image"},
		{"all empty accepted", func(r *CreateContributorRequest) {
			r.DisplayName, r.Bio, r.ProfileImage = "", "", ""
		}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validCreateReq()
			tt.mutate(&req)

			err := v.ValidateCreateContributorRequest(ctx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			requireErrContains(t, err, tt.wantErr)
		})
	}
}

// A7 (P0): update requires a non-zero ID (Required first, then checkID).
func TestValidateUpdate_IDRequired(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	t.Run("zero ID rejected as required", func(t *testing.T) {
		req := validUpdateReq()
		req.ID = 0
		requireErrContains(t, v.ValidateUpdateProfileRequest(ctx, req), ErrValidationRequired)
	})

	t.Run("non-zero ID accepted", func(t *testing.T) {
		req := validUpdateReq()
		req.ID = 7
		require.NoError(t, v.ValidateUpdateProfileRequest(ctx, req))
	})
}

// A8 (P1): update username required + same length window as create.
func TestValidateUpdate_UsernameRequired(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name     string
		username string
		wantErr  string
	}{
		{"empty rejected as required", "", ErrValidationRequired},
		{"2 chars rejected", "ab", ErrValidationLength3To100},
		{"3 chars accepted", "abc", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validUpdateReq()
			req.GitHubUsername = tt.username

			err := v.ValidateUpdateProfileRequest(ctx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			requireErrContains(t, err, tt.wantErr)
		})
	}
}

// A9 (P1): errors from different fields accumulate in one error value
// (ozzo validates every field, stopping at the first failing rule per field).
func TestValidateUpdate_AllErrorMessages(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	req := validUpdateReq()
	req.ID = 0               // -> "is required"
	req.GitHubUsername = "x" // -> length 3..100
	req.PrivacyMode = "bad"  // -> enum

	err := v.ValidateUpdateProfileRequest(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrValidationRequired)
	assert.Contains(t, err.Error(), ErrValidationLength3To100)
	assert.Contains(t, err.Error(), ErrValidationEnumPrivacy)
}

// A10 (P2): exported error-message constants are the literal strings the API
// emits — changing a constant changes the wire contract and must break tests.
func TestValidateCreate_ErrorMessageConstants(t *testing.T) {
	assert.Equal(t, "is required", ErrValidationRequired)
	assert.Equal(t, "must be 'real' or 'anonymous'", ErrValidationEnumPrivacy)
	assert.Equal(t, "must be 'public' or 'private'", ErrValidationPositive)
	assert.Equal(t, "must be between 3 and 100 characters", ErrValidationLength3To100)
	assert.Equal(t, "must be between 6 and 72 characters", ErrValidationLength6To72)

	// and the validator really emits them
	v := NewValidator()
	ctx := context.Background()

	req := validCreateReq()
	req.GitHubUsername = ""
	requireErrContains(t, v.ValidateCreateContributorRequest(ctx, req), ErrValidationRequired)

	req = validCreateReq()
	req.PrivacyMode = "hidden"
	requireErrContains(t, v.ValidateCreateContributorRequest(ctx, req), ErrValidationEnumPrivacy)
}

// A11 (P1): update-password boundaries — both old and new must be 6..72.
func TestValidatePassword_OldNewBoundary(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	tests := []struct {
		name    string
		mutate  func(*UpdatePasswordRequest)
		wantErr string
	}{
		{"old 5 rejected", func(r *UpdatePasswordRequest) { r.OldPassword = "12345" }, ErrValidationLength6To72},
		{"old 73 rejected", func(r *UpdatePasswordRequest) { r.OldPassword = strings.Repeat("o", 73) }, ErrValidationLength6To72},
		{"new 5 rejected", func(r *UpdatePasswordRequest) { r.NewPassword = "12345" }, ErrValidationLength6To72},
		{"new 73 rejected", func(r *UpdatePasswordRequest) { r.NewPassword = strings.Repeat("n", 73) }, ErrValidationLength6To72},
		{"old empty rejected", func(r *UpdatePasswordRequest) { r.OldPassword = "" }, ErrValidationRequired},
		{"new empty rejected", func(r *UpdatePasswordRequest) { r.NewPassword = "" }, ErrValidationRequired},
		{"6 and 72 accepted", func(r *UpdatePasswordRequest) {
			r.OldPassword = strings.Repeat("o", 6)
			r.NewPassword = strings.Repeat("n", 72)
		}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validPasswordReq()
			tt.mutate(&req)

			err := v.ValidateUpdatePasswordRequest(ctx, req)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			requireErrContains(t, err, tt.wantErr)
		})
	}
}

// A12 (P0): upsert (the CSV import path) validates ONLY github_username.
// This deliberately locks the V2 gap: arbitrary privacy_mode values are
// accepted here — see Q1 decision (document, fix in a separate issue).
func TestValidateUpsert_Minimal(t *testing.T) {
	v := NewValidator()

	t.Run("empty username rejected", func(t *testing.T) {
		err := v.ValidateUpsertContributorRequest(UpsertContributorRequest{
			GitHubUsername: "",
		})
		require.Error(t, err)
		assert.Equal(t, "invalid request", err.Error())
	})

	t.Run("arbitrary privacy accepted (V2 gap, documented)", func(t *testing.T) {
		err := v.ValidateUpsertContributorRequest(UpsertContributorRequest{
			GitHubID:       1,
			GitHubUsername: "alice01",
			PrivacyMode:    "TOTALLY_INVALID",
			Email:          "not-an-email",
		})
		require.NoError(t, err, "upsert must keep accepting any privacy/email (Q1-b decision)")
	})

	t.Run("empty privacy accepted (upsert is intentionally minimal)", func(t *testing.T) {
		err := v.ValidateUpsertContributorRequest(UpsertContributorRequest{
			GitHubUsername: "alice01",
			PrivacyMode:    "",
		})
		require.NoError(t, err)
	})
}

// A13 (P2): UpdateProfileRequest must never grow a password field
// (compile-time DTO shape — enforced via reflection at test time).
func TestValidateUpdate_NoPasswordField(t *testing.T) {
	typ := reflect.TypeOf(UpdateProfileRequest{})
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		assert.NotContains(t, strings.ToLower(name), "password",
			"UpdateProfileRequest.%s — profile update must not carry password data", name)
	}
}

// A14 (P0): privacy enum literals are part of the DB contract —
// these exact strings are stored in contributors.privacy_mode.
func TestPrivacyModeConstants(t *testing.T) {
	assert.Equal(t, "real", string(PrivacyModeReal))
	assert.Equal(t, "anonymous", string(PrivacyModeAnonymous))
	assert.NotEqual(t, PrivacyModeReal, PrivacyModeAnonymous)
}

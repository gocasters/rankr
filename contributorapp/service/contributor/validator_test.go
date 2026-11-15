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
		// the message text for this case is wrong — see TestKnownBug_ValidationMessages
		{"negative rejected", -5, "github_id"},
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
		// rejected, but with a misleading message — see TestKnownBug_ValidationMessages
		{"display_name 101 rejected", func(r *CreateContributorRequest) {
			r.DisplayName = strings.Repeat("d", 101)
		}, "display_name"},
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

// A12 (P0): upsert (the CSV import path) requires github_username.
// What it should ALSO reject is in TestKnownBug_ValidateUpsert_PrivacyMode.
func TestValidateUpsert_Minimal(t *testing.T) {
	v := NewValidator()

	t.Run("empty username rejected", func(t *testing.T) {
		err := v.ValidateUpsertContributorRequest(UpsertContributorRequest{
			GitHubUsername: "",
		})
		require.Error(t, err)
		assert.Equal(t, "invalid request", err.Error())
	})

	t.Run("valid row accepted", func(t *testing.T) {
		err := v.ValidateUpsertContributorRequest(UpsertContributorRequest{
			GitHubID:       1,
			GitHubUsername: "alice01",
			PrivacyMode:    PrivacyModeReal,
			Email:          "alice@example.com",
		})
		require.NoError(t, err)
	})
}

// KNOWN BUG #345 (bugImportValidation): ValidateUpsertContributorRequest checks only
// github_username, so the CSV/XLSX import path accepts any privacy_mode. The
// DB column is privacy_mode_enum, so such a row later fails as an UNEXPECTED
// database error instead of a per-row validation error. Expected: the import
// path enforces the same enum as create/update (see A1/A2).
func TestKnownBug_ValidateUpsert_PrivacyMode(t *testing.T) {
	v := NewValidator()
	invalid := []PrivacyMode{"TOTALLY_INVALID", "REAL", "yes"}

	errs := make(map[PrivacyMode]error, len(invalid))
	for _, p := range invalid {
		errs[p] = v.ValidateUpsertContributorRequest(UpsertContributorRequest{
			GitHubID: 1, GitHubUsername: "alice01", PrivacyMode: p,
		})
	}

	knownBug(t, bugImportValidation, func(a *assert.Assertions) {
		for p, err := range errs {
			a.Error(err, "upsert must reject privacy_mode %q exactly like create/update", p)
		}
	})
}

// KNOWN BUG #348 (bugValidationMessages): two validation messages describe the
// wrong rule — github_id < 1 reports "must be 'public' or 'private'", and a
// too-long display_name reports "between 3 and 100 characters" although 0..2
// characters are accepted. Expected: each message describes its own rule.
func TestKnownBug_ValidationMessages(t *testing.T) {
	v := NewValidator()
	ctx := context.Background()

	negID := validCreateReq()
	negID.GitHubID = -5
	errID := v.ValidateCreateContributorRequest(ctx, negID)
	require.Error(t, errID)

	longName := validCreateReq()
	longName.DisplayName = strings.Repeat("d", 101)
	errName := v.ValidateCreateContributorRequest(ctx, longName)
	require.Error(t, errName)

	knownBug(t, bugValidationMessages, func(a *assert.Assertions) {
		a.NotContains(errID.Error(), "public",
			"github_id error must describe the positive-number rule, not a privacy/visibility enum")
		a.NotContains(errName.Error(), "between 3 and",
			"display_name accepts 0..100 characters; the message must not claim a minimum of 3")
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

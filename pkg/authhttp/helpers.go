package authhttp

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	types "github.com/gocasters/rankr/type"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	xRefreshTokenHeader = "X-Refresh-Token"
	refreshTokenHeader  = "Refresh-Token"
	refreshTokenCookie  = "refresh_token"
)

func EncodeUserInfo(userID string) (string, error) {
	parsedID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(types.UserClaim{ID: types.ID(parsedID)})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

func ExtractBearerToken(r *http.Request) string {
	authz := r.Header.Get(authorizationHeader)
	if len(authz) > len(bearerPrefix) && strings.HasPrefix(authz, bearerPrefix) {
		return authz[len(bearerPrefix):]
	}
	return ""
}

func ExtractRefreshToken(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get(xRefreshTokenHeader)); token != "" {
		return token
	}
	if token := strings.TrimSpace(r.Header.Get(refreshTokenHeader)); token != "" {
		return token
	}
	if cookie, err := r.Cookie(refreshTokenCookie); err == nil {
		if token := strings.TrimSpace(cookie.Value); token != "" {
			return token
		}
	}
	return ""
}

package tokenservice

import "time"

type AuthService struct {
	accessManager  *JWTManager
	refreshManager *JWTManager
}

func NewAuthService(secret string, accessDuration, refreshDuration time.Duration) *AuthService {
	return &AuthService{
		accessManager:  NewJWTManager(secret, accessDuration),
		refreshManager: NewJWTManager(secret, refreshDuration),
	}
}

func (s *AuthService) IssueToken(userID, role string, access []string) (string, error) {
	return s.accessManager.Generate(userID, role, access)
}

func (s *AuthService) IssueTokens(userID, role string, access []string) (string, string, error) {
	accessToken, err := s.accessManager.Generate(userID, role, access)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.refreshManager.Generate(userID, role, access)
	if err != nil {
		return "", "", err
	}
	return accessToken, refresh, nil
}

func (s *AuthService) VerifyToken(token string) (*UserClaims, error) {
	return s.accessManager.Verify(token)
}

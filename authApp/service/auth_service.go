package service

import (
    "github.com/gocasters/rankr/authApp/auth"
)

type AuthService struct {
    jwtManager *auth.JWTManager
}

func NewAuthService(jwtManager *auth.JWTManager) *AuthService {
    return &AuthService{jwtManager: jwtManager}
}

func (s *AuthService) IssueToken(userID, role string) (string, error) {
    return s.jwtManager.Generate(userID, role)
}

func (s *AuthService) VerifyToken(token string) (*auth.UserClaims, error) {
    return s.jwtManager.Verify(token)
}


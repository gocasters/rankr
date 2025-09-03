package service

// import (
//     "github.com/gocasters/rankr/authapp/auth"
// )

type AuthService struct {
    jwtManager *JWTManager
}

func NewAuthService(jwtManager *JWTManager) *AuthService {
    return &AuthService{jwtManager: jwtManager}
}

func (s *AuthService) IssueToken(userID, role string) (string, error) {
    return s.jwtManager.Generate(userID, role)
}

func (s *AuthService) VerifyToken(token string) (*UserClaims, error) {
    return s.jwtManager.Verify(token)
}


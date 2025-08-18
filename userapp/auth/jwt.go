package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
    secretKey     string
    tokenDuration time.Duration
}

type UserClaims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func NewJWTManager(secretKey string, duration time.Duration) *JWTManager {
    return &JWTManager{secretKey, duration}
}

func (j *JWTManager) Generate(userID, role string) (string, error) {
    claims := UserClaims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(j.secretKey))
}

func (j *JWTManager) Verify(accessToken string) (*UserClaims, error) {
    token, err := jwt.ParseWithClaims(
        accessToken,
        &UserClaims{},
        func(t *jwt.Token) (interface{}, error) {
            // Verify signing method
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
            }
            return []byte(j.secretKey), nil
        },
    )
    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    claims, ok := token.Claims.(*UserClaims)
    if !ok {
        return nil, fmt.Errorf("invalid token claims")
    }
    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    return claims, nil
}

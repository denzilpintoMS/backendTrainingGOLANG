package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenLifetime = time.Hour
const issuer = "books-api"

type Tokens struct{ secret []byte }

func New(secret string) (*Tokens, error) {
	if secret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	return &Tokens{secret: []byte(secret)}, nil
}

func (t *Tokens) Issue(userID int) (string, error) {
	if userID <= 0 {
		return "", errors.New("invalid user ID")
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(TokenLifetime)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
}

func (t *Tokens) Verify(raw string) (int, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return t.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired(), jwt.WithIssuer(issuer), jwt.WithIssuedAt())
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(claims.Subject)
	if err != nil || id <= 0 || !token.Valid {
		return 0, errors.New("invalid token subject")
	}
	return id, nil
}

package utils

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type (
	JwtClaim struct {
		Username string `json:"username"`
		UUID     string `json:"uuid"`
		jwt.RegisteredClaims
	}

	Tokens struct {
		Username     string           `json:"username"`
		UUID         string           `json:"uuid"`
		Token        string           `json:"token"`
		RefreshToken string           `json:"refresh_token"`
		ExpiresAt    *jwt.NumericDate `json:"expires_at"`
	}
)

func GenerateJWT(uuid, username string) (*Tokens, error) {
	jwtKey := []byte(cfg.Server.JwtKey)

	generateToken := func(expiration time.Duration, uuid, username string) (string, error) {
		claims := registerToken(expiration, uuid, username)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString(jwtKey)
	}

	token, err := generateToken(15*time.Minute, uuid, username)
	if err != nil {
		log.Printf("Error signing access token: %s", err)
		return nil, err
	}

	refreshToken, err := generateToken(time.Hour, uuid, username)
	if err != nil {
		log.Printf("Error signing refresh token: %s", err)
		return nil, err
	}

	res := &Tokens{
		UUID:         uuid,
		Username:     username,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
	}

	return res, nil
}

func VerifyJWT(token, uuid string) error {
	if token == "" || uuid == "" {
		return errors.New("empty token or uuid")
	}

	jwtKey := []byte(cfg.Server.JwtKey)

	keyFunc := func(token *jwt.Token) (any, error) {
		return []byte(jwtKey), nil
	}

	t, err := jwt.ParseWithClaims(token, &JwtClaim{}, keyFunc)
	claims, _ := t.Claims.(*JwtClaim)

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return errors.New("token expired")
	}

	if claims.UUID != uuid {
		return errors.New("unauthorized - UUID mismatch")
	}

	if err != nil {
		return errors.New("invalid token or claims")
	}

	return nil
}

func RefreshJWT(uuid, username string) (string, error) {
	jwtKey := []byte(cfg.Server.JwtKey)

	refreshExpiration := time.Hour

	claims := registerToken(refreshExpiration, uuid, username)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	refreshTokenString, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return refreshTokenString, nil
}

func HandlePassword(action, password, hashedPassword string) (string, error) {
	switch action {
	case "hash":
		bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		if err != nil {
			return "", fmt.Errorf("failed to hash password: %w", err)
		}
		return string(bytes), nil

	case "check":
		err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			return "", fmt.Errorf("invalid password: %w", err)
		}
		return "", nil

	default:
		return "", errors.New("invalid action, must be 'hash' or 'check'")
	}
}

func registerToken(duration time.Duration, uuid, username string) (claims *JwtClaim) {
	claims = &JwtClaim{
		UUID:     uuid,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	return
}

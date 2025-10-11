package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"sheng-go-backend/config"
	"sheng-go-backend/pkg/entity/model"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"golang.org/x/crypto/argon2"
)

type key string

const (
	AccessTokenKey  key = "AuthToken"
	RefreshTokenKey key = "RefreshToken"
)

type CustomClaims struct {
	UserId string `json:"user_id"`
	jwt.RegisteredClaims
}

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// Default Argon2 parameters
var DefaultArgon2Params = Argon2Params{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

// GenerateAccessToken generates a new access token
func GenerateAccessToken(userId string) (string, error) {
	config.ReadConfig(config.ReadConfigOption{})

	// Define the expiration time for the access token (e.g., 15 minutes)
	expirationTime := time.Now().Add(15 * time.Minute)

	// Create the claims
	claims := CustomClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    config.C.AppName,
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(config.C.JwtTokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token
func GenerateRefreshToken(userId string) (string, error) {
	// Define the expiration time for the refresh token (e.g., 7 days)
	expirationTime := time.Now().Add(7 * 24 * time.Hour)

	// Create the claims
	claims := CustomClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    config.C.RefreshTokenSecret,
		},
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(config.C.RefreshTokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetTokenFromBearer gets id token from Bearer string.
func GetTokenFromBearer(str string) (string, error) {
	if !strings.Contains(str, "Bearer") {
		return "", model.NewAuthError(errors.New("Invalid token format"))
	}
	token := strings.TrimSpace(strings.Replace(str, "Bearer", "", 1))
	if token == "" {
		return "", model.NewAuthError(errors.New("Invalid token format"))
	}
	return token, nil
}

// WithToken sets token data to context.
func SetTokenToContext(ctx context.Context, token string) (context.Context, error) {
	if token == "" {
		return nil, model.NewAuthError(errors.New("Unable to set token in context. Token is empty"))
	}
	return context.WithValue(ctx, AccessTokenKey, token), nil
}

func GetTokenFromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value(AccessTokenKey).(string)

	if !ok {
		return "", model.NewAuthError(errors.New("jwt token is missing"))
	}
	return token, nil
}

func SetTokenRefreshToContext(ctx context.Context, token string) (context.Context, error) {
	if token == "" {
		return nil, model.NewAuthError(errors.New("Unable to set token in context. Token is empty"))
	}
	return context.WithValue(ctx, RefreshTokenKey, token), nil
}

func GetTokenRefreshFromContext(ctx context.Context) (string, error) {
	token, ok := ctx.Value(RefreshTokenKey).(string)

	if !ok {
		return "", model.NewAuthError(errors.New("refreshToken token is missing"))
	}
	return token, nil
}

func ValidateTokenAndReturnClaims(tokenString string, secret []byte) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func RefreshTokens(refreshToken string) (string, string, error) {
	config.ReadConfig(config.ReadConfigOption{})
	claims, err := ValidateTokenAndReturnClaims(refreshToken, []byte(config.C.RefreshTokenSecret))
	if err != nil {
		return "", "", fmt.Errorf("invalid or expired refresh token")
	}

	// Generate a new access token using the valid refresh token
	newAccessToken, err := GenerateAccessToken(claims.UserId)
	if err != nil {
		return "", "", err
	}
	newRefreshToken, err := GenerateRefreshToken(claims.UserId)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func IsJWTExpired(tokenString string) (bool, error) {
	config.ReadConfig(config.ReadConfigOption{})

	// Parse the token with claims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(config.C.JwtTokenSecret), nil
		},
	)
	if err != nil {
		// If the error is due to expiration, return true
		if errors.Is(err, jwt.ErrTokenExpired) {
			return true, nil
		}
		return false, err // Other parsing errors
	}

	// Extract claims
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return false, errors.New("invalid token claims")
	}

	// Check if the token is expired
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return true, nil
	}

	return false, nil
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("Password cannot be empty")
	}
	if utf8.RuneCountInString(password) < 8 {
		return "", errors.New("Password must be greater than 7 characters long")
	}

	// Generate a random salt.
	salt := make([]byte, DefaultArgon2Params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", model.NewAuthError(err)
	}

	// Compute the Argon2id hash of the password using the parameters and salt.
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		DefaultArgon2Params.Iterations,
		DefaultArgon2Params.Memory,
		DefaultArgon2Params.Parallelism,
		DefaultArgon2Params.KeyLength,
	)

	// Base64 encode the salt and hashed password.
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format the final hash string to include all parameters.
	// The format is: $argon2id$v=19$m=<memory>,t=<time>,p=<threads>$<salt>$<hash>
	encodedHash := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		DefaultArgon2Params.Memory,
		DefaultArgon2Params.Iterations,
		DefaultArgon2Params.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

// VerifyPassword compares a password with a stored Argon2 hash
func VerifyPassword(password, encodedHash string) error {
	// Expected format:
	// $argon2id$v=19$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return errors.New("invalid hash format")
	}

	// Use an Argon2Params struct to hold the parameters.
	var p Argon2Params
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return fmt.Errorf("invalid hash parameters: %w", err)
	}

	// Decode the salt.
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("invalid salt encoding: %w", err)
	}

	// Decode the stored hash.
	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("invalid hash encoding: %w", err)
	}

	// Compute the hash with the same parameters and salt.
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		uint32(len(expectedHash)),
	)

	// Compare computed and expected hashes using constant-time comparison.
	if subtle.ConstantTimeCompare(computedHash, expectedHash) != 1 {
		return errors.New("password does not match")
	}

	return nil
}

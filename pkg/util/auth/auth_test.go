package auth_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"sheng-go-backend/config"
	"sheng-go-backend/pkg/util/auth"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(t *testing.T) string
		act     func(t *testing.T, userId string) (string, error)
		assert  func(t *testing.T, token string, err error, userId string)
	}{
		{
			name: "Validate generated token",
			arrange: func(t *testing.T) string {
				return "12345"
			},
			act: func(t *testing.T, userId string) (string, error) {
				return auth.GenerateAccessToken(userId)
			},
			assert: func(t *testing.T, token string, err error, userId string) {
				if err != nil {
					t.Fatalf("Token generation failed:%v", err)

					if token == "" {
						t.Fatal("Expected a gone go empty string")
					}

					// Parse the token to check the claims

					parsedToken, err := jwt.Parse(
						token,
						func(token *jwt.Token) (interface{}, error) {
							// verify the token uses the expected signing method

							if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
								return nil, fmt.Errorf(
									"unexpected singing method:%v",
									token.Header["alg"],
								)
							}
							return []byte(config.C.JwtTokenSecret), nil
						},
					)
					if err != nil {
						t.Fatalf("Error parsing token:%v", err)
					}
					if !parsedToken.Valid {
						t.Fatal("Invalid token:")
					}

					claims, ok := parsedToken.Claims.(jwt.MapClaims)
					if !ok {
						t.Fatal("Expected token to be of type jwt.MapClaims")
					}
					// Check that the user_id claim matches.
					if claims["user_id"] != userId {
						t.Errorf("Expected user_id claim %q, got %v", userId, claims["user_id"])
					}

					// Verify that the exp claim is in the future.
					expFloat, ok := claims["exp"].(float64)
					if !ok {
						t.Errorf("Expected exp claim to be a number, got %T", claims["exp"])
					}
					expTime := time.Unix(int64(expFloat), 0)
					if expTime.Before(time.Now()) {
						t.Errorf("Expected exp claim to be in the future, got %v", expTime)
					}
				}
			},
		},
	}

	// Execute the tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userId := tt.arrange(t)

			token, err := tt.act(t, userId)

			tt.assert(t, token, err, userId)
		})
	}
}

func TestGetTokenFromBearer(t *testing.T) {
	tests := []struct {
		name    string
		arrange func() string
		act     func(t *testing.T, bearerToken string) (string, error)
		assert  func(t *testing.T, token string, err error)
	}{
		{
			name: "Should return correct token",
			arrange: func() string {
				return "Bearer sometoken"
			},
			act: func(t *testing.T, bearerToken string) (string, error) {
				return auth.GetTokenFromBearer(bearerToken)
			},
			assert: func(t *testing.T, token string, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, token)
				assert.Equal(t, token, "sometoken")
			},
		},
		{
			name: "Should return error if token is empty",
			arrange: func() string {
				return ""
			},
			act: func(t *testing.T, bearerToken string) (string, error) {
				return auth.GetTokenFromBearer(bearerToken)
			},
			assert: func(t *testing.T, token string, err error) {
				assert.Equal(t, token, "")
				assert.NotNil(t, err)
			},
		},
		{
			name: "Should return error Bearer is missing from token",
			arrange: func() string {
				return "sometoken"
			},
			act: func(t *testing.T, bearerToken string) (string, error) {
				return auth.GetTokenFromBearer(bearerToken)
			},
			assert: func(t *testing.T, token string, err error) {
				assert.Equal(t, token, "")
				assert.NotNil(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bearerToken := tt.arrange()
			token, err := tt.act(t, bearerToken)
			tt.assert(t, token, err)
		})
	}
}

func TestSetTokenToContext(t *testing.T) {
	tests := []struct {
		name    string
		arrange func() (string, context.Context)
		act     func(t *testing.T, token string, ctx context.Context) (context.Context, error)
		assert  func(t *testing.T, ctx context.Context, err error)
	}{
		{
			name: "Should set token to context",
			arrange: func() (string, context.Context) {
				ctx := context.Background()
				return "sometoken", ctx
			},
			act: func(t *testing.T, token string, ctx context.Context) (context.Context, error) {
				ctxWithToken, err := auth.SetTokenToContext(ctx, token)

				return ctxWithToken, err
			},
			assert: func(t *testing.T, ctx context.Context, err error) {
				assert.Nil(t, err)
				token, ok := ctx.Value(auth.AccessTokenKey).(string)
				assert.Equal(t, true, ok)
				assert.Equal(t, "sometoken", token)
			},
		},
		{
			name: "Should return error when token is empty",
			arrange: func() (string, context.Context) {
				ctx := context.Background()
				return "", ctx
			},
			act: func(t *testing.T, token string, ctx context.Context) (context.Context, error) {
				return auth.SetTokenToContext(ctx, token)
			},
			assert: func(t *testing.T, ctx context.Context, err error) {
				assert.NotNil(t, err)
				assert.Nil(t, ctx)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, ctx := tt.arrange()
			updatedContext, err := tt.act(t, token, ctx)
			tt.assert(t, updatedContext, err)
		})
	}
}

func TestGetTokenFromContext(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(t *testing.T) context.Context
		act     func(ctx context.Context) (string, error)
		assert  func(t *testing.T, token string, err error)
	}{{
		name: "Should get token from context",
		arrange: func(t *testing.T) context.Context {
			ctx := context.Background()
			ctxWithToken, err := auth.SetTokenToContext(ctx, "sometoken")
			if err != nil {
				t.Error(err)
				t.FailNow()
			}
			return ctxWithToken
		},
		act: func(ctx context.Context) (string, error) {
			return auth.GetTokenFromContext(ctx)
		},
		assert: func(t *testing.T, token string, err error) {
			assert.Nil(t, err)
			assert.Equal(t, "sometoken", token)
		},
	}, {
		name: "Should return error when no token is set ",
		arrange: func(t *testing.T) context.Context {
			return context.Background()
		},
		act: func(ctx context.Context) (string, error) {
			return auth.GetTokenFromContext(ctx)
		},
		assert: func(t *testing.T, token string, err error) {
			assert.Equal(t, token, "")
			assert.NotNil(t, err)
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.arrange(t)
			token, err := tt.act(ctx)
			tt.assert(t, token, err)
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name    string
		arrange func() string
		act     func(plainPassword string) (string, error)
		assert  func(t *testing.T, hashedPassword string, err error)
	}{{
		name: "Should hash the pasword",
		arrange: func() string {
			return "somepasword"
		},
		act: func(plainPassword string) (string, error) {
			return auth.HashPassword(plainPassword)
		},
		assert: func(t *testing.T, hashedPassword string, err error) {
			assert.Nil(t, err)
			assert.NotEmpty(t, hashedPassword)

			parts := strings.Split(hashedPassword, "$")
			assert.Equal(t, len(parts), 6)
			assert.Equal(t, parts[1], "argon2id")
			assert.Equal(t, parts[2], "v=19")
			assert.Equal(t, parts[3], "m=65536,t=3,p=2")
			assert.Equal(t, len(parts[4]), base64.RawStdEncoding.EncodedLen(16))
			assert.Equal(t, len(parts[5]), base64.RawStdEncoding.EncodedLen(32))
		},
	}, {
		name: "Should throw error if password provided is empty",
		arrange: func() string {
			return ""
		},
		act: func(plainPassword string) (string, error) {
			return auth.HashPassword(plainPassword)
		},
		assert: func(t *testing.T, hashedPassword string, err error) {
			assert.NotNil(t, err)
			assert.Equal(t, hashedPassword, "")
		},
	}, {
		name: "Should throw error if pasword is less than 8 chars long",
		arrange: func() string {
			return "passwod"
		},
		act: func(plainPassword string) (string, error) {
			return auth.HashPassword(plainPassword)
		},
		assert: func(t *testing.T, hashedPassword string, err error) {
			assert.NotNil(t, err)
			assert.Equal(t, hashedPassword, "")
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plainPassword := tt.arrange()
			hashedPassword, err := tt.act(plainPassword)
			tt.assert(t, hashedPassword, err)
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	tests := []struct {
		name    string
		arrange func() (plainPassword string, encodedHash string)
		act     func(plainPassword, encodedHash string) error
		assert  func(t *testing.T, err error)
	}{
		{
			name: "Should verify correct password",
			arrange: func() (string, string) {
				plain := "mysecret"
				encoded, err := auth.HashPassword(plain)
				if err != nil {
					t.Fatalf("failed to hash password: %v", err)
				}
				return plain, encoded
			},
			act: func(plainPassword, encodedHash string) error {
				return auth.VerifyPassword(plainPassword, encodedHash)
			},
			assert: func(t *testing.T, err error) {
				assert.Nil(t, err, "Expected no error when verifying correct password")
			},
		},
		{
			name: "Should fail verification for incorrect password",
			arrange: func() (string, string) {
				// Hash the correct password.
				correct := "mysecret"
				encoded, err := auth.HashPassword(correct)
				if err != nil {
					t.Fatalf("failed to hash password: %v", err)
				}
				// Use a different plain password for verification.
				return "wrongsecret", encoded
			},
			act: func(plainPassword, encodedHash string) error {
				return auth.VerifyPassword(plainPassword, encodedHash)
			},
			assert: func(t *testing.T, err error) {
				assert.NotNil(t, err, "Expected error when password does not match")
			},
		},
		{
			name: "Should fail for invalid hash format",
			arrange: func() (string, string) {
				// Provide a valid plain password but an invalid encoded hash.
				return "mysecret", "invalid-hash"
			},
			act: func(plainPassword, encodedHash string) error {
				return auth.VerifyPassword(plainPassword, encodedHash)
			},
			assert: func(t *testing.T, err error) {
				assert.NotNil(t, err, "Expected error for invalid hash format")
				assert.Contains(t, err.Error(), "invalid hash format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plainPassword, encodedHash := tt.arrange()
			err := tt.act(plainPassword, encodedHash)
			tt.assert(t, err)
		})
	}
}

func TestValidateTokenAndReturnClaims(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(t *testing.T) string
		act     func(token string) (*auth.CustomClaims, error)
		assert  func(t *testing.T, claims *auth.CustomClaims, err error)
	}{
		{
			name: "Should return valid claims",
			arrange: func(t *testing.T) string {
				token, err := auth.GenerateAccessToken("1234")
				if err != nil {

					t.Error(err)
					t.FailNow()
				}
				return token
			},
			act: func(token string) (*auth.CustomClaims, error) {
				config.ReadConfig(config.ReadConfigOption{})
				return auth.ValidateTokenAndReturnClaims(token, []byte(config.C.JwtTokenSecret))
			},
			assert: func(t *testing.T, claims *auth.CustomClaims, err error) {
				assert.Nil(t, err, "Error should be nil")
				assert.NotNil(t, claims, "Claim should not be null")
				assert.Equal(t, claims.UserId, "1234")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.arrange(t)
			claims, err := tt.act(token)
			tt.assert(t, claims, err)
		})
	}
}

func TestRefreshTokens(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(t *testing.T) string
		act     func(token string) (accessToken string, refreshToken string, err error)
		assert  func(t *testing.T, accessToken string, refreshToken string, err error)
	}{{
		name: "Should generate valid tokens",
		arrange: func(t *testing.T) string {
			refreshToken, err := auth.GenerateRefreshToken("1234")
			if err != nil {
				t.Error(err, "Failed to generate token")
				t.FailNow()
			}
			return refreshToken
		},
		act: func(token string) (accessToken string, refreshToken string, err error) {
			return auth.RefreshTokens(token)
		},
		assert: func(t *testing.T, accessToken string, refreshToken string, err error) {
			assert.Nil(t, err)
			assert.NotNil(t, accessToken)
			assert.NotNil(t, refreshToken)
		},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken := tt.arrange(t)
			accessToken, refrefreshToken, err := tt.act(refreshToken)
			tt.assert(t, accessToken, refrefreshToken, err)
		})
	}
}

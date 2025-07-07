package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jacosy/go-web-server/internal/auth"
)

const (
	tokenSecret = "chirpy_secret"
)

var userID = uuid.New()

func TestMakeJWT(t *testing.T) {
	jwt, err := auth.MakeJWT(userID, tokenSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	if jwt == "" {
		t.Fatal("JWT should not be empty")
	}
}

func TestValidateJWT(t *testing.T) {
	authToken, err := auth.MakeJWT(userID, tokenSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v", err)
	}

	testCases := []struct {
		name         string
		authToken    string
		tokenSecret  string
		expectError  bool
		expectUserID uuid.UUID
	}{
		{
			name:         "Invalid token",
			authToken:    "invalid.token.string",
			tokenSecret:  tokenSecret,
			expectError:  true,
			expectUserID: uuid.Nil,
		},
		{
			name:         "Wrong secret",
			authToken:    authToken,
			tokenSecret:  "fake_secret",
			expectError:  true,
			expectUserID: uuid.Nil,
		},
		{
			name:         "valid token",
			authToken:    authToken,
			tokenSecret:  tokenSecret,
			expectError:  false,
			expectUserID: userID,
		},
	}

	for _, tc := range testCases {
		parsedUserID, err := auth.ValidateJWT(tc.authToken, tc.tokenSecret)
		if tc.expectError && err == nil {
			t.Fatalf("Expected error for test case '%s', but got none", tc.name)
		}

		if !tc.expectError {
			if err != nil {
				t.Fatalf("Unexpected error for test case '%s': %v", tc.name, err)
			} else if parsedUserID != tc.expectUserID {
				t.Fatalf("Expected userID '%s' for test case '%s', but got '%s'", tc.expectUserID, tc.name, parsedUserID)
			}
		}
	}
}

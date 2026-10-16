package myjwt_test

import (
	"testing"

	myjwt "github.com/VialikovIvan/fitness_app/internal/pkg/jwt"
)

func TestJWTAuthManager(t *testing.T) {
	tests := []struct {
		name           string
		authSecret         string
		authSecretToFail   string // keep empty unless want the test case to fail because of wrong secret
		tokenStrToFail string // keep empty unless want the test case to fail because of invalid token
		wantErr        bool
	}{
		{
			name:    "success",
			authSecret:  "secret",
			wantErr: false,
		},
		{
			name:         "wrong secret",
			authSecret:       "secret",
			authSecretToFail: "anotherSecret",
			wantErr:      true,
		},
		{
			name:           "invalid token",
			authSecret:         "secret",
			tokenStrToFail: "invalidToken",
			wantErr:        true,
		},
	}

	id := int64(12345678)
	role := "user"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := myjwt.New(tt.authSecret, "")

			tokenStr, _ := m.GenerateAuth(t.Context(), id, role)

			if tt.authSecretToFail != "" {
				m = myjwt.New(tt.authSecretToFail, "")
			}

			if tt.tokenStrToFail != "" {
				tokenStr = tt.tokenStrToFail
			}

			gotID, gotRole, err := m.ValidateAuth(t.Context(), tokenStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && gotID != id {
				t.Errorf("got = %v, want %v", gotID, id)
			}

			if !tt.wantErr && gotRole != role {
				t.Errorf("got = %v, want %v", gotRole, role)
			}
		})
	}
}

package jwt_test

import (
	"testing"

	myjwt "github.com/VialikovIvan/fitness_app/internal/pkg/jwt"
)

func TestJWTManager(t *testing.T) {
	tests := []struct {
		name           string
		secret         string
		secretToFail   string // keep empty unless want the test case to fail because of wrong secret
		tokenStrToFail string // keep empty unless want the test case to fail because of invalid token
		wantErr        bool
	}{
		{
			name:    "success",
			secret:  "secret",
			wantErr: false,
		},
		{
			name:         "wrong secret",
			secret:       "secret",
			secretToFail: "anotherSecret",
			wantErr:      true,
		},
		{
			name:           "invalid token",
			secret:         "secret",
			tokenStrToFail: "invalidToken",
			wantErr:        true,
		},
	}

	id := int64(12345678)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := myjwt.New(tt.secret)

			tokenStr, _ := m.Generate(t.Context(), id)

			if tt.secretToFail != "" {
				m = myjwt.New(tt.secretToFail)
			}

			if tt.tokenStrToFail != "" {
				tokenStr = tt.tokenStrToFail
			}

			gotID, err := m.Validate(t.Context(), tokenStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && gotID != id {
				t.Errorf("got = %v, want %v", gotID, id)
			}
		})
	}
}

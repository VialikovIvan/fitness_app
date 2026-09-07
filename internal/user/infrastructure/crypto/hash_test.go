package crypto_test

import (
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/user/infrastructure/crypto"
)

func TestArgon2Manager(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		passwordToFail string // Keep empty unless want the test case to return false
		corruptHash    string // Keep empty unless want the test case to return error
		want           bool
		wantErr        bool
	}{
		{
			name:     "success",
			password: "12345678",
			want:     true,
			wantErr:  false,
		},
		{
			name:           "wrong password",
			password:       "12345678",
			passwordToFail: "12345679",
			want:           false,
			wantErr:        false,
		},
		{
			name:        "corrupt encoded password error",
			password:    "12345678",
			corruptHash: "$this$is$a$corrupt$hash",
			want:        false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := crypto.NewArgon2()

			var hash string
			if tt.corruptHash != "" {
				hash = tt.corruptHash
			} else {
				hash, _ = a.Hash(t.Context(), tt.password)
			}

			if tt.passwordToFail != "" {
				tt.password = tt.passwordToFail
			}

			got, err := a.Verify(hash, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

package domain_test

import (
	"strings"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/user/domain"
)

func TestIsValidPasswordLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   bool
	}{
		{"min", 8, true},
		{"max", 72, true},
		{"min - 1", 7, false},
		{"max + 1", 73, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password := strings.Repeat("A", tt.length)
			if got := domain.IsValidPasswordLength(password); got != tt.want {
				t.Errorf("length %d: got %v, want %v", tt.length, got, tt.want)
			}
		})
	}
}

func TestIsValidEmailFormat(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid email", "testemail@gmail.com", true},
		{"invalid email", "testemail.gmail.com", false},
		{"contains '<' symbol", "Test Email <test.email@gmail.com>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.IsValidEmailFormat(tt.email); got != tt.want {
				t.Errorf("email %v: got %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

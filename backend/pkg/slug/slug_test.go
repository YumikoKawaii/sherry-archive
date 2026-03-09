package slug_test

import (
	"testing"

	"github.com/yumikokawaii/sherry-archive/pkg/slug"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		title    string
		suffix   string
		expected string
	}{
		{"One Piece", "abc123", "one-piece-abc123"},
		{"One Piece", "", "one-piece"},
		{"Naruto: Shippuden!", "xyz", "naruto-shippuden-xyz"},
		{"  spaces  ", "", "spaces"},
	}

	for _, tt := range tests {
		got := slug.Generate(tt.title, tt.suffix)
		if got != tt.expected {
			t.Errorf("Generate(%q, %q) = %q, want %q", tt.title, tt.suffix, got, tt.expected)
		}
	}
}

func TestMake(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Bleach", "bleach"},
		{"Attack on Titan", "attack-on-titan"},
		{"Fullmetal Alchemist: Brotherhood", "fullmetal-alchemist-brotherhood"},
	}

	for _, tt := range tests {
		got := slug.Make(tt.input)
		if got != tt.expected {
			t.Errorf("Make(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

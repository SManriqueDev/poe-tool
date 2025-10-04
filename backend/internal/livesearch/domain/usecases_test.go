package domain

import (
	"testing"
)

func TestExtractSearchID(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		hasError bool
	}{
		{
			name:     "valid PoE trade URL",
			url:      "https://www.pathofexile.com/trade2/search/poe2/Rise%20of%20the%20Abyssal/4nVv4ggf9",
			expected: "4nVv4ggf9",
			hasError: false,
		},
		{
			name:     "valid URL with underscores",
			url:      "https://www.pathofexile.com/trade2/search/poe2/Standard/abc_123_def",
			expected: "abc_123_def",
			hasError: false,
		},
		{
			name:     "invalid URL - not PoE",
			url:      "https://google.com/search?q=test",
			expected: "",
			hasError: true,
		},
		{
			name:     "invalid URL - empty",
			url:      "",
			expected: "",
			hasError: true,
		},
		{
			name:     "invalid URL - no search ID",
			url:      "https://www.pathofexile.com/trade2/search/poe2/",
			expected: "",
			hasError: true,
		},
		{
			name:     "invalid search ID - special chars",
			url:      "https://www.pathofexile.com/trade2/search/poe2/Standard/test@123",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractSearchID(tt.url)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

package security

import "testing"

func TestRandom(t *testing.T) {
	scenarios := []struct {
		char     int
		expected int
	}{
		{
			char:     5,
			expected: 5,
		},
		{
			char:     15,
			expected: 15,
		},
		{
			char:     15,
			expected: 15,
		},
	}

	for _, tt := range scenarios {
		t.Run("", func(t *testing.T) {
			s := GenerateRandomString(tt.char)
			if len(s) != tt.expected {
				t.Errorf("(%d) Expected error", tt.expected)
			}
		})
	}
}

func TestFailRandom(t *testing.T) {
	scenarios := []struct {
		char     int
		expected int
	}{
		{
			char:     50,
			expected: 5,
		},
		{
			char:     10,
			expected: 15,
		},
		{
			char:     7,
			expected: 5,
		},
	}

	for _, tt := range scenarios {
		t.Run("", func(t *testing.T) {
			s := GenerateRandomString(tt.char)
			if len(s) == tt.expected {
				t.Errorf("(%d) Expected error got same", tt.expected)
			}
		})
	}
}

func TestGenerateRandomEmail(t *testing.T) {
	scenarios := []struct {
		char     int
		expected int
	}{
		{
			char:     50,
			expected: 60,
		},
		{
			char:     5,
			expected: 15,
		},
		{
			char:     50,
			expected: 60,
		},
	}

	for _, tt := range scenarios {
		t.Run("", func(t *testing.T) {
			s := GenerateRandomEmail(tt.char)
			if len(s) != tt.expected {
				t.Errorf("(%d) Expected error got same", tt.expected)
			}
		})
	}
}

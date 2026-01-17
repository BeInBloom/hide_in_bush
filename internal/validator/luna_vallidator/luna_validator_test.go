package lunavallidator

import "testing"

func TestValidate(t *testing.T) {
	validator := New()

	tests := []struct {
		name  string
		input []byte
		want  bool
	}{
		{"valid_79927398713", []byte("79927398713"), true},
		{"valid_12345678903", []byte("12345678903"), true},
		{"valid_49927398716", []byte("49927398716"), true},
		{"valid_0", []byte("0"), true},
		{"invalid_79927398714", []byte("79927398714"), false},
		{"invalid_12345678902", []byte("12345678902"), false},
		{"invalid_1234567890", []byte("1234567890"), false},
		{"empty", []byte(""), false},
		{"non_numeric", []byte("abc123"), false},
		{"with_spaces", []byte("7992 7398 713"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := validator.Validate(tt.input)
			if got != tt.want {
				t.Errorf("Validate(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidLuna(t *testing.T) {
	tests := []struct {
		name   string
		number int
		want   bool
	}{
		{"valid_0", 0, true},
		{"valid_18", 18, true},
		{"valid_79927398713", 79927398713, true},
		{"invalid_1", 1, false},
		{"invalid_12", 12, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidLuna(tt.number)
			if got != tt.want {
				t.Errorf("isValidLuna(%d) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}

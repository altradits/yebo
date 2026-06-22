package utils

import "testing"

func TestValidatePhone(t *testing.T) {
	cases := []struct {
		in string
		ok bool
	}{
		{"+254712345678", true},
		{"+1234567890", true},
		{"not-a-phone", false},
		{"", false},
		{"123", false},
	}
	for _, c := range cases {
		err := ValidatePhone(c.in)
		if c.ok && err != nil {
			t.Errorf("ValidatePhone(%q) unexpected error: %v", c.in, err)
		}
		if !c.ok && err == nil {
			t.Errorf("ValidatePhone(%q) expected error, got nil", c.in)
		}
	}
}

func TestNormalisePhone(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"0712345678", "+254712345678"},
		{"712345678", "+254712345678"},
		{"254712345678", "+254712345678"},
		{"+254712345678", "+254712345678"},
	}
	for _, c := range cases {
		got := NormalisePhone(c.in)
		if got != c.want {
			t.Errorf("NormalisePhone(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestValidateSatsAmount(t *testing.T) {
	if err := ValidateSatsAmount(1000, 100, 10000); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ValidateSatsAmount(0, 100, 10000); err == nil {
		t.Error("expected error for zero amount")
	}
	if err := ValidateSatsAmount(50, 100, 10000); err == nil {
		t.Error("expected error for below minimum")
	}
	if err := ValidateSatsAmount(20000, 100, 10000); err == nil {
		t.Error("expected error for above maximum")
	}
}

func TestValidatePIN(t *testing.T) {
	if err := ValidatePIN("123456"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := ValidatePIN("12345"); err == nil {
		t.Error("expected error for short PIN")
	}
	if err := ValidatePIN("12345a"); err == nil {
		t.Error("expected error for non-digit PIN")
	}
}

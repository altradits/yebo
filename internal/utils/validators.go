package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	phoneRE = regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)
	emailRE = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ValidatePhone validates an international phone number.
func ValidatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if !phoneRE.MatchString(phone) {
		return fmt.Errorf("invalid phone number: %q", phone)
	}
	return nil
}

// NormalisePhone converts 07XXXXXXXX → +2547XXXXXXXX for Kenya.
func NormalisePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if strings.HasPrefix(phone, "07") {
		return "+254" + phone[1:]
	}
	if strings.HasPrefix(phone, "7") && len(phone) == 9 {
		return "+254" + phone
	}
	if strings.HasPrefix(phone, "254") && !strings.HasPrefix(phone, "+") {
		return "+" + phone
	}
	return phone
}

// ValidateEmail validates an email address.
func ValidateEmail(email string) error {
	if !emailRE.MatchString(strings.TrimSpace(email)) {
		return fmt.Errorf("invalid email: %q", email)
	}
	return nil
}

// ValidateSatsAmount ensures an amount is positive and within reasonable range.
func ValidateSatsAmount(sats int64, minSats, maxSats int64) error {
	if sats <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	if sats < minSats {
		return fmt.Errorf("minimum amount is %d sats", minSats)
	}
	if maxSats > 0 && sats > maxSats {
		return fmt.Errorf("maximum amount is %d sats", maxSats)
	}
	return nil
}

// ValidatePIN ensures a PIN is exactly 6 digits.
func ValidatePIN(pin string) error {
	pin = strings.TrimSpace(pin)
	if len(pin) != 6 {
		return fmt.Errorf("PIN must be exactly 6 digits")
	}
	for _, c := range pin {
		if c < '0' || c > '9' {
			return fmt.Errorf("PIN must contain only digits")
		}
	}
	return nil
}

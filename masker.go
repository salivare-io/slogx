package slogx

import (
	"fmt"
	"strings"
)

// MaskType defines the category of data being masked to apply the appropriate redaction strategy.
type MaskType int

const (
	// MaskDefault replaces the value with a generic [MASKED] tag.
	MaskDefault MaskType = iota
	// MaskEmail redacts email addresses while preserving parts of the username and the domain.
	MaskEmail
	// MaskPhone redacts phone numbers while preserving the prefix and last few digits.
	MaskPhone
	// MaskCard redacts credit card numbers, showing only the first and last four digits.
	MaskCard
	// MaskSecret completely hides the value and replaces it with a [SECRET] tag.
	MaskSecret
)

// Masker is the interface that wraps the basic Mask method.
// Any custom masking logic should implement this interface.
type Masker interface {
	Mask(value any, mType MaskType) any
}

// DefaultMasker provides a standard implementation of the Masker interface
// with built-in rules for common sensitive data types.
type DefaultMasker struct{}

// Mask processes the input value based on the specified MaskType.
// It converts the value to a string representation before applying redaction rules.
func (m *DefaultMasker) Mask(value any, mType MaskType) any {
	valStr := fmt.Sprintf("%v", value)
	switch mType {
	case MaskEmail:
		return maskEmail(valStr)
	case MaskPhone:
		return maskPhone(valStr)
	case MaskCard:
		return maskCard(valStr)
	case MaskSecret:
		return "[SECRET]"
	default:
		return "[MASKED]"
	}
}

// maskEmail redacts an email address (e.g., "antonioh@gmail.com" -> "an***h@gmail.com").
func maskEmail(s string) string {
	parts := strings.Split(s, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	user := parts[0]
	domain := parts[1]

	if len(user) <= 2 {
		return user[:1] + "***@" + domain
	}
	// Take first 2 characters, add fixed asterisks and the last character
	return user[:2] + "***" + user[len(user)-1:] + "@" + domain
}

// maskPhone redacts a phone number (e.g., "+7 911 222 3456" -> "+7 9*******456").
func maskPhone(s string) string {
	if len(s) < 8 {
		return "***"
	}
	// Keeps the first 4 characters and the last 3
	return s[:4] + "*******" + s[len(s)-3:]
}

// maskCard redacts a credit card number (e.g., "2313 **** **** 4321").
func maskCard(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	if len(s) < 12 {
		return "**** **** ****"
	}
	return s[:4] + " **** **** " + s[len(s)-4:]
}

package slogx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskingFunctions(t *testing.T) {
	t.Run(
		"Email", func(t *testing.T) {
			assert.Equal(t, "an***h@gmail.com", maskEmail("antonioh@gmail.com"))
			assert.Equal(t, "a***@ya.ru", maskEmail("a@ya.ru"))
		},
	)

	t.Run(
		"Phone", func(t *testing.T) {
			assert.Equal(t, "+7 9*******456", maskPhone("+7 911 222 3456"))
		},
	)

	t.Run(
		"Card", func(t *testing.T) {
			assert.Equal(t, "4276 **** **** 0000", maskCard("4276123456780000"))
		},
	)
}

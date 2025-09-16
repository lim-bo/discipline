package service

import (
	"sync"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// Package for custom validations
var (
	validate *validator.Validate
	once     sync.Once
)

func InitValidator() {
	once.Do(func() {
		validate = validator.New()
		validate.RegisterValidation("alphanum_underscore", func(fl validator.FieldLevel) bool {
			value := fl.Field().String()
			for i, char := range value {
				// Cannot be started with a digit or underscore
				if i == 0 && (unicode.IsDigit(char) || char == '_') {
					return false
				}
				// Digits, letters or underscore
				if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
					return false
				}
			}
			return true
		})
	})
}

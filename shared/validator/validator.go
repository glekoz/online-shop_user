package validator

import (
	"regexp"
	"slices"
	"unicode"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// func (v *Validator) Error() string {
// 	var sb strings.Builder
// 	sb.WriteString("validation errors:\n")
// 	for field, msg := range v.Errors {
// 		sb.WriteString(" " + field + ": " + msg + ";\n")
// 	}
// 	return sb.String()
// }

func In(value string, list ...string) bool {
	return slices.Contains(list, value)
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Unique(values []string) bool {
	uniqueValues := make(map[string]struct{})

	for _, value := range values {
		uniqueValues[value] = struct{}{}
	}

	return len(values) == len(uniqueValues)
}

func ValidPassword(password string) bool {
	if len(password) < 8 || len(password) > 16 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		case unicode.IsSpace(r):
			return false
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

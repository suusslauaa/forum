package handlers

import (
	"fmt"
	"regexp"
)

func validateCredentials(req RegisterRequest) error {
	// Email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return fmt.Errorf("Incorrect email format")
	}

	// Username
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	if !usernameRegex.MatchString(req.Username) {
		return fmt.Errorf("The username must contain 3-20 characters (a-z, 0-9)")
	}

	// Password
	if len(req.Password) < 8 {
		return fmt.Errorf("The password must contain at least 8 characters.")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(req.Password) {
		return fmt.Errorf("The password must contain at least one uppercase letter.")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(req.Password) {
		return fmt.Errorf("The password must contain at least one digit.")
	}
	if !regexp.MustCompile(`[!@#$%^&*]`).MatchString(req.Password) {
		return fmt.Errorf("The password must contain at least one special character. (!@#$%^&*)")
	}

	return nil
}

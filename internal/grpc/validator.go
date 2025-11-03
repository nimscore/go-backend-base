package grpc

import (
	"errors"
	"regexp"
)

var ErrInvalid = errors.New("invalid")

func ValidateUserName(name string) error {
	if len(name) < 5 {
		return ErrInvalid
	}

	return nil
}

func ValidateUserEmail(email string) error {
	regex, err := regexp.Compile(`[a-z\-\_\.]+@[a-z\-\_\.]+`)
	if err != nil {
		return ErrInvalid
	}

	if !regex.MatchString(email) {
		return ErrInvalid
	}

	return nil
}

func ValidateCommunityName(name string) error {
	if len(name) < 5 {
		return ErrInvalid
	}

	return nil
}

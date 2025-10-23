package invitecode

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound       = errors.New("invite code not found")
	ErrAlreadyExists  = errors.New("invite code already exists")
	ErrInvalidCode    = errors.New("invalid invite code")
	ErrCodeExpired    = errors.New("invite code has expired")
	ErrCodeExhausted  = errors.New("invite code usage limit reached")
	ErrCodeRevoked    = errors.New("invite code has been revoked")
	ErrAlreadyUsed    = errors.New("you have already used an invite code")
	ErrHasQuota       = errors.New("you already have account quota")
	ErrInvalidMaxUses = errors.New("max uses must be -1 (unlimited) or positive number")
)

func NotFoundError(code string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, code)
}

func AlreadyExistsError(code string) error {
	return fmt.Errorf("%w: %s", ErrAlreadyExists, code)
}

func CodeExpiredError(code string) error {
	return fmt.Errorf("%w: %s", ErrCodeExpired, code)
}

func CodeExhaustedError(code string) error {
	return fmt.Errorf("%w: %s", ErrCodeExhausted, code)
}

func CodeRevokedError(code string) error {
	return fmt.Errorf("%w: %s", ErrCodeRevoked, code)
}

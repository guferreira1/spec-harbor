package domain

import (
	"errors"
	"fmt"
	"strings"
)

const maxChangeIDLength = 128

// ChangeID is a validated, single-path-segment OpenSpec change identifier.
type ChangeID struct {
	value string
}

func NewChangeID(raw string) (ChangeID, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ChangeID{}, errors.New("change id is required")
	}
	if strings.ContainsAny(value, "/\\") {
		return ChangeID{}, errors.New("change id must be a single path segment")
	}
	if value == "." || value == ".." || strings.Contains(value, "..") {
		return ChangeID{}, errors.New("change id must not contain '.' or '..' path sequences")
	}
	if strings.HasPrefix(value, ".") {
		return ChangeID{}, errors.New("change id must not start with '.'")
	}
	if strings.HasPrefix(value, "-") {
		return ChangeID{}, errors.New("change id must not start with '-'")
	}
	if len(value) > maxChangeIDLength {
		return ChangeID{}, fmt.Errorf("change id must be at most %d characters", maxChangeIDLength)
	}
	for _, character := range value {
		if !isChangeIDCharacter(character) {
			return ChangeID{}, fmt.Errorf("change id contains unsupported character %q", character)
		}
	}

	return ChangeID{value: value}, nil
}

func (changeID ChangeID) String() string {
	return changeID.value
}

func isChangeIDCharacter(character rune) bool {
	switch {
	case character >= 'a' && character <= 'z':
		return true
	case character >= 'A' && character <= 'Z':
		return true
	case character >= '0' && character <= '9':
		return true
	case character == '-' || character == '_' || character == '.':
		return true
	default:
		return false
	}
}

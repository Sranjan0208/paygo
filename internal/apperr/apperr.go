package apperr

import (
	"errors"
	"fmt"
)

type Kind int // Invalid, NotFound, Conflict, Declined, Unknown, Internal

const (
	Invalid  Kind = iota + 1 // bad input from the caller
	NotFound                 // the thing does not exist
	Conflict                 // clashes with the current state
	Declined                 // a final "no" from the bank
	Unknown                  // outcome unknown: money may have moved
	Internal                 // our bug or infrastructure fault
)

// String makes Kind print as a word in logs instead of a number
func (k Kind) String() string {
	switch k {
	case Invalid:
		return "invalid"
	case NotFound:
		return "not_found"
	case Conflict:
		return "conflict"
	case Declined:
		return "declined"
	case Unknown:
		return "unknown"
	case Internal:
		return "internal"
	default:
		return fmt.Sprintf("kind(%d)", int(k))
	}
}

// Error is the one error type used across paygo.
type Error struct {
	Kind Kind   // category: drives behaviour
	Code string // stable machine code, e.g. "payment_expired"
	Msg  string // human-readable, safe to show to API clients
	Err  error  // underlying cause, for logs only
}

// Compile-time check that *Error satisfies the error interface
var _ error = (*Error)(nil)

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Msg, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Msg)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// New creates an error where a problem is first detected
func New(kind Kind, code, msg string) *Error {
	return &Error{Kind: kind, Code: code, Msg: msg}
}

// Wrap classifies an existing error and keeps it as the cause
func Wrap(err error, kind Kind, code, msg string) *Error {
	return &Error{Kind: kind, Code: code, Msg: msg, Err: err}
}

// KindOf reports the kind of any error, however deeply it is wrapped
func KindOf(err error) Kind {
	if err == nil {
		return 0
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return Internal
}

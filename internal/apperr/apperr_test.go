package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestKindOf(t *testing.T) {
	t.Parallel()

	cause := errors.New("dial tcp: i/o timeout")

	tests := []struct {
		name string
		err  error
		want Kind
	}{
		{
			name: "nil error",
			err:  nil,
			want: 0,
		},
		{
			name: "plain error defaults to internal",
			err:  errors.New("boom"),
			want: Internal,
		},
		{
			name: "direct apperr",
			err:  New(Conflict, "already_captured", "payment already captured"),
			want: Conflict,
		},
		{
			name: "wrapped cause keeps our kind",
			err:  Wrap(cause, Unknown, "bank_timeout", "no response from bank"),
			want: Unknown,
		},
		{
			name: "fmt %w keeps kind",
			err:  fmt.Errorf("capture: %w", New(Declined, "card_declined", "card declined")),
			want: Declined,
		},
		{
			name: "wrapped twice keeps kind",
			err: fmt.Errorf("outer: %w",
				fmt.Errorf("inner: %w", New(NotFound, "order_not_found", "order not found"))),
			want: NotFound,
		},
		{
			name: "fmt %v loses kind",
			//nolint:errorlint // deliberate: documents what happens when %w is forgotten
			err:  fmt.Errorf("capture: %v", New(Declined, "card_declined", "card declined")),
			want: Internal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := KindOf(tc.err); got != tc.want {
				t.Errorf("KindOf() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestErrorsAsFindsTypedError(t *testing.T) {
	t.Parallel()

	inner := Wrap(errors.New("cause"), Unknown, "bank_timeout", "no response")
	err := fmt.Errorf("service: %w", inner)

	var e *Error
	if !errors.As(err, &e) {
		t.Fatal("errors.As did not find *Error in the chain")
	}
	if e.Code != "bank_timeout" {
		t.Errorf("Code = %q, want %q", e.Code, "bank_timeout")
	}
}

func TestUnwrap(t *testing.T) {
	t.Parallel()

	cause := errors.New("dial tcp: i/o timeout")
	err := Wrap(cause, Unknown, "bank_timeout", "no response from bank")

	if !errors.Is(err, cause) {
		t.Error("errors.Is cannot reach the cause through Unwrap")
	}
	if err.Unwrap() != cause {
		t.Error("Unwrap() did not return the original cause")
	}
	if got := New(Invalid, "bad_amount", "amount must be positive").Unwrap(); got != nil {
		t.Errorf("New().Unwrap() = %v, want nil", got)
	}
}

func TestErrorString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "without cause",
			err:  New(Conflict, "already_captured", "payment already captured"),
			want: "already_captured: payment already captured",
		},
		{
			name: "with cause",
			err:  Wrap(errors.New("i/o timeout"), Unknown, "bank_timeout", "no response from bank"),
			want: "bank_timeout: no response from bank: i/o timeout",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestKindString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		kind Kind
		want string
	}{
		{Invalid, "invalid"},
		{NotFound, "not_found"},
		{Conflict, "conflict"},
		{Declined, "declined"},
		{Unknown, "unknown"},
		{Internal, "internal"},
		{Kind(0), "kind(0)"},
		{Kind(99), "kind(99)"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			if got := tc.kind.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

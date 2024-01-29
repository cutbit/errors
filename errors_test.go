package errors

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		err  string
		want error
	}{
		{"", fmt.Errorf("")},
		{"foo", fmt.Errorf("foo")},
		{"foo", New("foo")},
		{"string with format specifiers: %v", errors.New("string with format specifiers: %v")},
	}

	for _, tt := range tests {
		got := New(tt.err)
		if got.Error() != tt.want.Error() {
			t.Errorf("New.Error(): got: %q, want %q", got, tt.want)
		}
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{Wrap(io.EOF), "EOF"},
	}

	for _, tt := range tests {
		got := Wrap(tt.err).Error()
		if got != tt.want {
			t.Errorf("Wrap(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{io.EOF, "EOF"},
		{Join(io.EOF, io.EOF), "EOF\nEOF"},
	}

	for _, tt := range tests {
		got := Join(tt.err).Error()
		if got != tt.want {
			t.Errorf("Join(%v): got: %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestTrack(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{io.EOF, false},
		{New("EOF"), true},
		{Wrap(io.EOF), true},
		{Join(io.EOF, New("EOF"), Wrap(io.EOF)), true},
	}

	for _, tt := range tests {
		_, got := Track(tt.err)
		if got != tt.want {
			t.Errorf("Track(%v): got: _,%v, want _,%v", tt.err, got, tt.want)
		}
	}
}

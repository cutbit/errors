package errors

import (
	"errors"
	"fmt"
	"io"
)

// structure basic error information
type structure struct {
	// errors that occur
	error
	// stack information
	*stack
}

// Cause reason for obtaining error
func (s *structure) Cause() error {
	if s == nil {
		return nil
	}
	return s.error
}

// Stack get error call stack
func (s *structure) Stack() *stack {
	if s == nil {
		return nil
	}
	return s.stack
}

// Error implement error interface
func (s *structure) Error() string {
	if s == nil {
		return ""
	}
	return s.error.Error()
}

// Unwrap error reason unpacking
func (s *structure) Unwrap() error {
	if s == nil {
		return nil
	}
	return s.error
}

// Format error formatting output
func (s *structure) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		if state.Flag('+') {
			_, _ = fmt.Fprintf(state, "%+v", s.Cause())
			s.stack.Format(state, verb)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(state, s.Error())
	case 'q':
		_, _ = fmt.Fprintf(state, "%q", s.Error())
	}
}

// structures error group slice
type structures []*structure

// Error implement error interface
func (s *structures) Error() string {
	var bytes []byte
	for index, item := range *s {
		if index > 0 {
			bytes = append(bytes, '\n')
		}
		bytes = append(bytes, item.Error()...)
	}
	return string(bytes)
}

// Unwrap error reason unpacking
func (s *structures) Unwrap() error {
	list := make([]error, 0, len(*s))
	for i := range *s {
		list = append(list, (*s)[i].Unwrap())
	}
	return errors.Join(list...)
}

// Format error formatting output
func (s *structures) Format(state fmt.State, verb rune) {
	for index := range *s {
		if index > 0 {
			_, _ = state.Write([]byte("\n"))
		}
		(*s)[index].Format(state, verb)
	}
}

// New create an error message containing the call stack
func New(cause string) error {
	return &structure{
		error: errors.New(cause),
		stack: callers(),
	}
}

// Wrap packaging error information
//
// record the stack information if no stack information is recorded
func Wrap(error error) error {
	var origin *structure
	ok := errors.As(error, &origin)
	if ok {
		return origin
	}
	return &structure{
		error: error,
		stack: callers(),
	}
}

// Join link multiple error messages
//
// add a stack for each error message that has not recorded stack information
func Join(causes ...error) error {
	list := make(structures, 0, len(causes))
	for index := range causes {
		var origin *structure
		ok := errors.As(causes[index], &origin)
		if ok {
			list = append(list, origin)
			continue
		}
		list = append(list, &structure{
			error: causes[index],
			stack: callers(),
		})
	}
	return &list
}

// Track tracking error messages
func Track(cause error) ([]StackTrace, bool) {
	var item *structure
	ok := errors.As(cause, &item)
	if ok {
		return []StackTrace{item.Stack().StackTrace()}, true
	}
	var list *structures
	ok = errors.As(cause, &list)
	if ok {
		traces := make([]StackTrace, 0, len(*list))
		for index := range *list {
			traces = append(traces, (*list)[index].Stack().StackTrace())
		}
		return traces, true
	}
	return nil, false
}

// Unwrap unpacking error message original error
func Unwrap(cause error) error {
	return errors.Unwrap(cause)
}

// Is determine whether the error message is targeted incorrectly
func Is(cause, target error) bool {
	return errors.Is(cause, target)
}

// As error message converted to target error
func As(cause error, target any) bool {
	return errors.As(cause, target)
}

package errors

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var initPC = caller()

type X struct{}

// val returns a Frame pointing to itself.
func (x X) val() Frame {
	return caller()
}

// ptr returns a Frame pointing to itself.
func (x *X) ptr() Frame {
	return caller()
}

func TestFrameFormat(t *testing.T) {
	var tests = []struct {
		Frame
		format string
		want   string
	}{{
		initPC,
		"%s",
		"stack_test.go",
	}, {
		initPC,
		"%+s",
		"github.com/cutbit/errors.init\n" +
			"\t.+/github.com/cutbit/errors/stack_test.go",
	}, {
		0,
		"%s",
		"unknown",
	}, {
		0,
		"%+s",
		"unknown",
	}, {
		initPC,
		"%d",
		"11",
	}, {
		0,
		"%d",
		"0",
	}, {
		initPC,
		"%n",
		"init",
	}, {
		func() Frame {
			var x X
			return x.ptr()
		}(),
		"%n",
		`\(\*X\).ptr`,
	}, {
		func() Frame {
			var x X
			return x.val()
		}(),
		"%n",
		"X.val",
	}, {
		0,
		"%n",
		"",
	}, {
		initPC,
		"%v",
		"stack_test.go:11",
	}, {
		initPC,
		"%+v",
		"github.com/cutbit/errors.init\n" +
			"\t.+/github.com/cutbit/errors/stack_test.go:11",
	}, {
		0,
		"%v",
		"unknown:0",
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.Frame, tt.format, tt.want)
	}
}

func TestFunctionName(t *testing.T) {
	tests := []struct {
		name, want string
	}{
		{"", ""},
		{"runtime.main", "main"},
		{"github.com/cutbit/errors.functionName", "functionName"},
		{"functionName", "functionName"},
		{"io.copyBuffer", "copyBuffer"},
		{"main.(*R).Write", "(*R).Write"},
	}

	for _, tt := range tests {
		got := functionName(tt.name)
		want := tt.want
		if got != want {
			t.Errorf("funcname(%q): want: %q, got %q", tt.name, want, got)
		}
	}
}

func TestStackTrace(t *testing.T) {
	tests := []struct {
		err  error
		want []string
	}{{
		New("ooh"), []string{
			"github.com/cutbit/errors.TestStackTrace\n" +
				"\t.+/github.com/pkg/errors/stack_test.go:123",
		},
	}, {
		Wrap(New("ooh")), []string{
			"github.com/cutbit/errors.TestStackTrace\n" +
				"\t.+/github.com/cutbit/errors/stack_test.go:128", // this is the stack of Wrap, not New
		},
	}, {
		func() error { return New("ooh") }(), []string{
			`github.com/cutbit/errors.TestStackTrace.func1` +
				"\n\t.+/github.com/pkg/errors/stack_test.go:133", // this is the stack of New
			"github.com/cutbit/errors.TestStackTrace\n" +
				"\t.+/github.com/cutbit/errors/stack_test.go:133", // this is the stack of New's caller
		},
	}}
	for i, tt := range tests {
		x, ok := tt.err.(interface {
			StackTrace() StackTrace
		})
		if !ok {
			t.Errorf("expected %#v to implement StackTrace() StackTrace", tt.err)
			continue
		}
		st := x.StackTrace()
		for j, want := range tt.want {
			testFormatRegexp(t, i, st[j], "%+v", want)
		}
	}
}

func stackTrace() StackTrace {
	const depth = 8
	var pcs [depth]uintptr
	n := runtime.Callers(1, pcs[:])
	var st stack = pcs[0:n]
	return st.StackTrace()
}

func TestStackTraceFormat(t *testing.T) {
	tests := []struct {
		StackTrace
		format string
		want   string
	}{{
		nil,
		"%s",
		`\[\]`,
	}, {
		nil,
		"%v",
		`\[\]`,
	}, {
		nil,
		"%+v",
		"",
	}, {
		nil,
		"%#v",
		`\[\]errors.Frame\(nil\)`,
	}, {
		make(StackTrace, 0),
		"%s",
		`\[\]`,
	}, {
		make(StackTrace, 0),
		"%v",
		`\[\]`,
	}, {
		make(StackTrace, 0),
		"%+v",
		"",
	}, {
		make(StackTrace, 0),
		"%#v",
		`\[\]errors.Frame{}`,
	}, {
		stackTrace()[:2],
		"%s",
		`\[stack_test.go stack_test.go\]`,
	}, {
		stackTrace()[:2],
		"%v",
		`\[stack_test.go:158 stack_test.go:205\]`,
	}, {
		stackTrace()[:2],
		"%+v",
		"\n" +
			"github.com/cutbit/errors.stackTrace\n" +
			"\t.+/github.com/cutbit/errors/stack_test.go:158\n" +
			"github.com/cutbit/errors.TestStackTraceFormat\n" +
			"\t.+/github.com/cutbit/errors/stack_test.go:209",
	}, {
		stackTrace()[:2],
		"%#v",
		`\[\]errors.Frame{stack_test.go:158, stack_test.go:217}`,
	}}

	for i, tt := range tests {
		testFormatRegexp(t, i, tt.StackTrace, tt.format, tt.want)
	}
}

// a version of runtime.Caller that returns a Frame, not a uintptr.
func caller() Frame {
	var pcs [3]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	frame, _ := frames.Next()
	return Frame(frame.PC)
}

func testFormatRegexp(t *testing.T, n int, arg interface{}, format, want string) {
	t.Helper()
	got := fmt.Sprintf(format, arg)
	gotLines := strings.SplitN(got, "\n", -1)
	wantLines := strings.SplitN(want, "\n", -1)

	if len(wantLines) > len(gotLines) {
		t.Errorf("test %d: wantLines(%d) > gotLines(%d):\n got: %q\nwant: %q", n+1, len(wantLines), len(gotLines), got, want)
		return
	}

	for i, w := range wantLines {
		match, err := regexp.MatchString(w, gotLines[i])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("test %d: line %d: fmt.Sprintf(%q, err):\n got: %q\nwant: %q", n+1, i+1, format, got, want)
		}
	}
}

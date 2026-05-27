package dbutils

import (
	"context"
	"time"
)

type (
	OptionExec    func(o *optionExec)
	OptionContext func(o *optionContext)
)

// ///////////////////////////////////////////////////////////////////////////

type defaulter interface {
	Default()
}

func apply[T any, O ~func(*T)](opts []O) *T {
	opt := new(T)
	for _, o := range opts {
		o(opt)
	}

	if d, ok := any(opt).(defaulter); ok {
		d.Default()
	}

	return opt
}

// ///////////////////////////////////////////////////////////////////////////
// funcs of optionExec

type optionExec struct {
	Values     map[string]string
	Timeout    time.Duration
	Ctx        context.Context
	SortFn     func(files []string)
	Extensions []string
}

func (o *optionExec) Default() {
	if o.Ctx == nil {
		o.Ctx = context.Background()
	}
	if len(o.Extensions) == 0 {
		o.Extensions = []string{".sql"}
	}
}

// WithValues sets os.Expand values inside the file content.
func WithValues(values map[string]string) OptionExec {
	return func(o *optionExec) {
		o.Values = values
	}
}

// WithTimeout sets the timeout for each file execution.
func WithTimeout(timeout time.Duration) OptionExec {
	return func(o *optionExec) {
		o.Timeout = timeout
	}
}

// WithContext sets the context for the file execution.
func WithExecContext(ctx context.Context) OptionExec {
	return func(o *optionExec) {
		o.Ctx = ctx
	}
}

// WithSort sets a custom sort function for files in ExecuteFolder.
// By default, files are sorted numerically by the leading number in their filename.
func WithSort(fn func(files []string)) OptionExec {
	return func(o *optionExec) {
		o.SortFn = fn
	}
}

// WithExtensions filters files in ExecuteFolder to only include files with the given extensions.
// Default is []string{".sql"}.
func WithExtensions(exts ...string) OptionExec {
	return func(o *optionExec) {
		o.Extensions = exts
	}
}

// ///////////////////////////////////////////////////////////////////////////
// funcs of optionContext

type optionContext struct {
	Ctx context.Context
}

func (o *optionContext) Default() {
	if o.Ctx == nil {
		o.Ctx = context.Background()
	}
}

// WithContext sets the context for the main function.
func WithContext(ctx context.Context) OptionContext {
	return func(o *optionContext) {
		o.Ctx = ctx
	}
}

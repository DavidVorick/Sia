package sialog

import (
	"fmt"
	"runtime"
	"strings"
)

// A ctxError is a set of contexts that characterize an error.
type ctxError struct {
	ctxs []string
}

// Error implements the error interface. It returns a formatted string
// containing the full context of the error.
func (ce ctxError) Error() string {
	return strings.Join(ce.ctxs, "\n\t")
}

// fnName returns the name of the calling function. Actually, it goes one level
// higher than that, since it is never called directly.
func fnName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		panic("function lookup failed")
	}
	pcname := runtime.FuncForPC(pc).Name()
	return strings.SplitN(pcname[strings.LastIndex(pcname, "/")+1:], ".", 2)[1]
}

// Error returns a new ctxError, including the first level of context.
func CtxError(v ...interface{}) ctxError {
	errString := strings.Trim(fmt.Sprintln(v...), "\n")
	ctx := []string{fmt.Sprintf("%s: %s", fnName(), errString)}
	return ctxError{ctx}
}

// Error returns a formatted ctxError, including the first level of context.
func CtxErrorf(fmtString string, v ...interface{}) ctxError {
	return CtxError(fmt.Sprintf(fmtString, v...))
}

// AddCtx adds context to an error message. If the error is a ctxError, the
// context is added to ctxs. If it is a standard error, a new ctxError is
// created and returned.
func AddCtx(err error, ctx string) (ce ctxError) {
	fullCtx := fmt.Sprintf("%s: %s", fnName(), ctx)
	switch err.(type) {
	case ctxError:
		ce = err.(ctxError)
		ce.ctxs = append([]string{fullCtx}, ce.ctxs...)
	case error:
		ce.ctxs = []string{fullCtx}
	}
	return
}

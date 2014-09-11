package sialog

import (
	"fmt"
	"runtime"
	"strings"
)

// A siaerror is a set of contexts that characterize an error.
type siaerror struct {
	ctxs []string
}

// Error implements the error interface. It returns a formatted string
// containing the full context of the error.
func (se siaerror) Error() string {
	return strings.Join(se.ctxs, "\n\t")
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

// Error returns a new siaerror, including the first level of context.
func Error(v ...interface{}) siaerror {
	errString := strings.Trim(fmt.Sprintln(v...), "\n")
	ctx := []string{fmt.Sprintf("%s: %s", fnName(), errString)}
	return siaerror{ctx}
}

// Error returns a formatted siaerror, including the first level of context.
func Errorf(fmtString string, v ...interface{}) siaerror {
	return Error(fmt.Sprintf(fmtString, v...))
}

// AddCtx adds context to an error message. If the error is a siaerror, the
// context is added to ctxs. If it is a standard error, a new siaerror is
// created and returned.
func AddCtx(err error, ctx string) (se siaerror) {
	fullCtx := fmt.Sprintf("%s: %s", fnName(), ctx)
	switch err.(type) {
	case siaerror:
		se = err.(siaerror)
		se.ctxs = append([]string{fullCtx}, se.ctxs...)
	case error:
		se.ctxs = []string{fullCtx}
	}
	return
}

package sialog

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// A siaerror is an error with context.
type siaerror struct {
	error
	ctxs []string
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
	return siaerror{
		errors.New(errString),
		[]string{fmt.Sprintf("%s: %s", fnName(), errString)},
	}
}

// Error returns a formatted siaerror, including the first level of context.
func Errorf(fmtString string, v ...interface{}) siaerror {
	return Error(fmt.Sprintf(fmtString, v...))
}

// AddCtx adds context to an error message. If the error is a siaerror, the
// context is added to ctxs. If it is a standard error, a new siaerror is
// created and returned.
func AddCtx(err error, ctx string) (se siaerror) {
	fullCtx := fmt.Sprint(fnName(), ": ", ctx)
	switch err.(type) {
	case error:
		se.error = err
		se.ctxs = []string{fullCtx}
	case siaerror:
		se = err.(siaerror)
		se.ctxs = append(se.ctxs, fullCtx)
	}
	return
}

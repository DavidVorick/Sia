package sialog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// These flags define the priority level of a log entry. Only one can be used
// at a time.
const (
	Ldebug = 1 << iota
	Ltrace
	Linfo
	Lwarn
	Lerror
	Lfatal

	// mask determines which priority levels are logged. This allows a debugger
	// to filter out irrelevant log entries. Note that by default, debug
	// and trace messages are excluded.
	mask = Linfo | Lwarn | Lerror | Lfatal
)

// These flags define the behavior of a log operation. Any number of them can be
const (
	Exit   = 1 << iota // terminate the program after logging this entry
	Stdout             // write this entry to stdout
	Stderr             // write this entry to stderr
	Trace              // append a stack trace to this entry
)

// A Logger is a logging object that can write to stdout, stderr, and/or a
// file. (Note that this makes it less flexible than the standard logger, which
// can write to any io.Writer.) It can safely be used from multiple goroutines.
type Logger struct {
	mu sync.Mutex
	w  io.Writer
}

// New returns a new Logger object capable of writing to the specified file.
func New(w io.Writer) *Logger {
	return &Logger{w: w}
}

// Default is the default logger.
var Default = New(os.Stderr)

// timestamp returns a properly formatted timestamp
func timestamp() string {
	timestamp := time.Now().Format("15:04:05.99")
	if len(timestamp) < 11 {
		timestamp += strings.Repeat("0", 11-len(timestamp))
	}
	return timestamp
}

// tag returns a properly formatted set of tags
func tag(priority uint32) string {
	return map[uint32]string{
		Ldebug: "DEBUG:",
		Ltrace: "TRACE:",
		Linfo:  "INFO: ",
		Lwarn:  "WARN: ",
		Lerror: "ERROR:",
		Lfatal: "FATAL:",
	}[priority]
}

// trace returns formatted output displaying the stack trace of the current
// call. sialog-specific calls are omitted.
func trace() string {
	var calls []string
	var prefixLen int
	for i := 1; ; i++ {
		// look up i-th caller
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			panic("stack trace failed")
		}

		// format variables
		pcname := runtime.FuncForPC(pc).Name()
		pkgproc := strings.SplitN(pcname[strings.LastIndex(pcname, "/")+1:], ".", 2)
		pkg, proc := pkgproc[0], pkgproc[1]
		file = file[strings.LastIndex(file, "/")+1:]

		// break if we've reached main or testing, skip if we're inside sialog
		if pkg == "runtime" && proc == "main" ||
			pkg == "testing" && proc == "tRunner" {
			break
		} else if pkg == "sialog" {
			continue
		}

		call := fmt.Sprintf("\t\t%s/%s:%d %s\n", pkg, file, line, proc)

		if pl := strings.Index(call, " "); pl > prefixLen {
			prefixLen = pl
		}

		calls = append(calls, call)
	}

	// pad as necessary to align columns
	for i := range calls {
		spaces := strings.Repeat(" ", prefixLen-strings.Index(calls[i], " ")+1)
		calls[i] = strings.Replace(calls[i], " ", spaces, -1)
	}

	return strings.Join(calls, "")
}

// Log is the most generic logging function. It allows the user to specify both
// the priority and flags of the logging operation. To print a formatted
// message, use fmt.Sprintf() to create the message string.
func (l *Logger) Log(msg string, priority, flags uint32) {
	// add timestamp and priority tag
	entry := fmt.Sprintf("[%s] %s %s", timestamp(), tag(priority), msg)
	// add newline if necessary
	if entry[len(entry)-1] != '\n' {
		entry += "\n"
	}

	// check that this priority level is unmasked
	if priority&mask == 0 {
		return
	}

	// if Trace flag is set, add trace output
	if flags&Trace != 0 {
		entry += "\tTrace:\n" + trace()
	}

	// print to each output
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.w.Write([]byte(entry)); err != nil {
		panic(err)
	}
	if flags&Stdout != 0 {
		if _, err := os.Stdout.WriteString(entry); err != nil {
			panic(err)
		}
	}
	if flags&Stderr != 0 && l != Default {
		if _, err := os.Stderr.WriteString(entry); err != nil {
			panic(err)
		}
	}

	// if Exit flag set, terminate program
	if flags&Exit != 0 {
		os.Exit(1)
	}
}

// Debug prints the log message with the DEBUG tag.
func (l *Logger) Debug(v ...interface{}) {
	l.Log(fmt.Sprintln(v...), Ldebug, 0)
}

// Trace prints the log message with the TRACE tag. It appends a trace.
func (l *Logger) Trace(v ...interface{}) {
	l.Log(fmt.Sprintln(v...), Ltrace, Trace)
}

// Info prints the log message with the INFO tag.
func (l *Logger) Info(v ...interface{}) {
	l.Log(fmt.Sprintln(v...), Linfo, 0)
}

// Warn prints the log message with the WARN tag.
func (l *Logger) Warn(v ...interface{}) {
	l.Log(fmt.Sprintln(v...), Lwarn, 0)
}

// Error prints the log message with the ERROR tag.
func (l *Logger) Error(v ...interface{}) {
	l.Log(fmt.Sprintln(v...), Lerror, 0)
}

// Fatal prints the log message with the FATAL tag. It appends a trace and
// calls os.Exit(1).
func (l *Logger) Fatal(v ...interface{}) {
	l.Log(fmt.Sprintln(v...), Lfatal, Trace|Exit)
}

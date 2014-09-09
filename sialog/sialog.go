package sialog

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// These flags define the priority level of a log entry. Only one can be used
// at a time.
const (
	Ldebug = "DEBUG:"
	Linfo  = "INFO: "
	Lwarn  = "WARN: "
	Lerror = "ERROR:"
	Lfatal = "FATAL:"

	// mask determines which priority levels are logged. This allows a debugger
	// to filter out irrelevant log entries.
	mask = Ldebug + Linfo + Lwarn + Lerror + Lfatal
)

// The flags define the behavior of a log operation. Any number of them can be
// OR'd together.
const (
	Exit   = 1 << iota // terminate program after logging this entry
	File               // write this entry to logFile
	Stdout             // write this entry to stdout
	Stderr             // write this entry to stderr
	Trace              // append a stack trace to this entry
)

// A Logger is a logging object that can write to stdout, stderr, and/or a
// file. (Note that this makes it less flexible than the standard logger, which
// can write to any io.Writer.) It can safely be used from multiple goroutines.
type Logger struct {
	writeLock sync.Mutex
	logFile   string
}

// New returns a new Logger object capable of writing to the specified file.
func New(logFile string) *Logger {
	return &Logger{logFile: logFile}
}

// timestamp returns a properly formatted timestamp
func timestamp() string {
	timestamp := time.Now().Format("15:04:05.99")
	if len(timestamp) < 11 {
		timestamp += strings.Repeat("0", 11-len(timestamp))
	}
	return timestamp
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
		if (pkg == "runtime" && proc == "main") || pkg == "testing" {
			break
		} else if pkg != "sialog" {
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
func (l *Logger) Log(msg string, priority string, flags uint32) {
	// add timestamp and priority tag
	entry := fmt.Sprintf("[%s] %s %s\n", timestamp(), priority, msg)

	// check that this priority level is unmasked
	if !strings.Contains(mask, priority) {
		return
	}

	if priority == Lfatal {
		fmt.Println(flags)
	}

	// if Trace flag is set, add trace output

	if flags&Trace != 0 {
		entry += "\tTrace:\n" + trace()
	}

	// print to each output
	l.writeLock.Lock()
	defer l.writeLock.Unlock()
	if flags&File != 0 {
		// create file if it doesn't already exist, and append to it
		f, err := os.OpenFile(l.logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
		if err != nil {
			panic(err)
		}
		if _, err := f.WriteString(entry); err != nil {
			panic(err)
		}
		f.Close()
	}
	if flags&Stdout != 0 {
		if _, err := os.Stdout.WriteString(entry); err != nil {
			panic(err)
		}
	}
	if flags&Stderr != 0 {
		if _, err := os.Stderr.WriteString(entry); err != nil {
			panic(err)
		}
	}

	// if Exit flag set, terminate program
	if flags&Exit != 0 {
		os.Exit(1)
	}
}

// Info prints the log message with the INFO tag.
func (l *Logger) Info(v ...interface{}) {
	l.Log(fmt.Sprint(v...), Linfo, File)
}

// Warn prints the log message with the WARN tag.
func (l *Logger) Warn(v ...interface{}) {
	l.Log(fmt.Sprint(v...), Lwarn, File)
}

// Error prints the log message with the ERROR tag.
func (l *Logger) Error(v ...interface{}) {
	l.Log(fmt.Sprint(v...), Lerror, File)
}

// Fatal prints the log message with a trace and calls os.Exit(1)
func (l *Logger) Fatal(v ...interface{}) {
	l.Log(fmt.Sprint(v...), Lfatal, File|Stderr|Trace|Exit)
}

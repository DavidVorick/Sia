package sialog

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// priority levels
const (
	Ldebug = "DEBUG:"
	Linfo  = "INFO: "
	Lwarn  = "WARN: "
	Lerror = "ERROR:"
	Lfatal = "FATAL:"

	// mask determines which log messages are printed
	mask = Ldebug + Linfo + Lwarn + Lerror + Lfatal
)

// behavior flags
const (
	Exit = 1 << iota
	File
	Stdout
	Stderr
	Trace
)

type Logger struct {
	sync.Mutex
	logFile string
}

func NewLogger(logFile string) *Logger {
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
func trace() (s string) {
	for i := 1; ; i++ {
		// look up i-th caller
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			panic("stack trace failed")
		}
		pcname := runtime.FuncForPC(pc).Name()
		pkg := strings.Split(pcname[strings.LastIndex(pcname, "/")+1:], ".")[0]
		proc := pcname[strings.LastIndex(pcname, ".")+1:]

		// break if we've reached main, and skip if we're inside sialog
		if pkg == "runtime" && proc == "main" {
			break
		} else if pkg == "testing" {
			break
		} else if pkg != "sialog" {
			continue
		}

		file = file[strings.LastIndex(file, "/")+1:]

		s += fmt.Sprintf("\t\t%s/%s:%d %s\n", pkg, file, line, proc)
	}
	return
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
	l.Lock()
	defer l.Unlock()
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
func (l *Logger) Info(msg string) {
	l.Log(msg, Linfo, File)
}

// Warn prints the log message with the WARN tag.
func (l *Logger) Warn(msg string) {
	l.Log(msg, Lwarn, File)
}

// Error prints the log message with the ERROR tag.
func (l *Logger) Error(msg string) {
	l.Log(msg, Lerror, File)
}

// Fatal prints the log message with a trace and calls os.Exit(1)
func (l *Logger) Fatal(msg string) {
	l.Log(msg, Lfatal, File|Stderr|Trace|Exit)
}

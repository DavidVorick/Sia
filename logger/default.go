package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

var std = NewPriorityLog(os.Stderr, PstdFlags, true)

// LogStored is a call to LogStored on the standard logger.
func LogStored() {
	std.LogStored()
}

// Clear is a call to Clear on the standard logger.
func Clear() {
	std.Clear()
}

// SetGlobalFlags is a call to SetGlobalFlags on the standard logger.
func SetGlobalFlags(newflags uint) {
	std.SetGlobalFlags(newflags)
}

// SetDeferedBehavior is a call to SetDeferedBehavior on the standard logger.
func SetDeferedBehavior(dispose bool) {
	std.SetDeferedBehavior(dispose)
}

// SetOutput is a call to SetOutput on the standard logger.
func SetOutput(out io.Writer) {
	std.SetOutput(out)
}

// Fatal is a call to Fatal on the standard logger.
func Fatal(v ...interface{}) {
	std.claim()
	std.log(time.Now(), Pfatal, fmt.Sprint(v...))
	std.unclaim()
	std.LogStored()
	os.Exit(1)
}

// Error is a call to Error on the standard logger.
func Error(v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Perror, fmt.Sprint(v...))
}

// Warning is a call to Warning on the standard logger.
func Warning(v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pwarning, fmt.Sprint(v...))
}

// Info is a call to Info on the standard logger.
func Info(v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pinfo, fmt.Sprint(v...))
}

// Debug is a call to Debug on the standard logger.
func Debug(v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pdebug, fmt.Sprint(v...))
}

// Fatalln is a call to Fatalln on the standard logger.
func Fatalln(v ...interface{}) {
	std.claim()
	std.log(time.Now(), Pfatal, fmt.Sprintln(v...))
	std.unclaim()
	std.LogStored()
	os.Exit(1)
}

// Errorln is a call to Errorln on the standard logger.
func Errorln(v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Perror, fmt.Sprintln(v...))
}

// Warningln is a call to Warningln on the standard logger.
func Warningln(v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pwarning, fmt.Sprintln(v...))
}

// Infoln is a call to Infoln on the standard logger.
func Infoln(v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pinfo, fmt.Sprintln(v...))
}

// Debugln is a call to Debugln on the standard logger.
func Debugln(v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pdebug, fmt.Sprintln(v...))
}

// Fatalf is a call to Fatalf on the standard logger.
func Fatalf(format string, v ...interface{}) {
	std.claim()
	std.log(time.Now(), Pfatal, fmt.Sprintf(format, v...))
	std.unclaim()
	std.LogStored()
	os.Exit(1)
}

// Errorf is a call to Errorf on the standard logger.
func Errorf(format string, v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Perror, fmt.Sprintf(format, v...))
}

// Warningf is a call to Warningf on the standard logger.
func Warningf(format string, v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pwarning, fmt.Sprintf(format, v...))
}

// Infof is a call to Infof on the standard logger.
func Infof(format string, v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pinfo, fmt.Sprintf(format, v...))
}

// Debugf is a call to Debugf on the standard logger.
func Debugf(format string, v ...interface{}) {
	std.claim()
	defer std.unclaim()
	std.log(time.Now(), Pdebug, fmt.Sprintf(format, v...))
}

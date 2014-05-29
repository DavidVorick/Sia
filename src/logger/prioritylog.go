// Package log implements a simple log with priority levels.
//
// There is a difference between a log request and logging for the purposes of
// this documentation. Logging is the writing of the entry to the Writer where
// the log is set to output to. A log request tells the log to handle a given
// message with an associated priority level. When handling an entry, if the
// priority level is set in the "now" flags of the logger, it is logged
// immediately. Otherwise, if the "dispose" flag of the logger is NOT set, the
// entry will be stored to be (potentially) logged on program termination.
// However, if the "dispose" flag is set, then the message will be discarded.
//
// The priority levels supported by this log are Fatal, Error, Warning, Info,
// and Debug. Aside from determining whether a given message is handled
// immediately, the priority level determines the associated "tag" with the
// message.
//
// Any message made at the Fatal priority level terminates the program
// immediately after making the log request, use it only with errors which
// cannot be recovered from
//
package logger

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// These are the priority level flags, it is advised to stick with the defaults.
const (
	Pfatal = 1 << iota
	Perror
	Pwarning
	Pinfo
	Pdebug
	PstdFlags = Pfatal | Perror | Pwarning
	// Use only in the full program for debug mode.
	PdebugFlags = Pfatal | Perror | Pwarning | Pinfo | Pdebug
)

// PriorityLog is a log with priority levels.
type PriorityLog struct {
	//
	dispose bool
	// now control the priority levels which are immediately logged and stored
	// to print in case of an error.
	now     uint
	records *recordHeap
	out     io.Writer
	// lock protects the above fields
	lock sync.Mutex
	// stop stops creation of new log entries.
	// THere are two mutexes solely to prevent deadlocks.
	stop sync.Mutex
	// Every logging function counts as two in pending.
	// The first for formatting the message, the second for creating the record.
	pending sync.WaitGroup
}

// NewPriorityLog creates a PriorityLog with given output Writer, which priority
// levels to handle immediately, and whether to dispose deferred entries.
func NewPriorityLog(out io.Writer, flags uint, dispose bool) *PriorityLog {
	pl := new(PriorityLog)
	pl.out = out
	pl.now = flags
	pl.dispose = dispose
	pl.records = new(recordHeap)
	return pl
}

func (pl *PriorityLog) claim() {
	// Must be done before calls to log
	// Lock is used to prevent one routine from continuously making entries
	pl.stop.Lock()
	defer pl.stop.Unlock()
	// One for the caller, one for log
	pl.pending.Add(2)
}

func (pl *PriorityLog) unclaim() {
	// No need to wait, finishing less harmful than starting
	pl.pending.Done()
}

func (pl *PriorityLog) log(now time.Time, priority uint, message string) {
	// Calling function called claim, so unclaim when done.
	defer pl.unclaim()
	// claim complete use over the variables of pl.
	pl.lock.Lock()
	defer pl.lock.Unlock()
	// Optimization!
	if pl.now&priority != 0 || !pl.dispose {
		pl.lock.Unlock()
		// Takes a while.
		// The calldepth will always be 3, if called directly.
		// Caller->helper->log->newRecord
		rec := newRecord(3, now, priority, message)
		pl.lock.Lock()
		if pl.now&priority != 0 {
			pl.out.Write(rec.Format())
		} else {
			heap.Push(pl.records, rec)
		}
	}
}

// All changes to the state of the log wait for all pending log entries
// to be logged/stored. In other words, avoid use except during
// initialization.

// SetGlobalFlags sets the the default flag for which priority levels are
// defered and which ones are handled immediately
func (pl *PriorityLog) SetGlobalFlags(newflags uint) {
	pl.stop.Lock()
	defer pl.stop.Unlock()
	pl.pending.Wait()
	pl.lock.Lock()
	defer pl.lock.Unlock()
	pl.now = newflags
}

// SetDeferedBehavior changes how defered log entries are handled.
// (i.e. disposed of or not)
func (pl *PriorityLog) SetDeferedBehavior(dispose bool) {
	pl.stop.Lock()
	defer pl.stop.Unlock()
	pl.pending.Wait()
	pl.lock.Lock()
	defer pl.lock.Unlock()
	pl.dispose = dispose
}

// SetOutput changes the Writer used to write log entries.
func (pl *PriorityLog) SetOutput(out io.Writer) {
	pl.stop.Lock()
	defer pl.stop.Unlock()
	pl.pending.Wait()
	pl.lock.Lock()
	defer pl.lock.Unlock()
	pl.out = out
}

// The following methods log a given message at a given priority level.
// There are three styles of formatting, Print, Println, and Printf.
// Details can be found at http://golang.org/pkg/fmt/

// The following methods use the Print formatting style.

// Log a message at the Fatal priority level, then terminate.
func (pl *PriorityLog) Fatal(v ...interface{}) {
	pl.claim()
	pl.log(time.Now(), Pfatal, fmt.Sprint(v...))
	pl.unclaim()
	pl.LogStored()
	os.Exit(1)
}

// Log a message at the Error priority level.
func (pl *PriorityLog) Error(v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Perror, fmt.Sprint(v...))
}

// Log a message at the Warning priority level.
func (pl *PriorityLog) Warning(v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pwarning, fmt.Sprint(v...))
}

// Log a message at the Info priority level.
func (pl *PriorityLog) Info(v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pinfo, fmt.Sprint(v...))
}

// Log a message at the Debug priority level.
func (pl *PriorityLog) Debug(v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pdebug, fmt.Sprint(v...))
}

// The following methods use the Println formatting style.

// Log a message at the Fatal priority level, then terminate.
func (pl *PriorityLog) Fatalln(v ...interface{}) {
	pl.claim()
	pl.log(time.Now(), Pfatal, fmt.Sprintln(v...))
	pl.unclaim()
	pl.LogStored()
	os.Exit(1)
}

// Log a message at the Error priority level.
func (pl *PriorityLog) Errorln(v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Perror, fmt.Sprintln(v...))
}

// Log a message at the Warning priority level.
func (pl *PriorityLog) Warningln(v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pwarning, fmt.Sprintln(v...))
}

// Log a message at the Info priority level.
func (pl *PriorityLog) Infoln(v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pinfo, fmt.Sprintln(v...))
}

// Log a message at the Debug priority level.
func (pl *PriorityLog) Debugln(v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pdebug, fmt.Sprintln(v...))
}

// The following methods use the Printf formatting style.

// Log a message at the Fatal priority level, then terminate.
func (pl *PriorityLog) Fatalf(format string, v ...interface{}) {
	pl.claim()
	pl.log(time.Now(), Pfatal, fmt.Sprintf(format, v...))
	pl.unclaim()
	pl.LogStored()
	os.Exit(1)
}

// Log a message at the Error priority level.
func (pl *PriorityLog) Errorf(format string, v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Perror, fmt.Sprintf(format, v...))
}

// Log a message at the Warning priority level.
func (pl *PriorityLog) Warningf(format string, v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pwarning, fmt.Sprintf(format, v...))
}

// Log a message at the Info priority level.
func (pl *PriorityLog) Infof(format string, v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pinfo, fmt.Sprintf(format, v...))
}

// Log a message at the Debug priority level.
func (pl *PriorityLog) Debugf(format string, v ...interface{}) {
	pl.claim()
	defer pl.unclaim()
	pl.log(time.Now(), Pdebug, fmt.Sprintf(format, v...))
}

// Log all deferred entries.
func (pl *PriorityLog) LogStored() {
	// Stop making new log entries
	pl.stop.Lock()
	defer pl.stop.Unlock()
	// Wait for any yet to be added log entries
	pl.pending.Wait()
	pl.lock.Lock()
	defer pl.lock.Unlock()
	// Write it all out
	for pl.records.Len() > 0 {
		rec := heap.Pop(pl.records).(*record)
		pl.out.Write(rec.Format())
	}
}

// Dispose of all deferred entries.
func (pl *PriorityLog) Clear() {
	// Uncertain if this should wait for all pending records or not.
	// For now, don't bother.
	pl.lock.Lock()
	defer pl.lock.Unlock()
	pl.records.Clear()
}

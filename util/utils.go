package util

import (
	"fmt"
	"log"
	"os"
	"time"
)

type PrintCallback func(format string, args ...interface{})

func init() {
	if PrintUtil == nil {
		InitPrinter(PrintErr)
	}
}

/*
 * Print messages to stderr
 */
func PrintErr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

/*
 * Print messages to log
 */
func PrintLog(format string, args ...interface{}) {
	log.Printf(format, args...)
}

/*
 * Discard messages silently.
 */
func Quiet(format string, args ...interface{}) {
	//discard message
}

var PrintUtil PrintCallback

func InitPrinter(callback PrintCallback) {
	PrintUtil = callback
}

//TimeTrack function for timing function calls. Usage:
// defer TimeTrack(time.Now()) at the beginning of the timed function
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	PrintUtil("%s took %s\n", name, elapsed)
}

//Exit type to handle exiting
type Exit struct{ Code int }

//HandleExit Looks at the panic for Exit codes vs actual panics
func HandleExit() {
	if e := recover(); e != nil {
		if exit, ok := e.(Exit); ok == true {
			os.Exit(exit.Code)
		}
		panic(e)
	}
}

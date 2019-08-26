package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"time"
)

type PrintCallback func(format string, args ...interface{})

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
var StdErr io.Writer
var StdOut io.Writer

func InitPrinter(callback PrintCallback, stderr, stdout io.Writer) {
	PrintUtil = callback
	StdErr = stderr
	StdOut = stdout
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

//CleanString takes a string and a list of args and returns the formatted string without excess whitespace
func CleanString(format string, args ...interface{}) string {
	temp := fmt.Sprintf(format, args)
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(temp, " ")
}

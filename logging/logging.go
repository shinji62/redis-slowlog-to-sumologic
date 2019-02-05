package logging

import (
	"io"
	"io/ioutil"
	"log"
)

// logging levels
const (
	LogTrace int = 1 << iota
	LogInfo
	LogWarning
	LogError
)

// logging interfaces
var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

// Init initialize the loggers with their buffers
func Init(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer,
	logLevel int) {

	// check if logger level is specified, if not, replace buffer with discard
	if logLevel&(LogTrace) == 0 {
		traceHandle = ioutil.Discard
	}
	if logLevel&(LogInfo) == 0 {
		infoHandle = ioutil.Discard
	}
	if logLevel&(LogWarning) == 0 {
		warningHandle = ioutil.Discard
	}
	if logLevel&(LogError) == 0 {
		errorHandle = ioutil.Discard
	}

	// initialize loggers
	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

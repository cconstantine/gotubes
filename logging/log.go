package logging

import (
  "io/ioutil"
  "os"
  "log"
)

var (
    Trace   *log.Logger
    Info    *log.Logger
    Error   *log.Logger
)

func Init(verbose bool) {
    trace_writer := ioutil.Discard
    if verbose {
      trace_writer = os.Stdout
    }
    Trace = log.New(trace_writer, "TRACE: ", log.Lshortfile)

    Info = log.New(os.Stdout, "INFO: ", 0)

    Error = log.New(os.Stderr, "ERROR: ", 0)
}


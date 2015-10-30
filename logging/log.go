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
    info_writer := ioutil.Discard
    if verbose {
      info_writer = os.Stdout
    }
    Trace = log.New(ioutil.Discard,"TRACE: ", log.Lshortfile)

    Info = log.New(info_writer, "INFO: ", 0)

    Error = log.New(os.Stderr, "ERROR: ", 0)
}


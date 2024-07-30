package main

import (
	"io"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/op/go-logging.v1"
)

// init exists here to silence the default logging behaviour of yqlib
func init() {
	logger := yqlib.GetLogger()
	if logger == nil {
		return
	}
	backend := logging.NewLogBackend(io.Discard, "", 0)
	levelled := logging.AddModuleLevel(backend)
	levelled.SetLevel(logging.ERROR, "")
	logger.SetBackend(levelled)
}

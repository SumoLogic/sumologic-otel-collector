package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func helpAction(_ *pflag.Flag, fs *pflag.FlagSet) error {
	fmt.Printf("%s: configure otelcol-sumo\n", os.Args[0])
	fs.PrintDefaults()
	return nil
}

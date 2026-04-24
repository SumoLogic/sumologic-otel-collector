//go:build goexperiment.fips140

package main

import (
	"crypto/fips140"
	"log"
)

func init() {
	if fips140.Enabled() {
		log.Print("Running in FIPS 140-3 mode")
	}
}

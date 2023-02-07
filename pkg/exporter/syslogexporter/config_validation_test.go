package syslogexporter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	validator "gopkg.in/go-playground/validator.v9"
)

var protocolInput = []struct {
	protocol string
	result   bool
}{
	{"tcp", true},
	{"udp", true},
	{"ftp", false},
}

var portInput = []struct {
	port   int
	result bool
}{
	{514, true},
}

var endpointInput = []struct {
	endpoint string
	result   bool
}{
	{"test.com", true},
	{"test", false},
}

var formatInput = []struct {
	format string
	result bool
}{
	{"any", true},
	{"RFC5424", true},
	{"RFC3164", true},
	{"all", false},
}

func TestValidateProtocol(t *testing.T) {
	validate := validator.New()
	var merr []error
	err := validate.RegisterValidation("protocol type", protocols)
	merr = append(merr, err)
	fmt.Println(merr)
	for _, item := range protocolInput {
		err := validate.Var(item.protocol, "protocol type")
		if item.result {
			assert.Nil(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestValidatePort(t *testing.T) {
	validate := validator.New()
	var merr []error
	err := validate.RegisterValidation("port", numericPort)
	merr = append(merr, err)
	fmt.Println(merr)
	for _, item := range portInput {
		err := validate.Var(item.port, "port")
		if item.result {
			assert.Nil(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestValidateformat(t *testing.T) {
	validate := validator.New()
	var merr []error
	err := validate.RegisterValidation("format", format)
	merr = append(merr, err)
	fmt.Println(merr)
	for _, item := range formatInput {
		err := validate.Var(item.format, "format")
		if item.result {
			assert.Nil(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestValidateEndpoint(t *testing.T) {
	validate := validator.New()
	for _, item := range endpointInput {
		err := validate.Var(item.endpoint, "fqdn")
		if item.result {
			assert.Nil(t, err)
		} else {
			assert.Error(t, err)
		}
	}
}

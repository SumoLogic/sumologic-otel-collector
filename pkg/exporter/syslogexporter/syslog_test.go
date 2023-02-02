package syslogexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddStructuredData(t *testing.T) {
	s := Syslog{
		additionalStructuredData: []string{"9HFxoa6+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo+w4H7PmZm8H3mSEKxPl0Q@41123", "tab=abc"},
	}

	msg := s.addStructuredData("<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts.")
	expected := "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - [9HFxoa6+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo+w4H7PmZm8H3mSEKxPl0Q@41123 tab=abc] %% It's time to make the do-nuts."
	assert.Equal(t, expected, msg)

	msg = s.addStructuredData("<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - [app=test cpu=CPU1] %% It's time to make the do-nuts.")
	expected = "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - [9HFxoa6+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo+w4H7PmZm8H3mSEKxPl0Q@41123 tab=abc app=test cpu=CPU1] %% It's time to make the do-nuts."
	assert.Equal(t, expected, msg)
}

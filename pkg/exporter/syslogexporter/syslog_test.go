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

func TestIsRFC5424(t *testing.T) {
	msg := "<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - BOM'su root' failed for lonvick on /dev/pts/8"
	assert.True(t, isRFC5424(msg))

	msg = "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the do-nuts."
	assert.True(t, isRFC5424(msg))

	msg = "<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut=\"3\" eventSource=\"Application\" eventID=\"1011\"] BOMAn application event log entry..."
	assert.True(t, isRFC5424(msg))

	msg = "<165>1 2003-10-11T22:14:15.003Z"
	assert.False(t, isRFC5424(msg))

	msg = "message"
	assert.False(t, isRFC5424(msg))
}

func TestRFC3164(t *testing.T) {
	msg := "<34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8"
	assert.True(t, isRFC3164(msg))

	msg = "<13>Feb  5 17:32:18 10.0.0.99 Use the BFG!"
	assert.True(t, isRFC3164(msg))

	msg = "<13>Feb  5 17:32:18"
	assert.False(t, isRFC3164(msg))

	msg = "message"
	assert.False(t, isRFC3164(msg))
}

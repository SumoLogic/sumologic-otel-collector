package syslogexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddStructuredData(t *testing.T) {
	s := Syslog{
		additionalStructuredData: []string{"9HFxoa6+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo+w4H7PmZm8H3mSEKxPl0Q@41123"},
	}
	msg2 := map[string]any{"timestamp": "2003-08-24T05:14:15.000003-07:00", "appname": "myproc", "facility": 20,
		"hostname": "192.0.2.1", "log.file.name": "syslog", "message": "It's time to make the do-nuts.",
		"priority": 165, "proc_id": "8710", "version": 1}
	s.addStructuredData(msg2)
	expected := "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - " +
		"[9HFxoa6+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo+w4H7PmZm8H3mSEKxPl0Q@41123] It's time to make the do-nuts."
	assert.Equal(t, expected, formatRFC5424(msg2))
}

func TestNoAddStructuredData(t *testing.T) {
	s := Syslog{}
	msg2 := map[string]any{"timestamp": "2003-08-24T05:14:15.000003-07:00", "appname": "myproc", "facility": 20,
		"hostname": "192.0.2.1", "log.file.name": "syslog", "message": "It's time to make the do-nuts.", "priority": 165,
		"proc_id": "8710", "version": 1}
	s.addStructuredData(msg2)
	expected := "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - It's time to make the do-nuts."
	assert.Equal(t, expected, formatRFC5424(msg2))
}

func TestIsRFC5424(t *testing.T) {
	msg1 := map[string]any{"timestamp": "2003-10-11T22:14:15.003Z", "appname": "evntslog", "facility": 20,
		"hostname": "mymachine.example.com", "log.file.name": "syslog", "message": "BOMAn application event log entry...",
		"msg_id": "ID47", "priority": 165, "proc_id": "111", "version": 1}
	msgExpected := "<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog 111 ID47 - BOMAn application event log entry..."
	assert.Equal(t, msgExpected, formatRFC5424(msg1))

	msg2 := map[string]any{"timestamp": "2003-08-24T05:14:15.000003-07:00", "appname": "myproc",
		"facility": 20, "hostname": "192.0.2.1", "log.file.name": "syslog", "message": "It's time to make the do-nuts.",
		"priority": 165, "proc_id": "8710", "version": 1}
	msgExpected = "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - It's time to make the do-nuts."
	assert.Equal(t, msgExpected, formatRFC5424(msg2))

	s := Syslog{
		additionalStructuredData: []string{"exampleSDID@32473"},
	}
	msgExpected = "<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473] " +
		"BOMAn application event log entry..."
	msg3 := map[string]any{"timestamp": "2003-10-11T22:14:15.003Z", "appname": "evntslog", "facility": 20,
		"hostname": "mymachine.example.com", "log.file.name": "syslog", "message": "BOMAn application event log entry...",
		"msg_id": "ID47", "priority": 165, "version": 1}
	s.addStructuredData(msg3)
	assert.Equal(t, msgExpected, formatRFC5424(msg3))
}

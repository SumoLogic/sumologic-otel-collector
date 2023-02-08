package syslogexporter

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatRFC5424(t *testing.T) {
	s := Syslog{
		format:                   formatRFC5424Str,
		additionalStructuredData: []string{"9HFxoa6+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo+w4H7PmZm8H3mSEKxPl0Q@41123"},
	}
	msg := map[string]any{"timestamp": "2003-08-24T05:14:15.000003-07:00", "appname": "myproc", "facility": 20,
		"hostname": "192.0.2.1", "log.file.name": "syslog", "message": "It's time to make the do-nuts.",
		"priority": 165, "proc_id": "8710", "version": 1}
	expected := "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - " +
		"[9HFxoa6+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo+w4H7PmZm8H3mSEKxPl0Q@41123] It's time to make the do-nuts."
	assert.Equal(t, expected, s.formatRFC5424(msg))

	s2 := Syslog{
		format:                   formatRFC5424Str,
		additionalStructuredData: []string{"exampleSDID@32473"},
	}
	msg2 := map[string]any{"timestamp": "2003-10-11T22:14:15.003Z", "appname": "evntslog", "facility": 20,
		"hostname": "mymachine.example.com", "log.file.name": "syslog", "message": "BOMAn application event log entry...",
		"msg_id": "ID47", "priority": 165, "version": 1}
	expected2 := "<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473] " +
		"BOMAn application event log entry..."
	assert.Equal(t, expected2, s2.formatRFC5424(msg2))

	s3 := Syslog{
		format:                   formatRFC5424Str,
		additionalStructuredData: []string{"9HFxoa6+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo+w4H7PmZm8H3mSEKxPl0Q@41123"},
	}
	msg3 := map[string]any{"timestamp": "2003-08-24T05:14:15.000003-07:00", "appname": "myproc", "facility": 20,
		"hostname": "192.0.2.1", "log.file.name": "syslog", "message": "It's time to make the do-nuts.",
		"priority": 165, "proc_id": "8710", "version": 1,
		"structured_data": map[string]map[string]string{"SecureAuth@27389": {
			"PEN":             "27389",
			"Realm":           "SecureAuth0",
			"UserHostAddress": "192.168.2.132",
			"UserID":          "Tester2",
		},
		},
	}

	expectedForm := "\\<165\\>1 2003-08-24T05:14:15\\.000003-07:00 192\\.0\\.2\\.1 myproc 8710 - \\[9HFxoa6\\+lXBmvSM9koPjGzvTaxXDQvJ4POE/WCURPAo\\+w4H7PmZm8H3mSEKxPl0Q@41123 \\S+ \\S+ \\S+ \\S+ \\S+\\] It's time to make the do-nuts\\."
	formattedMsg := s3.formatRFC5424(msg3)
	matched, err := regexp.MatchString(expectedForm, formattedMsg)
	assert.Nil(t, err)
	assert.Equal(t, true, matched, fmt.Sprintf("unexpected form of formatted message, formatted message: %s, regexp: %s", formattedMsg, expectedForm))
	assert.Equal(t, true, strings.Contains(formattedMsg, "Realm=\"SecureAuth0\""))
	assert.Equal(t, true, strings.Contains(formattedMsg, "UserHostAddress=\"192.168.2.132\""))
	assert.Equal(t, true, strings.Contains(formattedMsg, "UserID=\"Tester2\""))
	assert.Equal(t, true, strings.Contains(formattedMsg, "PEN=\"27389\""))
}

func TestFormatRFC5424NoAddStructuredData(t *testing.T) {
	s := Syslog{format: formatRFC5424Str}

	msg := map[string]any{"timestamp": "2003-08-24T05:14:15.000003-07:00", "appname": "myproc", "facility": 20,
		"hostname": "192.0.2.1", "log.file.name": "syslog", "message": "It's time to make the do-nuts.", "priority": 165,
		"proc_id": "8710", "version": 1}
	expected := "<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - It's time to make the do-nuts."
	assert.Equal(t, expected, s.formatRFC5424(msg))

	msg2 := map[string]any{"timestamp": "2003-10-11T22:14:15.003Z", "appname": "evntslog", "facility": 20,
		"hostname": "mymachine.example.com", "log.file.name": "syslog", "message": "BOMAn application event log entry...",
		"msg_id": "ID47", "priority": 165, "proc_id": "111", "version": 1}
	expected2 := "<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog 111 ID47 - BOMAn application event log entry..."
	assert.Equal(t, expected2, s.formatRFC5424(msg2))
}

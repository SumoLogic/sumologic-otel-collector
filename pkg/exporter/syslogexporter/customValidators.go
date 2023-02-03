package syslogexporter

import (
	"strconv"

	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
)

func numeric_port(fl validator.FieldLevel) bool {
	port := fl.Field().String()
	_, err := strconv.Atoi(port)
	if err != nil {
		return true
	}
	return false
}

func numeric_port_translator(ut ut.Translator) error {
	return ut.Add("numeric_port", "Invalid port, {0} must be a number", true) // see universal-translator for details
}

func numeric_port_validator(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("numeric_port", fe.Field())
	return t
}

func protocols(fl validator.FieldLevel) bool {
	protocol := fl.Field().String()
	if protocol == "udp" || protocol == "tcp" {
		return true
	}
	return false
}

func protocols_translator(ut ut.Translator) error {
	return ut.Add("protocols", "Invalid protocol, {0} must be tcp/udp", true) // see universal-translator for details
}

func protocols_validator(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("protocols", fe.Field())
	return t
}

func format(fl validator.FieldLevel) bool {
	format := fl.Field().String()
	if format == formatRFC5424 || format == formatRFC3164 || format == formatAny {
		return true
	}
	return false
}

func format_translator(ut ut.Translator) error {
	return ut.Add("format", "Invalid format, {0} must be any/RFC5424/RFC3164", true) // see universal-translator for details
}

func format_validator(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("format", fe.Field())
	return t
}

func required_translator(ut ut.Translator) error {
	return ut.Add("required", "{0} is a required field", true) // see universal-translator for details
}

func required_validator(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("required", fe.Field())
	return t
}

func fqdn_translator(ut ut.Translator) error {
	return ut.Add("fqdn", "Invalid endpoint, {0} must be a valid FQDN", true) // see universal-translator for details
}

func fqdn_validator(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("fqdn", fe.Field())
	return t
}

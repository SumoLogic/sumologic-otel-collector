package syslogexporter

import (
	"strconv"

	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
)

func numericPort(fl validator.FieldLevel) bool {
	port := fl.Field().String()
	_, err := strconv.Atoi(port)
	return err != nil
}

func numericPortTranslator(ut ut.Translator) error {
	return ut.Add("numeric_port", "Invalid port, {0} must be a number", true) // see universal-translator for details
}

func numericPortValidator(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T("numeric_port", fe.Field())
	if err != nil {
		return ""
	}
	return t
}

func protocols(fl validator.FieldLevel) bool {
	protocol := fl.Field().String()
	if protocol == "udp" || protocol == "tcp" {
		return true
	}
	return false
}

func protocolsTranslator(ut ut.Translator) error {
	return ut.Add("protocols", "Invalid protocol, {0} must be tcp/udp", true) // see universal-translator for details
}

func protocolsValidator(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T("protocols", fe.Field())
	if err != nil {
		return ""
	}
	return t
}

func format(fl validator.FieldLevel) bool {
	format := fl.Field().String()
	if format == formatRFC5424Str || format == formatRFC3164Str {
		return true
	}
	return false
}

func formatTranslator(ut ut.Translator) error {
	return ut.Add("format", "Invalid format, {0} must be any/rfc5424/rfc3164", true) // see universal-translator for details
}

func formatValidator(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T("format", fe.Field())
	if err != nil {
		return ""
	}
	return t
}

func requiredTranslator(ut ut.Translator) error {
	return ut.Add("required", "{0} is a required field", true) // see universal-translator for details
}

func requiredValidator(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T("required", fe.Field())
	if err != nil {
		return ""
	}
	return t
}

func fqdnTranslator(ut ut.Translator) error {
	return ut.Add("fqdn", "Invalid endpoint, {0} must be a valid FQDN", true) // see universal-translator for details
}

func fqdnValidator(ut ut.Translator, fe validator.FieldError) string {
	t, err := ut.T("fqdn", fe.Field())
	if err != nil {
		return ""
	}
	return t
}

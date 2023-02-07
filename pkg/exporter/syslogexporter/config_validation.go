package syslogexporter

import (
	"errors"
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

func getTranslator() (ut.Translator, bool) {
	translator := en.New()
	uni := ut.New(translator, translator)
	trans, found := uni.GetTranslator("en")
	return trans, found
}

func protocolsValidation(trans ut.Translator, validate validator.Validate) []error {
	var merr []error
	err1 := validate.RegisterTranslation("protocol type", trans, protocolsTranslator, protocolsValidator)
	merr = append(merr, err1)
	err2 := validate.RegisterValidation("protocol type", protocols)
	merr = append(merr, err2)
	return merr
}

func portValidation(trans ut.Translator, validate validator.Validate) []error {
	var merr []error
	err1 := validate.RegisterTranslation("port", trans, numericPortTranslator, numericPortValidator)
	merr = append(merr, err1)
	err2 := validate.RegisterValidation("port", numericPort)
	merr = append(merr, err2)
	return merr
}

func formatValidation(trans ut.Translator, validate validator.Validate) []error {
	var merr []error
	err1 := validate.RegisterTranslation("format", trans, formatTranslator, formatValidator)
	merr = append(merr, err1)
	err2 := validate.RegisterValidation("format", format)
	merr = append(merr, err2)
	return merr
}

func validation(cfg *Config) (error, ut.Translator) {
	var merr []error
	validate := validator.New()
	trans, found := getTranslator()
	if !found {
		return errors.New("Unsupported translator"), nil
	}
	err1 := en_translations.RegisterDefaultTranslations(validate, trans)
	err2 := validate.RegisterTranslation("required", trans, requiredTranslator, requiredValidator)
	err3 := validate.RegisterTranslation("fqdn", trans, fqdnTranslator, fqdnValidator)
	merr = append(merr, err1)
	merr = append(merr, err2)
	merr = append(merr, err3)
	fmt.Println(merr)
	protocolsValidation(trans, *validate)
	portValidation(trans, *validate)
	formatValidation(trans, *validate)
	err := validate.Struct(cfg)
	return err, trans
}

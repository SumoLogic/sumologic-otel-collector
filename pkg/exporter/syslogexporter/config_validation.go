package syslogexporter

import (
	"errors"

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

func protocolsValidation(trans ut.Translator, validate validator.Validate) {
	validate.RegisterTranslation("protocol type", trans, protocolsTranslator, protocolsValidator)
	validate.RegisterValidation("protocol type", protocols)
}

func portValidation(trans ut.Translator, validate validator.Validate) {
	validate.RegisterTranslation("port", trans, numericPortTranslator, numericPortValidator)
	validate.RegisterValidation("port", numericPort)
}

func formatValidation(trans ut.Translator, validate validator.Validate) {
	validate.RegisterTranslation("format", trans, formatTranslator, formatValidator)
	validate.RegisterValidation("format", format)
}

func validation(cfg *Config) (error, ut.Translator) {
	validate := validator.New()
	trans, found := getTranslator()
	if !found {
		return errors.New("Unsupported translator"), nil
	}
	en_translations.RegisterDefaultTranslations(validate, trans)
	validate.RegisterTranslation("required", trans, requiredTranslator, requiredValidator)
	validate.RegisterTranslation("fqdn", trans, fqdnTranslator, fqdnValidator)
	protocolsValidation(trans, *validate)
	portValidation(trans, *validate)
	formatValidation(trans, *validate)
	err := validate.Struct(cfg)
	return err, trans
}

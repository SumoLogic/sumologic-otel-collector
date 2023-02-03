package syslogexporter

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

func validation(cfg *Config) (error, ut.Translator) {
	translator := en.New()
	validate := validator.New()
	uni := ut.New(translator, translator)
	trans, found := uni.GetTranslator("en")
	if !found {
		return errors.New("Unsupported translator"), nil
	}
	en_translations.RegisterDefaultTranslations(validate, trans)
	validate.RegisterTranslation("required", trans, required_translator, required_validator)
	validate.RegisterTranslation("fqdn", trans, fqdn_translator, fqdn_validator)
	validate.RegisterTranslation("port", trans, numeric_port_translator, numeric_port_validator)
	validate.RegisterValidation("port", numeric_port)
	validate.RegisterTranslation("protocol type", trans, protocols_translator, protocols_validator)
	validate.RegisterValidation("protocol type", protocols)
	validate.RegisterTranslation("format", trans, format_translator, format_validator)
	validate.RegisterValidation("format", format)
	err := validate.Struct(cfg)
	return err, trans
}

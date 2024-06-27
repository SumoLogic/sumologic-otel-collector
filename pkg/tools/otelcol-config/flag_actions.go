package main

import (
	"errors"

	"github.com/spf13/pflag"
)

var notImplementedError = errors.New("not implemented")

func notImplementedAction(*pflag.Flag, *pflag.FlagSet) error {
	return notImplementedError
}

type action func(actionFlag *pflag.Flag, allFlags *pflag.FlagSet) error

var flagActions = map[string]action{
	flagHelp:                 helpAction,
	flagAddTag:               notImplementedAction,
	flagDeleteTag:            notImplementedAction,
	flagSetInstallationToken: notImplementedAction,
	flagEnableHostmetrics:    notImplementedAction,
	flagDisableHostmetrics:   notImplementedAction,
	flagEnableEphemeral:      notImplementedAction,
	flagDisableEphemeral:     notImplementedAction,
	flagSetAPIURL:            notImplementedAction,
	flagEnableRemoteControl:  notImplementedAction,
	flagDisableRemoteControl: notImplementedAction,
	flagSetOpAmpEndpoint:     notImplementedAction,
	flagConfigDir:            notImplementedAction,
	flagSetKV:                notImplementedAction,
	flagDelKV:                notImplementedAction,
	flagGetKV:                notImplementedAction,
	flagAppendKV:             notImplementedAction,
}

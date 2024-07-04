package main

import (
	"errors"
)

var notImplementedError = errors.New("not implemented")

func notImplementedAction(*actionContext) error {
	return notImplementedError
}

type action func(context *actionContext) error

var flagActions = map[string]action{
	flagHelp:                 helpAction,
	flagConfigDir:            nullAction,
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
	flagWriteKV:              WriteKVAction,
	flagReadKV:               ReadKVAction,
	flagOverride:             nullAction,
}

func nullAction(*actionContext) error {
	return nil
}

var actionOrder = []string{
	flagHelp,
	flagConfigDir,
	flagAddTag,
	flagDeleteTag,
	flagSetInstallationToken,
	flagEnableHostmetrics,
	flagDisableHostmetrics,
	flagEnableEphemeral,
	flagDisableEphemeral,
	flagSetAPIURL,
	flagEnableRemoteControl,
	flagDisableRemoteControl,
	flagSetOpAmpEndpoint,
	flagWriteKV,
	flagReadKV,
}

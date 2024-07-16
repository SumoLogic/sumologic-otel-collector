package main

import (
	"errors"
)

var errNotImplemented = errors.New("not implemented")

func notImplementedAction(*actionContext) error {
	return errNotImplemented
}

type action func(context *actionContext) error

var flagActions = map[string]action{
	flagHelp:                 helpAction,
	flagConfigDir:            nullAction,
	flagAddTag:               notImplementedAction,
	flagDeleteTag:            notImplementedAction,
	flagSetInstallationToken: SetInstallationTokenAction,
	flagEnableHostmetrics:    notImplementedAction,
	flagDisableHostmetrics:   notImplementedAction,
	flagEnableEphemeral:      notImplementedAction,
	flagDisableEphemeral:     notImplementedAction,
	flagSetAPIURL:            notImplementedAction,
	flagEnableRemoteControl:  EnableRemoteControlAction,
	flagDisableRemoteControl: DisableRemoteControlAction,
	flagSetOpAmpEndpoint:     SetOpAmpEndpointAction,
	flagWriteKV:              WriteKVAction,
	flagReadKV:               ReadKVAction,
	flagOverride:             nullAction,
}

func nullAction(*actionContext) error {
	return nil
}

// actionOrder specifies the order in which actions will be applied
var actionOrder = []string{
	flagHelp,
	flagConfigDir,
	flagEnableRemoteControl,
	flagDisableRemoteControl,
	flagAddTag,
	flagDeleteTag,
	flagSetInstallationToken,
	flagEnableHostmetrics,
	flagDisableHostmetrics,
	flagEnableEphemeral,
	flagDisableEphemeral,
	flagSetAPIURL,
	flagSetOpAmpEndpoint,
	flagWriteKV,
	flagReadKV,
}

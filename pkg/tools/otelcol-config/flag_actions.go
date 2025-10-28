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
	flagLaunchdDir:           nullAction,
	flagAddTag:               AddTagAction,
	flagDeleteTag:            DeleteTagAction,
	flagSetInstallationToken: SetInstallationTokenAction,
	flagEnableHostmetrics:    EnableHostmetricsAction,
	flagDisableHostmetrics:   DisableHostmetricsAction,
	flagEnableEphemeral:      EnableEphemeralAction,
	flagDisableEphemeral:     DisableEphemeralAction,
	flagSetTimezone:          SetTimezoneAction,
	flagSetAPIURL:            SetAPIURLAction,
	flagEnableRemoteControl:  EnableRemoteControlAction,
	flagDisableRemoteControl: DisableRemoteControlAction,
	flagSetOpAmpEndpoint:     SetOpAmpEndpointAction,
	flagWriteKV:              WriteKVAction,
	flagReadKV:               ReadKVAction,
	flagOverride:             nullAction,
	flagEnableClobber:        EnableClobberAction,
	flagDisableClobber:       DisableClobberAction,
}

func nullAction(*actionContext) error {
	return nil
}

// actionOrder specifies the order in which actions will be applied.
// This order has been chosen specifically so that actions do not conflict
// with one another. Use care when adding to this list or reordering it.
var actionOrder = []string{
	flagHelp,
	flagConfigDir,
	flagLaunchdDir,
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
	flagSetTimezone,
	flagWriteKV,
	flagReadKV,
	flagEnableClobber,
	flagEnableClobber,
}

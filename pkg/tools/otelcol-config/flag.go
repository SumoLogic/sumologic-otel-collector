package main

import (
	"os"

	"github.com/spf13/pflag"
)

const (
	flagHelp                 = "help"
	flagAddTag               = "add-tag"
	flagDeleteTag            = "delete-tag"
	flagSetInstallationToken = "set-installation-token"
	flagSetOpAmpEndpoint     = "set-opamp-endpoint"
	flagSetAPIURL            = "set-api-url"
	flagEnableHostmetrics    = "enable-hostmetrics"
	flagDisableHostmetrics   = "disable-hostmetrics"
	flagEnableEphemeral      = "enable-ephemeral"
	flagDisableEphemeral     = "disable-ephemeral"
	flagEnableRemoteControl  = "enable-remote-control"
	flagDisableRemoteControl = "disable-remote-control"
	flagConfigDir            = "config"
	flagSetKV                = "set-kv"
	flagDelKV                = "del-kv"
	flagGetKV                = "get-kv"
	flagAppendKV             = "append-kv"
)

const (
	addTagUsage               = "adds tags (eg: '--add-tag foo=bar --add-tag bar=baz' adds foo and bar tags)"
	delTagUsage               = "deletes tags (eg: '--delete-tag foo bar' deletes tags foo and bar)"
	sitUsage                  = "sets the sumo logic installation token"
	enableHMUsage             = "enables hostmetrics"
	disableHMUsage            = "disables hostmetrics"
	enableEphemeralUsage      = "enables ephemeral mode"
	disableEphemeralUsage     = "disables ephemeral mode"
	enableRemoteControlUsage  = "enables remote control via opamp"
	disableRemoteControlUsage = "disables remote control, uses local configuration only"
	setOpAmpEndpointUsage     = "sets the opamp endpoint (eg: wss://example.com)"
	configUsage               = "path to sumologic.yaml configuration file"
	setKVUsage                = "set key-value in sumologic.yaml with yq path (eg: --set-kv extensions.sumologic.foo=bar)"
	getKVUsage                = "get key-value from sumologic.yaml with yq path (eg: --get-kv extensions.sumologic.foo)"
	delKVUsage                = "delete key-value from sumologic.yaml with yq path (eg: --del-kv foo.bar)"
	appendKVUsage             = `append key-value to sumologic.yaml with yq path (eg: --append-kv 'extensions={"foo":"bar"}')`
)

type flagValues struct {
	AddTags              map[string]string
	DeleteTags           []string
	InstallationToken    string
	EnableHostmetrics    bool
	DisableHostmetrics   bool
	EnableEphemeral      bool
	DisableEphemeral     bool
	EnableRemoteControl  bool
	DisableRemoteControl bool
	SetOpAmpEndpoint     string
	Help                 bool
	ConfigDir            string
	SetKV                map[string]string
	DelKV                []string
	GetKV                []string
	AppendKV             map[string]string
}

func newFlagValues() *flagValues {
	return &flagValues{AddTags: make(map[string]string)}
}

func makeFlagSet(fv *flagValues) *pflag.FlagSet {
	flags := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)

	flags.SortFlags = true

	flags.StringToStringVarP(&fv.AddTags, flagAddTag, "a", nil, addTagUsage)
	flags.StringArrayVarP(&fv.DeleteTags, flagDeleteTag, "d", nil, delTagUsage)
	flags.StringVarP(&fv.InstallationToken, flagSetInstallationToken, "t", "", sitUsage)
	flags.BoolVar(&fv.EnableHostmetrics, flagEnableHostmetrics, false, enableHMUsage)
	flags.BoolVar(&fv.DisableHostmetrics, flagDisableHostmetrics, false, disableHMUsage)
	flags.BoolVar(&fv.EnableEphemeral, flagEnableEphemeral, false, enableEphemeralUsage)
	flags.BoolVar(&fv.DisableEphemeral, flagDisableEphemeral, false, disableEphemeralUsage)
	flags.BoolVar(&fv.EnableRemoteControl, flagEnableRemoteControl, false, enableRemoteControlUsage)
	flags.BoolVar(&fv.DisableRemoteControl, flagDisableRemoteControl, false, disableRemoteControlUsage)
	flags.StringVarP(&fv.SetOpAmpEndpoint, flagSetOpAmpEndpoint, "e", "", setOpAmpEndpointUsage)
	flags.StringVarP(&fv.ConfigDir, flagConfigDir, "c", "", configUsage)
	flags.StringToStringVar(&fv.SetKV, flagSetKV, nil, setKVUsage)
	flags.StringArrayVar(&fv.GetKV, flagGetKV, nil, getKVUsage)
	flags.StringArrayVar(&fv.DelKV, flagDelKV, nil, delKVUsage)
	flags.StringToStringVar(&fv.AppendKV, flagAppendKV, nil, appendKVUsage)

	return flags
}

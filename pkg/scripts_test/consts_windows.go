//go:build windows

package sumologic_scripts_tests

import "golang.org/x/sys/windows"

const (
	// See: https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/manage/understand-security-identifiers
	localSystemSID string = "S-1-5-18"

	packageName string = "OpenTelemetry Collector"

	binaryPath            string = `C:\Program Files\Sumo Logic\OpenTelemetry Collector\bin\otelcol-sumo.exe`
	libPath               string = `C:\ProgramData\Sumo Logic\OpenTelemetry Collector\data`
	fileStoragePath       string = libPath + `\file_storage`
	etcPath               string = `C:\ProgramData\Sumo Logic\OpenTelemetry Collector\config`
	scriptPath            string = "../../scripts/install.ps1"
	configPath                   = etcPath + `\sumologic.yaml`
	confDPath                    = etcPath + `\conf.d`
	opampDPath                   = etcPath + `\opamp.d`
	userConfigPath               = confDPath + `\common.yaml`
	hostmetricsConfigPath        = confDPath + `\hostmetrics.yaml`

	installToken    string = "token"
	installTokenEnv string = "SUMOLOGIC_INSTALLATION_TOKEN"
	apiBaseURL      string = "https://open-collectors.sumologic.com"

	allFilePermissions = windows.STANDARD_RIGHTS_ALL | windows.FILE_GENERIC_READ | windows.FILE_GENERIC_WRITE | windows.FILE_GENERIC_EXECUTE | 0x00000040
)

var (
	pathPermissions = []ACLRecord{
		{
			SID:               localSystemSID,
			AccessPermissions: allFilePermissions,
			AccessMode:        windows.GRANT_ACCESS,
		},
	}
	filePermissions           = []ACLRecord{}
	opampDPermissions         = pathPermissions
	configPathDirPermissions  = pathPermissions
	configPathFilePermissions = filePermissions
)

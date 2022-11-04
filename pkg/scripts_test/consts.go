package sumologic_scripts_tests

const (
	binaryPath            string = "/usr/local/bin/otelcol-sumo"
	libPath               string = "/var/lib/otelcol-sumo"
	fileStoragePath       string = libPath + "/file_storage"
	etcPath               string = "/etc/otelcol-sumo"
	systemdPath           string = "/etc/systemd/system/otelcol-sumo.service"
	scriptPath            string = "../../scripts/install.sh"
	configPath            string = etcPath + "/sumologic.yaml"
	configPathPermissions uint32 = 0440
	confDPath             string = etcPath + "/conf.d"
	userConfigPath        string = confDPath + "/common.yaml"

	systemdDirectoryPath string = "/run/systemd/system"

	installToken string = "token"
	apiBaseURL   string = "https://open-collectors.sumologic.com"

	systemUser string = "otelcol-sumo"
)

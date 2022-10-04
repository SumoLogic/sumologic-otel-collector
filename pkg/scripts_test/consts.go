package sumologic_scripts_tests

const (
	binaryPath      string = "/usr/local/bin/otelcol-sumo"
	fileStoragePath string = "/var/lib/sumologic/file_storage"
	etcPath         string = "/etc/otelcol-sumo"
	systemdPath     string = "/etc/systemd/system/otelcol-sumo.service"
	scriptPath      string = "../../scripts/install.sh"
	installToken    string = "token"
	configPath      string = etcPath + "/sumologic.yaml"

	systemdDirectoryPath string = "/run/systemd/system"
)

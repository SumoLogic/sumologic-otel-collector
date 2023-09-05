package sumologic_scripts_tests

const (
	systemdDirectoryPath  string = "/run/systemd/system"
	systemdPath           string = "/lib/systemd/system/otelcol-sumo.service"
	deprecatedSystemdPath string = "/etc/systemd/system/otelcol-sumo.service"

	rootGroup   string = "root"
	rootUser    string = "root"
	systemGroup string = "otelcol-sumo"
	systemUser  string = "otelcol-sumo"
)

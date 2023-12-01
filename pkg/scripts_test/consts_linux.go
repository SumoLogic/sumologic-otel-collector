package sumologic_scripts_tests

const (
	envDirectoryPath     string = etcPath + "/env"
	systemdDirectoryPath string = "/run/systemd/system"
	systemdPath          string = "/etc/systemd/system/otelcol-sumo.service"
	tokenEnvFilePath     string = envDirectoryPath + "/token.env"

	// TODO: fix mismatch between package permissions & expected permissions
	commonConfigPathFilePermissions uint32 = 0550
	configPathDirPermissions        uint32 = 0550
	configPathFilePermissions       uint32 = 0440
	confDPathFilePermissions        uint32 = 0644
	etcPathPermissions              uint32 = 0551
	opampDPermissions               uint32 = 0750

	rootGroup   string = "root"
	rootUser    string = "root"
	systemGroup string = "otelcol-sumo"
	systemUser  string = "otelcol-sumo"
)

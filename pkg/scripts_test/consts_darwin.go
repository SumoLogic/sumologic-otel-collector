package sumologic_scripts_tests

const (
	appSupportDirPath          string = "/Library/Application Support/otelcol-sumo"
	packageName                string = "otelcol-sumo.pkg"
	launchdPath                string = "/Library/LaunchDaemons/com.sumologic.otelcol-sumo.plist"
	launchdPathFilePermissions uint32 = 0640
	uninstallScriptPath        string = appSupportDirPath + "/uninstall.sh"

	// TODO: fix mismatch between darwin permissions & linux binary install permissions
	configPathDirPermissions  uint32 = 0770
	configPathFilePermissions uint32 = 0440
	confDPathFilePermissions  uint32 = 0644
	etcPathPermissions        uint32 = 0755

	rootGroup   string = "wheel"
	rootUser    string = "root"
	systemGroup string = "_otelcol-sumo"
	systemUser  string = "_otelcol-sumo"
)

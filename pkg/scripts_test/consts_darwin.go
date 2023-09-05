package sumologic_scripts_tests

const (
	appSupportDirPath          string = "/Library/Application Support/otelcol-sumo"
	packageName                string = "otelcol-sumo.pkg"
	launchdPath                string = "/Library/LaunchDaemons/com.sumologic.otelcol-sumo.plist"
	launchdPathFilePermissions uint32 = 0640
	uninstallScriptPath        string = appSupportDirPath + "/uninstall.sh"

	rootGroup   string = "wheel"
	rootUser    string = "root"
	systemGroup string = "_otelcol-sumo"
	systemUser  string = "_otelcol-sumo"
)

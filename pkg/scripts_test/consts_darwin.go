package sumologic_scripts_tests

const (
	appSupportDirPath          string = "/Library/Application Support/otelcol-sumo"
	packageName                string = "otelcol-sumo.pkg"
	launchdPath                string = "/Library/LaunchDaemons/com.sumologic.otelcol-sumo.plist"
	launchdPathFilePermissions uint32 = 0640
	rootGroup                  string = "wheel"
	uninstallScriptPath        string = appSupportDirPath + "/uninstall.sh"

	systemUser  string = "_otelcol-sumo"
	systemGroup string = "otelcol-sumo"
)

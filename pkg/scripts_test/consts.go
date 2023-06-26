package sumologic_scripts_tests

const (
	binaryPath                 string = "/usr/local/bin/otelcol-sumo"
	libPath                    string = "/var/lib/otelcol-sumo"
	fileStoragePath            string = libPath + "/file_storage"
	etcPath                    string = "/etc/otelcol-sumo"
	etcPathPermissions         uint32 = 0551
	systemdPath                string = "/etc/systemd/system/otelcol-sumo.service"
	launchdPath                string = "/Library/LaunchDaemons/com.sumologic.otelcol-sumo.plist"
	launchdPathFilePermissions uint32 = 0640
	scriptPath                 string = "../../scripts/install.sh"
	appSupportDirPath          string = "/Library/Application Support/otelcol-sumo"
	configPath                 string = etcPath + "/sumologic.yaml"
	configPathFilePermissions  uint32 = 0440
	configPathDirPermissions   uint32 = 0550
	confDPath                  string = etcPath + "/conf.d"
	userConfigPath             string = confDPath + "/common.yaml"
	hostmetricsConfigPath      string = confDPath + "/hostmetrics.yaml"
	envDirectoryPath           string = etcPath + "/env"
	tokenEnvFilePath           string = envDirectoryPath + "/token.env"
	cacheDirectory             string = "/var/cache/otelcol-sumo/"
	logDirPath                 string = "/var/log/otelcol-sumo"

	systemdDirectoryPath string = "/run/systemd/system"

	installToken              string = "token"
	installTokenEnv           string = "SUMOLOGIC_INSTALLATION_TOKEN"
	deprecatedInstallTokenEnv string = "SUMOLOGIC_INSTALL_TOKEN"
	apiBaseURL                string = "https://open-collectors.sumologic.com"

	systemUser  string = "otelcol-sumo"
	systemGroup string = "otelcol-sumo"

	darwinPackageName         string = "otelcol-sumo.pkg"
	darwinUninstallScriptPath string = appSupportDirPath + "/uninstall.sh"

	curlTimeoutErrorCode int = 28
)

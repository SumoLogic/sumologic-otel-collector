package sumologic_scripts_tests

const (
	binaryPath                string = "/usr/local/bin/otelcol-sumo"
	libPath                   string = "/var/lib/otelcol-sumo"
	fileStoragePath           string = libPath + "/file_storage"
	etcPath                   string = "/etc/otelcol-sumo"
	etcPathPermissions        uint32 = 0551
	scriptPath                string = "../../scripts/install.sh"
	configPath                string = etcPath + "/sumologic.yaml"
	configPathFilePermissions uint32 = 0440
	configPathDirPermissions  uint32 = 0550
	confDPath                 string = etcPath + "/conf.d"
	userConfigPath            string = confDPath + "/common.yaml"
	hostmetricsConfigPath     string = confDPath + "/hostmetrics.yaml"
	envDirectoryPath          string = etcPath + "/env"
	tokenEnvFilePath          string = envDirectoryPath + "/token.env"
	cacheDirectory            string = "/var/cache/otelcol-sumo/"
	logDirPath                string = "/var/log/otelcol-sumo"

	installToken              string = "token"
	installTokenEnv           string = "SUMOLOGIC_INSTALLATION_TOKEN"
	deprecatedInstallTokenEnv string = "SUMOLOGIC_INSTALL_TOKEN"
	apiBaseURL                string = "https://open-collectors.sumologic.com"

	curlTimeoutErrorCode int = 28
)

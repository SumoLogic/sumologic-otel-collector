package sumologic_scripts_tests

const (
	binaryPath            string = "/usr/local/bin/otelcol-sumo"
	libPath               string = "/var/lib/otelcol-sumo"
	fileStoragePath       string = libPath + "/file_storage"
	etcPath               string = "/etc/otelcol-sumo"
	scriptPath            string = "../../scripts/install.sh"
	configPath            string = etcPath + "/sumologic.yaml"
	confDPath             string = etcPath + "/conf.d"
	userConfigPath        string = confDPath + "/common.yaml"
	hostmetricsConfigPath string = confDPath + "/hostmetrics.yaml"
	cacheDirectory        string = "/var/cache/otelcol-sumo/"
	logDirPath            string = "/var/log/otelcol-sumo"
	envDirectoryPath      string = etcPath + "/env"
	tokenEnvFilePath      string = envDirectoryPath + "/token.env"

	installToken              string = "token"
	installTokenEnv           string = "SUMOLOGIC_INSTALLATION_TOKEN"
	deprecatedInstallTokenEnv string = "SUMOLOGIC_INSTALL_TOKEN"
	apiBaseURL                string = "https://open-collectors.sumologic.com"
	mockAPIBaseURL            string = "http://127.0.0.1:3333"

	// previousBinaryVersion and previousPackageVersion specify a previous
	// version to test upgrades from. It is necessary to upgrade from an older
	// version as same-version package upgrades can behave differently than
	// proper upgrades.
	previousBinaryVersion  string = "0.82.0-sumo-0"
	previousPackageVersion string = "0.83.0-551"

	curlTimeoutErrorCode int = 28

	commonConfigPathFilePermissions uint32 = 0660
	configPathDirPermissions        uint32 = 0770
	configPathFilePermissions       uint32 = 0440
	confDPathFilePermissions        uint32 = 0644
	etcPathPermissions              uint32 = 0751
)

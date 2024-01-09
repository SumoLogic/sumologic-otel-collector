package sumologic_scripts_tests

const (
	binaryPath            string = "/usr/local/bin/otelcol-sumo"
	libPath               string = "/var/lib/otelcol-sumo"
	fileStoragePath       string = libPath + "/file_storage"
	etcPath               string = "/etc/otelcol-sumo"
	scriptPath            string = "../../scripts/install.sh"
	configPath            string = etcPath + "/sumologic.yaml"
	confDPath             string = etcPath + "/conf.d"
	opampDPath            string = etcPath + "/opamp.d"
	userConfigPath        string = confDPath + "/common.yaml"
	hostmetricsConfigPath string = confDPath + "/hostmetrics.yaml"
	cacheDirectory        string = "/var/cache/otelcol-sumo/"
	logDirPath            string = "/var/log/otelcol-sumo"

	installToken    string = "token"
	installTokenEnv string = "SUMOLOGIC_INSTALLATION_TOKEN"
	apiBaseURL      string = "https://open-collectors.sumologic.com"

	curlTimeoutErrorCode int = 28
)

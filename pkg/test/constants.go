package sumologic_tests

const (
	binaryPath                = "../../otelcolbuilder/cmd/otelcol-sumo"
	configTag                 = "--config"
	logFilePath               = "output/test.log"
	validateCommand           = "validate"
	credentialsDir            = "/tmp/lib-temp/otelcol-sumo/credentials"
	sumlogicMockURL           = "http://localhost:3000"
	sumologicMockLogCountPath = "/logs/count?test=ValidateIngestionSumologicExporter"
)

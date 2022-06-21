package mysqlrecordsreceiver

import (
	"bytes"
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
)

func TestMySQLReceiverIntegration(t *testing.T) {

	container := getContainer(t, containerRequest8_0)
	defer func() {
		require.NoError(t, container.Terminate(context.Background()))
	}()
	hostname, err := container.Host(context.Background())
	require.NoError(t, err)

	f := NewFactory()
	cfg := f.CreateDefaultConfig().(*Config)
	cfg.Endpoint = net.JoinHostPort(hostname, "3306")
	cfg.Username = "otel"
	cfg.Password = "otel"
	cfg.Database = "information_schema"
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Show tables"

	consumer := new(consumertest.LogsSink)
	settings := componenttest.NewNopReceiverCreateSettings()
	receiver, err := f.CreateLogsReceiver(context.Background(), settings, cfg, consumer)
	require.NoError(t, err, "failed creating logs receiver")
	require.NoError(t, receiver.Start(context.Background(), componenttest.NewNopHost()))
	require.Eventuallyf(t, func() bool {
		return len(consumer.AllLogs()) > 0
	}, 2*time.Minute, 1*time.Second, "failed to receive more than 0 logs")
	actualLog := consumer.AllLogs()[0]
	logsMarshaler := plog.NewJSONMarshaler()
	buf, err := logsMarshaler.MarshalLogs(actualLog)
	require.NoError(t, err, "failed marshalling log record")
	logRecord := bytes.NewBuffer(buf).String()
	require.NotEmpty(t, logRecord)
	require.NoError(t, receiver.Shutdown(context.Background()))
}

var (
	containerRequest8_0 = testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    filepath.Join("testdata", "integration"),
			Dockerfile: "Dockerfile.mysql.8_0",
		},
		ExposedPorts: []string{"3306:3306"},
		WaitingFor: wait.ForListeningPort("3306").
			WithStartupTimeout(2 * time.Minute),
	}
)

func getContainer(t *testing.T, req testcontainers.ContainerRequest) testcontainers.Container {
	require.NoError(t, req.Validate())
	container, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	require.NoError(t, err)

	code, err := container.Exec(context.Background(), []string{"/setup.sh"})
	require.NoError(t, err)
	require.Equal(t, 0, code)
	return container
}

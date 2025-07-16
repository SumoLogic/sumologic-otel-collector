package telegrafreceiver

import "github.com/influxdata/telegraf"

type channelOutput struct{ outc chan telegraf.Metric }

func (*channelOutput) SampleConfig() string {
	return "null"
}

func (*channelOutput) Connect() error { return nil }

func (c *channelOutput) Close() error {
	close(c.outc)
	return nil
}
func (c *channelOutput) Write(metrics []telegraf.Metric) error {
	for i := range metrics {
		c.outc <- metrics[i]
	}
	return nil
}

func newChannelOutput(outc chan telegraf.Metric) *channelOutput {
	return &channelOutput{
		outc: outc,
	}
}

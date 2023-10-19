package sumologicextension

import "os"

func init() {
	hostname = os.Hostname()
}

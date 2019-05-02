package daemon

import (
	"os"
)

func newDaemon(exepath, appName, serverName string) (daemon, error) {
	descrip := appName + " server daemon"
	depends := []string{"network.target"}

	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return &systemDaemon{exepath, serverName, descrip, depends}, nil
	}
	if _, err := os.Stat("/sbin/initctl"); err == nil {
		return &upstartDaemon{exepath, serverName, descrip, depends}, nil
	}
	return &systemVDaemon{exepath, serverName, descrip, depends}, nil
}

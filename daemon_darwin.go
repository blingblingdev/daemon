package daemon

import (
	"os"
	"os/exec"
	"regexp"
	"text/template"
)

type darwinDaemon struct {
	exePath string
	name    string
	descrip string
}

func newDaemon(exepath, appName, serverName string) (daemon, error) {
	descrip := appName + " server daemon"
	return &darwinDaemon{exepath, serverName, descrip}, nil
}

func (darwin *darwinDaemon) servicePlistPath() string {
	return "/Library/LaunchDaemons/com.nomadli." + darwin.name + ".plist"
}

func (darwin *darwinDaemon) IsInstalled() bool {
	_, err := os.Stat(darwin.servicePlistPath())
	return err == nil
}

func (darwin *darwinDaemon) isRunning() bool {
	stdout, err := exec.Command("launchctl", "list", darwin.name).Output()
	if err != nil {
		return false
	}

	matched, err := regexp.MatchString(darwin.name, string(stdout))
	return err == nil && matched
}

func (darwin *darwinDaemon) Install(args ...string) error {
	if !checkRootGroup() {
		return errPermit
	}

	if darwin.IsInstalled() {
		return nil
	}

	path := darwin.servicePlistPath()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	templ, err := template.New("DarwinTemplate").Parse(DarwinTemplate)
	if err != nil {
		return err
	}

	if err := templ.Execute(
		file,
		&struct {
			Name, Path string
			Args       []string
		}{darwin.name, darwin.exePath, args},
	); err != nil {
		return err
	}

	return nil
}

func (darwin *darwinDaemon) UnInstall() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !darwin.IsInstalled() {
		return nil
	}

	if darwin.isRunning() {
		if err := darwin.Stop(); err != nil {
			return err
		}
	}

	return os.Remove(darwin.servicePlistPath())
}

func (darwin *darwinDaemon) Start() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !darwin.IsInstalled() {
		return errNoInstall
	}

	if darwin.isRunning() {
		return nil
	}

	return exec.Command("launchctl", "load", darwin.servicePlistPath()).Run()
}

func (darwin *darwinDaemon) Stop() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !darwin.IsInstalled() {
		return errNoInstall
	}

	if !darwin.isRunning() {
		return nil
	}

	return exec.Command("launchctl", "unload", darwin.servicePlistPath()).Run()
}

func (darwin *darwinDaemon) Restart() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !darwin.IsInstalled() {
		return errNoInstall
	}

	return exec.Command("launchctl", "reload", darwin.servicePlistPath()).Run()
}

func (darwin *darwinDaemon) Status() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !darwin.IsInstalled() {
		return errNoInstall
	}

	if darwin.isRunning() {
		return nil
	}

	return errStopped
}

func (darwin *darwinDaemon) Run() error {
	return nil
}

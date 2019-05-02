package daemon

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
)

type systemVDaemon struct {
	exePath  string
	name     string
	descrip  string
	dependes []string
}

func (da *systemVDaemon) serviceScrpitPath() string {
	return "/etc/init.d/" + da.name
}

func (da *systemVDaemon) IsInstalled() bool {
	_, err := os.Stat(da.serviceScrpitPath())
	return err == nil
}

func (da *systemVDaemon) isRunning() bool {
	stdout, err := exec.Command("service", da.name, "status").Output()
	if err != nil {
		return false
	}
	matched, err := regexp.MatchString(da.name, string(stdout))
	return err == nil && matched
}

func (da *systemVDaemon) Install(args ...string) error {
	if !checkRootGroup() {
		return errPermit
	}

	if da.IsInstalled() {
		return nil
	}

	path := da.serviceScrpitPath()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	templ, err := template.New("LinuxSystemVTemplate").Parse(LinuxSystemVTemplate)
	if err != nil {
		return err
	}

	if err := templ.Execute(
		file,
		&struct {
			Name, Path, Description, WorkDir, Args string
		}{da.name, da.exePath, da.descrip, strings.TrimRight(da.exePath, da.name), strings.Join(args, " ")},
	); err != nil {
		return err
	}

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	// Create logratate conf
	path = "/etc/logrotate.d/" + da.name
	logfile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer logfile.Close()

	if err = os.MkdirAll("/var/log/"+da.name, 0644); err != nil {
		return err
	}

	templ, err = template.New("LinuxLogRotateTemplate").Parse(LinuxLogRotateTemplate)
	if err != nil {
		return err
	}

	if err = templ.Execute(
		logfile,
		&struct {
			Name string
		}{da.name},
	); err != nil {
		return err
	}
	if err = os.Chmod(path, 0755); err != nil {
		return err
	}

	return exec.Command("chkconfig", "--add", da.name).Run()
}

func (da *systemVDaemon) UnInstall() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return nil
	}

	if da.isRunning() {
		if err := da.Stop(); err != nil {
			return err
		}
	}

	if err := exec.Command("chkconfig", "--del", da.name).Run(); err != nil {
		return err
	}

	return os.Remove(da.serviceScrpitPath())
}

func (da *systemVDaemon) Start() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	if da.isRunning() {
		return nil
	}

	return exec.Command("service", da.name, "start").Run()
}

func (da *systemVDaemon) Stop() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	if !da.isRunning() {
		return nil
	}

	return exec.Command("service", da.name, "stop").Run()
}

func (da *systemVDaemon) Restart() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	return exec.Command("service", da.name, "restart").Run()
}

func (da *systemVDaemon) Status() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	if da.isRunning() {
		return nil
	}

	return errStopped
}

func (da *systemVDaemon) Run() error {
	return nil
}

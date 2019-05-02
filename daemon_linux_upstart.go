package daemon

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
)

type upstartDaemon struct {
	exePath  string
	name     string
	descrip  string
	dependes []string
}

func (da *upstartDaemon) serviceScrpitPath() string {
	return "/etc/init/" + da.name + ".conf"
}

func (da *upstartDaemon) isRunning() bool {
	stdout, err := exec.Command("status", da.name).Output()
	if err != nil {
		return false
	}

	matched, err := regexp.MatchString(da.name+" start/running", string(stdout))
	return err == nil && matched
}

func (da *upstartDaemon) IsInstalled() bool {
	_, err := os.Stat(da.serviceScrpitPath())
	return err == nil
}

func (da *upstartDaemon) Install(args ...string) error {
	if !checkRootGroup() {
		return errPermit
	}

	if da.IsInstalled() {
		return nil
	}

	// Create Upstart conf
	path := da.serviceScrpitPath()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	templ, err := template.New("LinuxUpTemplate").Parse(LinuxUpTemplate)
	if err != nil {
		return err
	}

	if err := templ.Execute(
		file,
		&struct {
			Name, Description, Path, WorkDir, Args string
		}{da.name, da.descrip, da.exePath, strings.TrimRight(da.exePath, da.name), strings.Join(args, " ")},
	); err != nil {
		return err
	}
	if err = os.Chmod(path, 0644); err != nil {
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

	return os.Chmod(path, 0755)
}

func (da *upstartDaemon) UnInstall() error {
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

	return os.Remove(da.serviceScrpitPath())
}

func (da *upstartDaemon) Start() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	if da.isRunning() {
		return nil
	}

	return exec.Command("start", da.name).Run()
}

func (da *upstartDaemon) Stop() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	if !da.isRunning() {
		return nil
	}

	return exec.Command("stop", da.name).Run()
}

func (da *upstartDaemon) Restart() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	return exec.Command("restart", da.name).Run()
}

func (da *upstartDaemon) Status() error {
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

func (da *upstartDaemon) Run() error {
	return nil
}

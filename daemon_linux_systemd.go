package daemon

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
)

type systemDaemon struct {
	exePath  string
	name     string
	descrip  string
	dependes []string
}

func (da *systemDaemon) serviceScrpitPath() string {
	return "/etc/systemd/system/" + da.name + ".service"
}

func (da *systemDaemon) IsInstalled() bool {
	_, err := os.Stat(da.serviceScrpitPath())
	return err == nil
}

func (da *systemDaemon) isRunning() bool {
	stdout, err := exec.Command("systemctl", "status", da.name).Output()
	if err != nil {
		return false
	}

	matched, err := regexp.MatchString("Active: active", string(stdout))
	return err == nil && matched
}

func (da *systemDaemon) Install(args ...string) error {
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

	templ, err := template.New("LinuxSystemDTemplate").Parse(LinuxSystemDTemplate)
	if err != nil {
		return err
	}

	if err := templ.Execute(
		file,
		&struct {
			Description, Dependencies, WorkDir, Name, Path, Args string
		}{da.descrip, strings.Join(da.dependes, " "), strings.TrimRight(da.exePath, da.name), da.name, da.exePath, strings.Join(args, " ")},
	); err != nil {
		return err
	}

	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return err
	}

	return exec.Command("systemctl", "enable", da.name).Run()
}

func (da *systemDaemon) UnInstall() error {
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

	if err := exec.Command("systemctl", "disable", da.name).Run(); err != nil {
		return err
	}

	return os.Remove(da.serviceScrpitPath())
}

func (da *systemDaemon) Start() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	if da.isRunning() {
		return nil
	}

	return exec.Command("systemctl", "start", da.name).Run()
}

func (da *systemDaemon) Stop() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	if !da.isRunning() {
		return nil
	}

	return exec.Command("systemctl", "stop", da.name).Run()
}

func (da *systemDaemon) Restart() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !da.IsInstalled() {
		return errNoInstall
	}

	return exec.Command("systemctl", "restart", da.name).Run()
}

func (da *systemDaemon) Status() error {
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

func (da *systemDaemon) Run() error {
	return nil
}

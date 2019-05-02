// +build dragonfly freebsd netbsd openbsd

package daemon

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
)

type bsdDaemon struct {
	exePath string
	name    string
	descrip string
}

func newDaemon(exepath, appName, serverName string) (daemon, error) {
	descrip := appName + " server daemon"
	return &bsdDaemon{exepath, serverName, descrip}, nil
}

func (bsd *bsdDaemon) serviceScrpitPath() string {
	return "/usr/local/etc/rc.d/" + bsd.name
}

func (bsd *bsdDaemon) IsInstalled() bool {
	_, err := os.Stat(bsd.serviceScrpitPath())
	return err == nil
}

func (bsd *bsdDaemon) isEnabled() bool {
	rcConf, err := os.Open("/etc/rc.conf")
	if err != nil {
		return false
	}
	defer rcConf.Close()
	rcData, _ := ioutil.ReadAll(rcConf)
	r, _ := regexp.Compile(`.*` + bsd.name + `_enable="YES".*`)
	v := string(r.Find(rcData))
	var chrFound, sharpFound bool
	for _, c := range v {
		if c == '#' && !chrFound {
			sharpFound = true
			break
		} else if !sharpFound && c != ' ' {
			chrFound = true
			break
		}
	}
	return chrFound
}

func (bsd *bsdDaemon) getCmd(cmd string) string {
	if !bsd.isEnabled() {
		fmt.Println("Service is not enabled, using \"one" + cmd + "\" instead")
		cmd = "one" + cmd
	}
	return cmd
}

func (bsd *bsdDaemon) isRunning() bool {
	stdout, err := exec.Command("service", bsd.name, bsd.getCmd("status")).Output()
	if err != nil {
		return false
	}
	matched, err := regexp.MatchString(bsd.name, string(stdout))
	return err == nil && matched
}

func (bsd *bsdDaemon) Install(args ...string) error {
	if !checkRootGroup() {
		return errPermit
	}

	if bsd.IsInstalled() {
		return nil
	}

	path := bsd.serviceScrpitPath()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	templ, err := template.New("FreeBSDTemplate").Parse(FreeBSDTemplate)
	if err != nil {
		return err
	}

	if err := templ.Execute(
		file,
		&struct {
			Name, Description, Path, WorkDir, Args string
		}{bsd.name, bsd.descrip, bsd.exePath, strings.TrimRight(bsd.exePath, bsd.name), strings.Join(args, " ")},
	); err != nil {
		return err
	}

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	return nil
}

func (bsd *bsdDaemon) UnInstall() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !bsd.IsInstalled() {
		return nil
	}

	if bsd.isRunning() {
		if err := bsd.Stop(); err != nil {
			return err
		}
	}

	return os.Remove(bsd.serviceScrpitPath())
}

func (bsd *bsdDaemon) Start() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !bsd.IsInstalled() {
		return errNoInstall
	}

	if bsd.isRunning() {
		return nil
	}

	if err := exec.Command("service", bsd.name, bsd.getCmd("start")).Run(); err != nil {
		return err
	}

	return nil
}

func (bsd *bsdDaemon) Stop() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !bsd.IsInstalled() {
		return errNoInstall
	}

	if !bsd.isRunning() {
		return nil
	}

	if err := exec.Command("service", bsd.name, bsd.getCmd("stop")).Run(); err != nil {
		return err
	}

	return nil
}

func (bsd *bsdDaemon) Restart() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !bsd.IsInstalled() {
		return errNoInstall
	}

	if err := exec.Command("service", bsd.name, bsd.getCmd("restart")).Run(); err != nil {
		return err
	}

	return nil
}

func (bsd *bsdDaemon) Status() error {
	if !checkRootGroup() {
		return errPermit
	}

	if !bsd.IsInstalled() {
		return errNoInstall
	}

	if bsd.isRunning() {
		return nil
	}

	return errStopped
}

func (bsd *bsdDaemon) Run() error {
	return nil
}

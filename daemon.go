package daemon

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var (
	errPermit    = errors.New("You must have root privileges")
	errNoInstall = errors.New("Service is not installed")
	errStopped   = errors.New("stopped")
)

type daemon interface {
	IsInstalled() bool
	Install(args ...string) error
	UnInstall() error
	Start() error
	Stop() error
	Status() error
	Restart() error
	Run() error
}

func winServerRun() {
	var exepath string
	var err error
	if exepath, err = filepath.Abs(os.Args[0]); err != nil {
		fmt.Printf("get the %s path error %v\n", os.Args[0], err)
		os.Exit(1)
	}
	appName := filepath.Base(exepath)
	serverName := strings.Join(strings.Fields(appName), "_")

	var d daemon
	if d, err = newDaemon(exepath, appName, serverName); err != nil {
		fmt.Printf("call %s daemon error %v\n", appName, err)
		os.Exit(2)
	}
	d.Run()
}

//RunDaemon add daemon fun
//change DarwinTemplate、LinuxSystemDTemplate、LinuxUpTemplater、LinuxSystemVTemplate
func RunDaemon() {
	cmd := ""
	var l int
	if l = len(os.Args); l > 1 {
		cmd = os.Args[l-1]
	}
	switch cmd {
	case "start":
	case "restart":
	case "stop":
	case "status":
	case "install":
	case "uninstall":
	case "-h":
	case "Daemon":
		os.Args = os.Args[:l-1]
		if runtime.GOOS == "windows" {
			winServerRun()
		}
		fallthrough
	default:
		return
	}

	os.Args = os.Args[:l-1]
	var exepath string
	var err error
	if exepath, err = filepath.Abs(os.Args[0]); err != nil {
		fmt.Printf("get the %s path error %v\n", os.Args[0], err)
		os.Exit(1)
	}
	appName := filepath.Base(exepath)
	serverName := strings.Join(strings.Fields(appName), "_")

	var d daemon
	if d, err = newDaemon(exepath, appName, serverName); err != nil {
		fmt.Printf("call %s daemon error %v\n", appName, err)
		os.Exit(2)
	}

	switch cmd {
	case "start":
		if !d.IsInstalled() {
			args := append(os.Args[1:], "Daemon")
			err = d.Install(args...)
		}
		if err == nil {
			err = d.Start()
		}
	case "restart":
		err = d.Restart()
	case "stop":
		err = d.Stop()
	case "status":
		if !d.IsInstalled() {
			fmt.Printf("%s is not install\n", serverName)
			break
		}
		if err = d.Status(); err != nil {
			fmt.Printf("%s is dead\n", serverName)
		} else {
			fmt.Printf("%s is running\n", serverName)
		}
		err = nil
	case "install":
		args := append(os.Args[1:], "Daemon")
		err = d.Install(args...)
	case "uninstall":
		err = d.UnInstall()
	case "-h":
		os.Args = append(os.Args, "-h")
		fmt.Printf("=========================Daemon help=========================\n")
		fmt.Printf("\nUsage: %s start|restart|stop|status|install|uninstall|-h\n", appName)
		fmt.Printf("%s args start \tto start %s service\n", appName, serverName)
		fmt.Printf("%s restart \t\tto restart %s service\n", appName, serverName)
		fmt.Printf("%s stop \t\tto stop %s service\n", appName, serverName)
		fmt.Printf("%s status \t\tto show %s service status\n", appName, serverName)
		fmt.Printf("sudo %s args install \tto install %s service\n", appName, serverName)
		fmt.Printf("sudo %s uninstall \tto uninstall %s service\n", appName, serverName)
		fmt.Printf("-h \t\t\t show this page\n")
		fmt.Printf("\n\n=========================App help=========================\n")
		return
	}
	if err != nil {
		fmt.Printf("to %s %s err:%v\n", cmd, serverName, err)
	}
	os.Exit(0)
}

func checkRootGroup() bool {
	stdout, err := exec.Command("id", "-g").Output()
	if err != nil {
		return false
	}

	gid, err := strconv.ParseUint(strings.TrimSpace(string(stdout)), 10, 32)
	if err != nil {
		return false
	}

	return gid == 0
}

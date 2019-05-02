package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type windowsDaemon struct {
	exePath  string
	name     string
	descrip  string
	dependes []string
}

type systemError struct {
	Title       string
	Description string
	Action      string
}

var (
	// WinErrCode contains necessary windows error codes.
	WinErrCode = map[int]systemError{
		5: systemError{
			Title:       "ERROR_ACCESS_DENIED",
			Description: "Access denied.",
			Action:      "Administrator access is needed to install a service.",
		},
		1051: systemError{
			Title:       "ERROR_DEPENDENT_SERVICES_RUNNING",
			Description: "A stop control has been sent to a service that other running services are dependent on.",
		},
		1052: systemError{
			Title:       "ERROR_INVALID_SERVICE_CONTROL",
			Description: "The requested control is not valid for this service.",
		},
		1053: systemError{
			Title:       "ERROR_SERVICE_REQUEST_TIMEOUT",
			Description: "The service did not respond to the start or control request in a timely fashion.",
		},
		1054: systemError{
			Title:       "ERROR_SERVICE_NO_THREAD",
			Description: "A thread could not be created for the service.",
		},
		1055: systemError{
			Title:       "ERROR_SERVICE_DATABASE_LOCKED",
			Description: "The service database is locked.",
		},
		1056: systemError{
			Title:       "ERROR_SERVICE_ALREADY_RUNNING",
			Description: "An instance of the service is already running.",
		},
		1057: systemError{
			Title:       "ERROR_INVALID_SERVICE_ACCOUNT",
			Description: "The account name is invalid or does not exist, or the password is invalid for the account name specified.",
		},
		1058: systemError{
			Title:       "ERROR_SERVICE_DISABLED",
			Description: "The service cannot be started, either because it is disabled or because it has no enabled devices associated with it.",
		},
		1060: systemError{
			Title:       "ERROR_SERVICE_DOES_NOT_EXIST",
			Description: "The specified service does not exist as an installed service.",
		},
		1061: systemError{
			Title:       "ERROR_SERVICE_CANNOT_ACCEPT_CTRL",
			Description: "The service cannot accept control messages at this time.",
		},
		1062: systemError{
			Title:       "ERROR_SERVICE_NOT_ACTIVE",
			Description: "The service has not been started.",
		},
		1063: systemError{
			Title:       "ERROR_FAILED_SERVICE_CONTROLLER_CONNECT",
			Description: "The service process could not connect to the service controller.",
		},
		1064: systemError{
			Title:       "ERROR_EXCEPTION_IN_SERVICE",
			Description: "An exception occurred in the service when handling the control request.",
		},
		1066: systemError{
			Title:       "ERROR_SERVICE_SPECIFIC_ERROR",
			Description: "The service has returned a service-specific error code.",
		},
		1068: systemError{
			Title:       "ERROR_SERVICE_DEPENDENCY_FAIL",
			Description: "The dependency service or group failed to start.",
		},
		1069: systemError{
			Title:       "ERROR_SERVICE_LOGON_FAILED",
			Description: "The service did not start due to a logon failure.",
		},
		1070: systemError{
			Title:       "ERROR_SERVICE_START_HANG",
			Description: "After starting, the service hung in a start-pending state.",
		},
		1071: systemError{
			Title:       "ERROR_INVALID_SERVICE_LOCK",
			Description: "The specified service database lock is invalid.",
		},
		1072: systemError{
			Title:       "ERROR_SERVICE_MARKED_FOR_DELETE",
			Description: "The specified service has been marked for deletion.",
		},
		1073: systemError{
			Title:       "ERROR_SERVICE_EXISTS",
			Description: "The specified service already exists.",
		},
		1075: systemError{
			Title:       "ERROR_SERVICE_DEPENDENCY_DELETED",
			Description: "The dependency service does not exist or has been marked for deletion.",
		},
		1077: systemError{
			Title:       "ERROR_SERVICE_NEVER_STARTED",
			Description: "No attempts to start the service have been made since the last boot.",
		},
		1078: systemError{
			Title:       "ERROR_DUPLICATE_SERVICE_NAME",
			Description: "The name is already in use as either a service name or a service display name.",
		},
		1079: systemError{
			Title:       "ERROR_DIFFERENT_SERVICE_ACCOUNT",
			Description: "The account specified for this service is different from the account specified for other services running in the same process.",
		},
		1083: systemError{
			Title:       "ERROR_SERVICE_NOT_IN_EXE",
			Description: "The executable program that this service is configured to run in does not implement the service.",
		},
		1084: systemError{
			Title:       "ERROR_NOT_SAFEBOOT_SERVICE",
			Description: "This service cannot be started in Safe Mode.",
		},
	}
)

func newDaemon(exepath, appName, serverName string) (daemon, error) {
	return &windowsDaemon{exepath, serverName, appName + " server daemon", []string{""}}, nil
}

func toWinError(err error) error {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			if sysErr, ok := WinErrCode[status.ExitStatus()]; ok {
				return fmt.Errorf("\n %s: %s \n %s", sysErr.Title, sysErr.Description, sysErr.Action)
			}
		}
	}

	return err
}

func (win *windowsDaemon) IsInstalled() bool {
	m, err := mgr.Connect()
	if err != nil {
		return false
	}
	defer m.Disconnect()

	s, err := m.OpenService(win.name)
	if err != nil {
		return false
	}

	s.Close()
	return true
}

func (win *windowsDaemon) Install(args ...string) error {
	// var n uint32
	// b := make([]uint16, syscall.MAX_PATH)
	// size := uint32(len(b))

	// r0, _, err := syscall.MustLoadDLL(
	// 	"kernel32.dll",
	// ).MustFindProc(
	// 	"GetModuleFileNameW",
	// ).Call(0, uintptr(unsafe.Pointer(&b[0])), uintptr(size))
	// n = uint32(r0)
	// if n == 0 {
	// 	return err
	// }
	// execp := string(utf16.Decode(b[0:n]))

	m, err := mgr.Connect()
	if err != nil {
		return toWinError(err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(win.name)
	if err == nil {
		s.Close()
		return nil
	}

	s, err = m.CreateService(win.name, win.exePath, mgr.Config{
		DisplayName:  win.name,
		Description:  win.descrip,
		StartType:    mgr.StartAutomatic,
		Dependencies: win.dependes,
	}, args...)
	if err != nil {
		return toWinError(err)
	}
	defer s.Close()

	r := []mgr.RecoveryAction{
		mgr.RecoveryAction{
			Type:  mgr.ServiceRestart,
			Delay: 5 * time.Second,
		},
		mgr.RecoveryAction{
			Type:  mgr.ServiceRestart,
			Delay: 5 * time.Second,
		},
		mgr.RecoveryAction{
			Type:  mgr.ServiceRestart,
			Delay: 5 * time.Second,
		},
	}
	err = s.SetRecoveryActions(r, 0)
	if err != nil {
		return toWinError(err)
	}

	return nil
}

func (win *windowsDaemon) UnInstall() error {
	win.Stop()
	m, err := mgr.Connect()
	if err != nil {
		return toWinError(err)
	}
	defer m.Disconnect()
	s, err := m.OpenService(win.name)
	if err != nil {
		return nil
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return toWinError(err)
	}

	return nil
}

func (win *windowsDaemon) Start() error {
	m, err := mgr.Connect()
	if err != nil {
		return toWinError(err)
	}
	defer m.Disconnect()
	s, err := m.OpenService(win.name)
	if err != nil {
		return toWinError(err)
	}
	defer s.Close()
	if err = s.Start(); err != nil {
		return toWinError(err)
	}

	return nil
}

func (win *windowsDaemon) Stop() error {
	m, err := mgr.Connect()
	if err != nil {
		return toWinError(err)
	}
	defer m.Disconnect()
	s, err := m.OpenService(win.name)
	if err != nil {
		return toWinError(err)
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return toWinError(err)
	}

	timeout := time.Millisecond * 20000
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control`, registry.READ)
	for err == nil {
		sv, _, err := key.GetStringValue("WaitToKillServiceTimeout")
		if err != nil {
			break
		}
		v, err := strconv.Atoi(sv)
		if err != nil {
			break
		}
		timeout = time.Millisecond * time.Duration(v)
		break
	}

	timeDuration := time.Millisecond * 50
	waite := time.After(timeout + timeDuration*2)
	tick := time.NewTicker(timeDuration)
	defer tick.Stop()

	for status.State != svc.Stopped {
		select {
		case <-tick.C:
			status, err = s.Query()
			if err != nil {
				return toWinError(err)
			}
		case <-waite:
			return fmt.Errorf("Stop %s timeout", win.name)
		}
	}
	return nil
}

func (win *windowsDaemon) status() svc.State {
	m, err := mgr.Connect()
	if err != nil {
		return svc.State(8)
	}
	defer m.Disconnect()
	s, err := m.OpenService(win.name)
	if err != nil {
		return svc.State(8)
	}
	defer s.Close()
	status, err := s.Query()
	if err != nil {
		return svc.State(8)
	}

	return status.State
}

func (win *windowsDaemon) Status() error {
	switch win.status() {
	case svc.StartPending:
		fallthrough
	case svc.Running:
		fallthrough
	case svc.ContinuePending:
		return nil
	}
	return errStopped
}

func (win *windowsDaemon) Restart() error {
	if win.Status() == nil {
		if err := win.Stop(); err != nil {
			return err
		}
	}
	return win.Start()
}

func (win *windowsDaemon) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}

	//call true server run function

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.Stopped}
				go win.exit(0)
			case svc.Pause:
				changes <- svc.Status{State: svc.Stopped}
				go win.exit(0)
			case svc.Continue:
				changes <- svc.Status{State: svc.Stopped}
				go win.exit(0)
			default:
			}
		}
	}
}

func (win *windowsDaemon) Run() error {
	//interactive, err := svc.IsAnInteractiveSession()
	err := svc.Run(win.name, win)
	if err != nil {
		return toWinError(err)
	}

	return nil
}

func (win *windowsDaemon) exit(code int) {
	time.Sleep(time.Millisecond * 30)
	os.Exit(code)
}

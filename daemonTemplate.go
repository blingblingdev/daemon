package daemon

var (
	//DarwinTemplate for mac osx service template
	DarwinTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>KeepAlive</key>
	<true/>
	<key>Label</key>
	<string>{{.Name}}</string>
	<key>ProgramArguments</key>
	<array>
	    <string>{{.Path}}</string>
		{{range .Args}}<string>{{.}}</string>{{end}}
	</array>
	<key>RunAtLoad</key>
	<true/>
    <key>WorkingDirectory</key>
    <string>/usr/local/var</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/{{.Name}}.err</string>
    <key>StandardOutPath</key>
    <string>/usr/local/var/log/{{.Name}}.log</string>
</dict>
</plist>
`

	// FreeBSDTemplate for freebsd service template
	FreeBSDTemplate = `#!/bin/sh

. /etc/rc.subr
    
name="{{.Name}}"
rcvar="{{.Name}}_enable"
command="{{.Path}}"
pidfile="/var/run/$name.pid"

start_cmd="cd {{.WorkDir}} && /usr/sbin/daemon -p $pidfile -f $command {{.Args}}"
load_rc_config $name
run_rc_command "$1"
`

	//LinuxSystemDTemplate for Linux super systemctl service template
	LinuxSystemDTemplate = `[Unit]
Description={{.Description}}
Requires={{.Dependencies}}
After={{.Dependencies}}

[Service]
WorkingDirectory={{.WorkDir}}
PIDFile=/var/run/{{.Name}}.pid
User=root
Group=root
ExecStartPre=/bin/rm -f /var/run/{{.Name}}.pid
ExecStart={{.Path}} {{.Args}}
ExecStopPost=/bin/rm -f /var/run/{{.Name}}.pid
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`

	//LinuxUpTemplate for Linux super initctl service template
	LinuxUpTemplate = `# {{.Name}} {{.Description}}

description     "{{.Description}}"
author          "nomadli <dzym79@qq.com>"

start on runlevel [2345]
stop on runlevel [016]

#expect fork

respawn
respawn limit 10 5

chdir {{.WorkDir}}

script
    exec {{.Path}} {{.Args}} 2>&1 >> /var/log/{{.Name}}/{{.Name}}.log
end script
`

	//LinuxSystemVTemplate for Linux super systemV service template
	LinuxSystemVTemplate = `#! /bin/sh
# chkconfig: 2345 98 17
# description: Starts and stops a single {{.Name}} instance on this system


if [ -f /etc/rc.d/init.d/functions ]; then
    . /etc/rc.d/init.d/functions
fi

exec="{{.Path}}"
servname="{{.Description}}"

proc="{{.Name}}"
pidfile="/var/run/$proc.pid"
lockfile="/var/lock/subsys/$proc"
logfile="/var/log/{{.Name}}/$proc.log"

[ -d $(dirname $lockfile) ] || mkdir -p $(dirname $lockfile)

[ -d $(dirname $logfile) ] || mkdir -p $(dirname $logfile)

[ -e /etc/sysconfig/$proc ] && . /etc/sysconfig/$proc

start() {
    [ -x $exec ] || exit 5

    if [ -f $pidfile ]; then
        if ! [ -d "/proc/$(cat $pidfile)" ]; then
            rm $pidfile
            if [ -f $lockfile ]; then
                rm $lockfile
            fi
        fi
    fi

    if ! [ -f $pidfile ]; then
        printf "Starting $servname:\t"
        echo "$(date)" >> $logfile
        su root -c cd {{.WorkDir}}
        su root -c $exec {{.Args}} 2>&1 >> $logfile &
        echo $! > $pidfile
        touch $lockfile
        success
        echo
    else
        echo
        printf "$pidfile still exists...\n"
        exit 7
    fi
}

stop() {
    echo -n $"Stopping $servname: "
    killproc -p $pidfile $proc
    retval=$?
    echo
    [ $retval -eq 0 ] && rm -f $lockfile
    return $retval
}

restart() {
    stop
    start
}

rh_status() {
    status -p $pidfile $proc
}

rh_status_q() {
    rh_status >/dev/null 2>&1
}

case "$1" in
    start)
        rh_status_q && exit 0
        $1
        ;;
    stop)
        rh_status_q || exit 0
        $1
        ;;
    restart)
        $1
        ;;
    status)
        rh_status
        ;;
    *)
        echo $"Usage: $0 {start|stop|status|restart}"
        exit 2
esac

exit $?
`
	// LinuxLogRotateTemplate for logrotate template
	LinuxLogRotateTemplate = `
/var/log/{{.Name}}/*.log {
    weekly
    maxsize 10M
    rotate 10
    copytruncate
    delaycompress
    compress
    notifempty
    missingok
    su root root
}
`
)

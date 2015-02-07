package main

import (
	"github.com/hironobu-s/conoha-vps/command"
	"github.com/hironobu-s/conoha-vps/lib"
	"os"
)

func main() {
	var err error

	log := lib.GetLogInstance()

	var cmd command.Commander
	var subcommand string = ""

	if len(os.Args) > 1 {
		subcommand = os.Args[1]
	}

	log.Debugf("starting %s subcommand...", subcommand)

	switch subcommand {
	case "login":
		cmd = command.NewLogin()
	case "stat":
		cmd = command.NewVpsStat()
	case "list":
		cmd = command.NewVpsList()
	case "add":
		cmd = command.NewVpsAdd()
	case "remove":
		cmd = command.NewVpsRemove()
	case "label":
		cmd = command.NewVpsLabel()
	case "ssh-key":
		cmd = command.NewSshKey()
	case "ssh":
		cmd = command.NewSsh()
	case "logout":
		cmd = command.NewLogout()
	case "version":
		cmd = command.NewVersion()
	default:
		cmd = command.NewNocommand()
	}
	defer cmd.Shutdown()

	// login, logout, version, nocommand以外はログインが必須
	_, nocommand := cmd.(*command.Nocommand)
	if subcommand != "login" && subcommand != "version" && subcommand != "logout" && !nocommand {
		l := command.NewLogin()

		loggedIn, _ := l.LoggedIn()
		if !loggedIn {
			log.Debugf("Session is timed out. try relogin...")

			// 再ログイン
			loggedIn, err = l.Relogin()
			if !loggedIn {
				log.Errorf("Session is timed out. Please log in.")
				return
			}
		}
	}

	if err = cmd.Run(); err != nil {
		// ShowUsageErrorの場合はUsage()を表示してるだけなのでログは表示しない
		_, ok := err.(*command.ShowUsageError)
		if !ok {
			log.Error(err)
		}
	}
}

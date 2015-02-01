package command

import (
	"errors"
	"fmt"
	flag "github.com/ogier/pflag"
	"os"
)

type Logout struct {
	*Command
}

func NewLogout() *Logout {
	return &Logout{
		Command: NewCommand(),
	}
}

func (cmd *Logout) parseFlag() error {
	var help bool

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	fs.BoolVarP(&help, "help", "h", false, "help")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		return err
	}

	if help {
		fs.Usage()
		return errors.New("")
	}

	return nil
}

func (cmd *Logout) Usage() {
	fmt.Println(`Usage: conoha logout [OPTIONS ...]

DESCRIPTION
    Remove an authenticate file(~/.conoha-vps).

OPTIONS
    -h: --help:     Show usage.      
`)
}

func (cmd *Logout) Run() error {
	if err := cmd.parseFlag(); err != nil {
		return err
	}

	cmd.config.Remove()
	return nil
}

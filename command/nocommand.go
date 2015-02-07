package command

import (
	"fmt"
)

type Nocommand struct {
	*Command
}

func NewNocommand() *Nocommand {
	return &Nocommand{
		Command: NewCommand(),
	}
}

func (cmd *Nocommand) parseFlag() error {
	return nil
}

func (cmd *Nocommand) Usage() {
	fmt.Println(`Usage: conoha COMMAND [OPTIONS]

DESCRIPTION
    A CLI-Tool for ConoHa VPS.

COMMANDS
    add      Add VPS.
    label    Change VPS label.
    list     List VPS.
    login    Authenticate an account.
    logout   Remove an authenticate file(~/.conoha-vps).
    remove   Remove VPS.
    ssh-key  Download and store SSH Private key.
    ssh      Login to VPS via SSH.
    stat     Display VPS information.
    version  Display version.
`)
}

func (cmd *Nocommand) Run() error {
	cmd.Usage()
	return &ShowUsageError{}
}

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
    login    Authenticate an account.
    list     List VPS.
    add      Add VPS.
    remove   Remove VPS.
    ssh-key  Download and store SSH Private key.
    ssh      Login to VPS via SSH.
    logout   Remove an authenticate file(~/.conoha-vps).
    version  Print version.
`)
}

func (cmd *Nocommand) Run() error {
	cmd.Usage()
	return nil
}

package command

import (
	"errors"
	"fmt"
	flag "github.com/ogier/pflag"
	"os"
	"os/exec"
)

type Ssh struct {
	// SSH接続先のVM
	vmId string

	// SSHコマンドに渡すオプション
	sshOptions []string

	*Vps
}

func NewSsh() *Ssh {
	return &Ssh{
		Vps: NewVps(),
	}
}

func (cmd *Ssh) parseFlag() error {
	var help bool

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	fs.BoolVarP(&help, "help", "h", false, "help")

	fs.Parse(os.Args[1:])

	if help {
		fs.Usage()
		return errors.New("")
	}

	if len(fs.Args()) < 2 {
		vm, err := cmd.Vps.vpsSelectMenu()
		if err != nil {
			return err
		}

		cmd.vmId = vm.Id

	} else {
		// 接続先のVmのID
		cmd.vmId = os.Args[2]
		cmd.sshOptions = os.Args[3:]
	}

	return nil
}

func (cmd *Ssh) Usage() {
	fmt.Println(`Usage: conoha ssh <VPS-ID> [OPTIONS ...]

DESCRIPTION
    Login to VPS via SSH.
    There needs to be installed SSH client and all of option parameters will be passed into SSH command.

    It may not work on Windows.

<VPS-ID> VPS-ID to get the stats. It may be confirmed by LIST subcommand.
         If not set, It will be selected from the list menu.

OPTIONS
    -h: --help:     Show usage.      
`)
}

func (cmd *Ssh) Run() error {
	var err error
	if err = cmd.parseFlag(); err != nil {
		return err
	}

	vpsList := NewVpsList()
	vm := vpsList.Vm(cmd.vmId)
	if vm == nil {
		msg := fmt.Sprintf("VPS not found(id=%s).", cmd.vmId)
		return errors.New(msg)
	}

	vpsStat := NewVpsStat()
	stat, err := vpsStat.Stat(vm.Id)
	if err != nil {
		return err
	}

	cmd.Connect(stat.IPv4, "root", cmd.sshOptions)

	return nil
}

func (cmd *Ssh) Connect(host string, user string, args []string) {

	options := []string{
		user + "@" + host,
	}

	options = append(options, args...)

	sshCmd := "ssh"

	ssh := exec.Command(sshCmd, options...)
	ssh.Stdin = os.Stdin
	ssh.Stdout = os.Stdout
	ssh.Stderr = os.Stderr

	ssh.Run()
}

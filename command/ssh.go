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

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Parse(os.Args[1:])

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

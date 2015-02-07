package command

import (
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-vps/lib"
	flag "github.com/ogier/pflag"
	"os"
	"os/exec"
	"strings"
)

type Ssh struct {
	// SSH接続先のVM
	vmId string

	// SSHユーザ
	sshUser string

	// SSHコマンドに渡すオプション
	sshOptions []string

	*Vps
}

func NewSsh() *Ssh {
	return &Ssh{
		sshUser: "root",
		Vps:     NewVps(),
	}
}

func (cmd *Ssh) parseFlag() error {
	var help bool

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	// pflagsはparse()すると設定していないフラグが全てエラーになってしまう。
	// 仕方ないので、ssh コマンドにオプションを渡せるようにするため、自前でパースする。
	options := []string{}
	for i := 2; i < len(os.Args); i++ {
		// 最初の引数は - で開始してない場合はVPS-IDとみなす
		if i == 2 && !strings.HasPrefix(os.Args[i], "-") {
			cmd.vmId = os.Args[i]
			continue
		}

		if os.Args[i] == "-h" {
			help = true
		} else if os.Args[i] == "-u" {
			cmd.sshUser = os.Args[i+1]
			i++
		} else {
			options = append(options, os.Args[i])
		}
	}

	if help {
		fs.Usage()
		return &ShowUsageError{}
	}

	if cmd.vmId == "" {
		vm, err := cmd.Vps.vpsSelectMenu()
		if err != nil {
			return err
		}

		// 接続先のVmのID
		cmd.vmId = vm.Id
	}

	// SSHオプション
	cmd.sshOptions = options

	return nil
}

func (cmd *Ssh) Usage() {
	fmt.Println(`Usage: conoha ssh <VPS-ID> [OPTIONS ...]

DESCRIPTION
    Login to VPS via SSH.
    There needs to be installed SSH client and all of option parameters will be passed into SSH command.

    It may not work on Windows.

<VPS-ID> (Optional)VPS-ID to get the stats. It may be confirmed by LIST subcommand.
         If not set, It will be selected from the list menu.

OPTIONS
    -u: --user:     SSH username.
    -h: --help:     Show usage.      
`)
}

func (cmd *Ssh) Run() error {

	log := lib.GetLogInstance()

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

	// Windwsプランの場合は何もしない
	if strings.Index(vm.Plan, "Windows") >= 0 {
		log.Infof("ID=%s. Windows plan is not supported ssh connect.", vm.Id)
		return nil
	}

	vpsStat := NewVpsStat()
	stat, err := vpsStat.Stat(vm.Id)
	if err != nil {
		return err
	}

	cmd.Connect(stat.IPv4, cmd.sshUser, cmd.sshOptions)

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

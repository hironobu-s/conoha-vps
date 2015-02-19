package command

import (
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"github.com/hironobu-s/conoha-vps/lib"
	flag "github.com/ogier/pflag"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

//COMMAND
const (
	BOOT     = "Boot"
	REBOOT   = "Reboot"
	SHUTDOWN = "Shutdown"
	STOP     = "Stop"
)

type VpsPower struct {
	vmId      string
	command   string
	forceSend bool

	*Vps
}

func NewVpsPower() *VpsPower {
	return &VpsPower{
		Vps: NewVps(),
	}
}

func (cmd *VpsPower) parseFlag() error {
	var help bool
	var command string

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	fs.BoolVarP(&help, "help", "h", false, "help")
	fs.StringVarP(&command, "command", "c", "", "power command")
	fs.BoolVarP(&cmd.forceSend, "force-send", "f", false, "force send")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	if help {
		fs.Usage()
		return &ShowUsageError{}
	}

	if command == "" {
		fs.Usage()
		return errors.New("Not enough arguments.")
	}

	switch command {
	case "boot":
		cmd.command = BOOT
	case "reboot":
		cmd.command = REBOOT
	case "shutdown":
		cmd.command = SHUTDOWN
	case "stop":
		cmd.command = STOP
	default:
		return errors.New(fmt.Sprintf(`Undefined command "%s".`, command))
	}

	if len(fs.Args()) < 2 {
		// コマンドライン引数で指定されていない場合は、標準入力から受け付ける
		vm, err := cmd.Vps.vpsSelectMenu()
		if err != nil {
			return err
		}
		cmd.vmId = vm.Id
	}

	return nil
}

func (cmd *VpsPower) Usage() {
	fmt.Println(`Usage: conoha power <VPS-ID> [OPTIONS]

DESCRIPTION
    Send power-command to VPS.

<VPS-ID> (Optional)VPS-ID to get the stats. It may be confirmed by LIST subcommand.
         If not set, It will be selected with prompting for VPS list.

OPTIONS
    -c: --command:     Power command. It should be one of following.
                       ("boot" "reboot" "shutdown" "stop")

    -f: --force-send:  Attempt to send without prompting for confirmation.

    -h: --help:        Show usage.
`)
}

func (cmd *VpsPower) Run() error {
	if err := cmd.parseFlag(); err != nil {
		return err
	}

	return cmd.SendCommand(cmd.vmId, cmd.command)
}

// 電源の状態を変更するコマンドを送信する
func (cmd *VpsPower) SendCommand(vmId string, command string) error {

	// 対象のVMを特定する
	vpsList := NewVpsList()
	vm := vpsList.Vm(vmId)
	if vm == nil {
		return errors.New(fmt.Sprintf("VPS not found(id=%s).", vmId))
	}

	// VPSのステータスを取得する
	stat, _ := cmd.GetVMStatus(vmId)

	// BOOTコマンドは停止中のVPSにのみ送信できる
	if command == BOOT && stat != StatusOffline {
		return errors.New(fmt.Sprintf(`Could not send "%s" command. VPS is already running.`, command))

		// それ以外のコマンドは稼働中のVPSにのみ送信できる
	} else if command != BOOT && stat != StatusRunning {
		return errors.New(fmt.Sprintf(`Could not send "%s" command.  VPS might be offiline.`, command))
	}

	// 確認ダイアログ
	if !cmd.forceSend {
		if !cmd.confirmation(vm, command) {
			return nil
		}
	}

	// コマンドを送信する
	var act *cpanel.Action
	var err error

	act = &cpanel.Action{
		Request: &vpsPowerRequest{
			vmId:    vmId,
			command: command,
		},
		Result: &vpsPowerResult{},
	}

	cmd.browser.AddAction(act)

	if err = cmd.browser.Run(); err != nil {
		return err
	}

	log := lib.GetLogInstance()
	log.Infof(`"%s" command was sent to VPS(id=%s).`, command, vmId)

	return nil
}

// 確認ダイアログ
func (cmd *VpsPower) confirmation(vm *Vm, command string) bool {

	fmt.Printf(`Send "%s" command to VPS(Label=%s). Are you sure?`, command, vm.Label)
	fmt.Println("")
	fmt.Print("[y/N]: ")

	var no string
	if _, err := fmt.Scanf("%s", &no); err != nil {
		return false
	}

	if no == "y" {
		return true
	} else {
		return false
	}
}

type vpsPowerRequest struct {
	vmId    string
	command string
}

func (r *vpsPowerRequest) NewRequest(values url.Values) (*http.Request, error) {
	values = url.Values{}
	values.Add("command", r.command)
	values.Add("evid", r.vmId)
	values.Add("_", strconv.FormatInt(time.Now().Unix(), 10)) // unix epoch

	req, err := http.NewRequest("GET", "https://cp.conoha.jp/Service/VPS/Control/CommandSender.aspx?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

type vpsPowerResult struct {
}

func (r *vpsPowerResult) Populate(resp *http.Response) error {

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Server returned the errror status code(%d).", resp.StatusCode))
	}

	return nil
}

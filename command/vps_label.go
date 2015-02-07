package command

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"github.com/hironobu-s/conoha-vps/lib"
	flag "github.com/ogier/pflag"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// VPSのラベルを変更する

type VpsLabel struct {
	vmId  string
	label string
	*Vps
}

func NewVpsLabel() *VpsLabel {
	return &VpsLabel{
		Vps: NewVps(),
	}
}

func (cmd *VpsLabel) parseFlag() error {
	var help bool

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	fs.BoolVarP(&help, "help", "h", false, "help")
	fs.StringVarP(&cmd.label, "label", "l", "", "Label")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		return err
	}

	if help {
		fs.Usage()
		return &ShowUsageError{}
	}

	if cmd.label == "" {
		return errors.New("Not enough arguments.")
	}

	if len(cmd.label) > 20 {
		return errors.New("Label is too long(should be 20 characters or less). ")
	}

	// VPS-ID
	if len(fs.Args()) < 2 {
		// コマンドライン引数で指定されていない場合は、標準入力から受け付ける
		vm, err := cmd.Vps.vpsSelectMenu()
		if err != nil {
			return err
		}
		cmd.vmId = vm.Id

	} else {
		cmd.vmId = os.Args[2]
	}
	return nil
}

func (cmd *VpsLabel) Usage() {
	fmt.Println(`Usage: conoha label <VPS-ID> [OPTIONS ...]

DESCRIPTION
    Change VPS label.

<VPS-ID> (Optional) VPS-ID to get the stats. It may be confirmed by LIST subcommand.
         If not set, It will be selected from the list menu.

OPTIONS
    -l: --label:    name of label.
    -h: --help:     Show usage.      
`)
}

func (cmd *VpsLabel) Run() error {
	var err error
	if err = cmd.parseFlag(); err != nil {
		return err
	}

	return cmd.Change(cmd.vmId, cmd.label)
}

func (cmd *VpsLabel) Change(vmId string, label string) error {
	var act *cpanel.Action
	act = &cpanel.Action{
		Request: &labelChangeRequest{
			vmId:  vmId,
			label: label,
		},
		Result: &labelChangeResult{},
	}
	cmd.browser.AddAction(act)

	if err := cmd.browser.Run(); err != nil {
		return err
	}

	log := lib.GetLogInstance()
	log.Infof(`VPS Label was changed to "%s"`, cmd.label)

	return nil
}

type labelChangeRequest struct {
	vmId  string
	label string
}

func (r *labelChangeRequest) NewRequest(values url.Values) (*http.Request, error) {
	values = url.Values{}
	values.Add("eid", r.vmId)
	values.Add("label", r.label)
	values.Add("type", "vm") // 固定値

	req, err := http.NewRequest("POST", "https://cp.conoha.jp/Service/ChangeLabel.aspx", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

type labelChangeResult struct{}

func (r *labelChangeResult) Populate(resp *http.Response, doc *goquery.Document) error {

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Server returned the errror status code(%d).", resp.StatusCode))
	}

	return nil
}

package command

// VPSを削除する
// https://cp.conoha.jp/Service/VPS/Del/* のスクレイパー

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

type VpsRemove struct {
	vmId        string
	forceRemove bool
	*Vps
}

func NewVpsRemove() *VpsRemove {
	return &VpsRemove{
		Vps: NewVps(),
	}
}

func (cmd *VpsRemove) parseFlag() error {
	var help bool

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	fs.BoolVarP(&help, "help", "h", false, "help")
	fs.BoolVarP(&cmd.forceRemove, "force-remove", "f", false, "force remove.")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		return err
	}

	if help {
		fs.Usage()
		return errors.New("")
	}

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

func (cd *VpsRemove) Usage() {
	fmt.Println(`Usage: conoha stat <VPS-ID> [OPTIONS]

DESCRIPTION
    Remove VPS.

<VPS-ID> VPS-ID to get the stats. It may be confirmed by LIST subcommand.
         If not set, It will be selected with prompting for VPS list.

OPTIONS
    -h: --help:          Show usage.
    -f: --force-remove:  Attempt to remove the VPS without prompting for comfirmation .
`)
}

func (cmd *VpsRemove) Run() error {
	var err error
	if err = cmd.parseFlag(); err != nil {
		return err
	}

	err = cmd.Remove(cmd.vmId)
	if err != nil {
		return err
	}
	return nil
}

func (cmd *VpsRemove) Remove(vmId string) error {

	log := lib.GetLogInstance()

	// 削除対象のVMを特定する
	vpsList := NewVpsList()
	vm := vpsList.Vm(vmId)
	if vm == nil {
		msg := fmt.Sprintf("VPS not found(id=%s).", vmId)
		return errors.New(msg)
	}

	// 削除確認
	if !cmd.forceRemove {
		if !cmd.confirmationRemove(vm) {
			return nil
		}
	}

	// 削除実行
	var act *cpanel.Action
	act = &cpanel.Action{
		Request: &removeFormRequest{
			vm: vm,
		},
		Result: &removeFormResult{},
	}
	cmd.browser.AddAction(act)

	act = &cpanel.Action{
		Request: &removeConfirmRequest{},
		Result:  &removeConfirmResult{},
	}
	cmd.browser.AddAction(act)

	act = &cpanel.Action{
		Request: &removeSubmitRequest{},
		Result:  &removeSubmitResult{},
	}
	cmd.browser.AddAction(act)

	if err := cmd.browser.Run(); err != nil {
		return err
	}

	log.Infof("Removing VPS is complete.")

	return nil
}

// 削除確認ダイアログ
func (cmd *VpsRemove) confirmationRemove(vm *Vm) bool {

	fmt.Printf("Remove VPS[Label=%s]. Are you sure?\n", vm.Label)
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

type removeFormRequest struct {
	vm *Vm
}

func (r *removeFormRequest) NewRequest(values url.Values) (*http.Request, error) {
	// VPSのIDを指定
	values.Set("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$gridServiceList$"+r.vm.TrId+"$ctl01", "on")

	// これが削除ページのトリガになっているらしい
	values.Set("__EVENTTARGET", "ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$btnDel")

	// フォームを取得
	req, err := http.NewRequest("POST", "https://cp.conoha.jp/Service/VPS/", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://cp.conoha.jp/Service/VPS/")

	return req, nil
}

type removeFormResult struct{}

func (r *removeFormResult) Populate(resp *http.Response, doc *goquery.Document) error {

	// b, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("body: " + string(b))

	// 確認ボタンが表示されていることを確認
	sel := doc.Find("#ContentPlaceHolder1_ContentPlaceHolder1_btnConfirm")
	v, _ := sel.Attr("value")
	if v == "" {
		return errors.New("Server returned the invalid body(Confirm button is not included).")
	}
	return nil
}

// ---------------------------

type removeConfirmRequest struct{}

func (r *removeConfirmRequest) NewRequest(values url.Values) (*http.Request, error) {
	values.Set("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$btnConfirm", "確認")

	// フォームを取得
	req, err := http.NewRequest("POST", "https://cp.conoha.jp/Service/VPS/Del/Default.aspx", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

type removeConfirmResult struct{}

func (r *removeConfirmResult) Populate(resp *http.Response, doc *goquery.Document) error {
	// 決定ボタンが表示されていることを確認
	sel := doc.Find("#ContentPlaceHolder1_ContentPlaceHolder1_btnConfirm")
	v, _ := sel.Attr("value")
	if v == "" {
		return errors.New("Server returned the invalid body(Submit button is not included).")
	}
	return nil
}

// ---------------------------

type removeSubmitRequest struct{}

func (r *removeSubmitRequest) NewRequest(values url.Values) (*http.Request, error) {
	values.Set("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$btnConfirm", "決定")

	// フォームを取得
	req, err := http.NewRequest("POST", "https://cp.conoha.jp/Service/VPS/Del/Confirm.aspx", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://cp.conoha.jp/Service/VPS/Del/Default.aspx")
	return req, nil
}

type removeSubmitResult struct{}

func (r *removeSubmitResult) Populate(resp *http.Response, doc *goquery.Document) error {
	// 削除に成功するとBodyに通知メッセージが含まれている
	sel := doc.Find("#ltInfoMessage")
	if sel.Text() != "" {
		return nil
	} else {
		msg := fmt.Sprintf("Server returned the invalid body(Info Message is not include).")
		return errors.New(msg)
	}
}

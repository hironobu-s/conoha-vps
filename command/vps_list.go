package command

// VPSの一覧を取得する
// https://cp.conoha.jp/Service/VPS/ のスクレイパー

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	flag "github.com/ogier/pflag"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type VpsList struct {
	*Vps
	idOnly  bool
	verbose bool
}

func NewVpsList() *VpsList {
	return &VpsList{
		Vps: NewVps(),
	}
}

func (cmd *VpsList) parseFlag() error {
	var help bool

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	fs.BoolVarP(&help, "help", "h", false, "help")
	fs.BoolVarP(&cmd.idOnly, "id-only", "i", false, "id-only")
	fs.BoolVarP(&cmd.verbose, "Verbose", "v", false, "Verbose output.")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		return err
	}

	if help {
		fs.Usage()
		return &ShowUsageError{}
	}

	return nil
}

func (cd *VpsList) Usage() {
	fmt.Println(`Usage: conoha list [OPTIONS]

DESCRIPTION
    List VPS status.

OPTIONS
    -h: --help:     Show usage.
    -i: --id-only:  Show VPS-ID only.
    -v: --verbose:  Verbose output.
                    Including the server status, but slowly.
`)
}

func (cmd *VpsList) Run() error {
	var err error

	if err = cmd.parseFlag(); err != nil {
		return err
	}

	var servers []*Vm
	servers, err = cmd.List(cmd.verbose)
	if err != nil {
		return err
	}

	if cmd.idOnly {
		format := "%-20s\n"
		for _, vm := range servers {
			fmt.Printf(format, vm.Id)
		}

	} else {

		var maxPlan int = 10
		var maxLabel int = 10
		for _, vm := range servers {
			if len(vm.Label) > maxLabel {
				maxLabel = len(vm.Label)
			}
			if len(vm.Plan) > maxPlan {
				maxPlan = len(vm.Plan)
			}
		}

		format := "%-20s\t%-" + strconv.Itoa(maxLabel) + "s\t%-" + strconv.Itoa(maxPlan) + "s\t%-15s\t%-20s\t%s\n"

		fmt.Printf(
			format,
			"VPS ID           ",
			"Label",
			"Plan",
			"Server Status",
			"Service Status",
			"CreatedAt",
		)
		for _, vm := range servers {
			fmt.Printf(
				format,
				vm.Id,
				vm.Label,
				vm.Plan,
				vm.ServerStatus,
				vm.ServiceStatus,
				vm.CreatedAt.Format("2006/01/02 15:04 MST"),
			)
		}
	}
	return nil
}

// Vmを取得する
// 引数のIDのVmが見つかった場合はその構造体を、見つからない場合はnilを返す。
func (cmd *VpsList) Vm(vmId string) *Vm {
	var target *Vm

	servers, err := cmd.List(false)
	if err != nil {
		return nil
	}

	for _, vps := range servers {
		if vps.Id == vmId {
			target = vps
		}
	}

	return target
}

// VPSの一覧を取得して、IDをキー、Vm構造体のポインタを値としたスライスを返す
// 引数のdeepCrawlをtrueにすると、VMのステータスも取得する
func (cmd *VpsList) List(deep bool) (servers []*Vm, err error) {

	var act *cpanel.Action

	r := &listResult{}

	act = &cpanel.Action{
		Request: &listRequest{},
		Result:  r,
	}
	cmd.browser.AddAction(act)

	if err := cmd.browser.Run(); err != nil {
		return nil, err
	}

	// サーバーステータスを取得する
	if deep {
		for _, vm := range r.servers {
			status, err := cmd.GetVMStatus(vm.Id)
			if err != nil {
				return r.servers, err
			}
			vm.ServerStatus = status
		}
	} else {
		for _, vm := range servers {
			vm.ServerStatus = StatusNoinformation
		}
	}

	return r.servers, nil
}

// VPS一覧を取得するリクエスト
type listRequest struct {
}

func (r *listRequest) NewRequest(values url.Values) (*http.Request, error) {
	return http.NewRequest("GET", "https://cp.conoha.jp/Service/VPS/", nil)
}

type listResult struct {
	servers []*Vm
}

func (r *listResult) Populate(resp *http.Response, doc *goquery.Document) error {

	// VPSの一覧を取得する
	sel := doc.Find("#gridServiceList TR")

	servers := []*Vm{}
	for i := range sel.Nodes {
		tr := sel.Eq(i)
		tds := tr.Find("TD")

		if len(tds.Nodes) == 0 {
			continue
		}

		// Vm構造体を準備
		vm := &Vm{}

		// TrIDを取得
		trid, exists := tr.Attr("id")
		if !exists {
			return errors.New("TrID not exists")
		}
		vm.TrId = trid

		// VMの各要素を取得
		c := 0
		for j := range tds.Nodes {

			value := strings.Trim(tds.Eq(j).Text(), " \t\r\n")
			switch c {
			case 1:
				// GetVMStatus()で設定するのでここでは初期値を設定
				vm.ServerStatus = StatusNoinformation
			case 2:
				vm.Label = value

				// VPSのIDを取得
				href, exists := tds.Eq(j).Find("A").Attr("href")
				if exists {
					sp := strings.Split(href, "/")
					vm.Id = sp[2]
				} else {
					// VPSの作成待ちの場合はIDが存在しない場合がある
					vm.Id = ""
				}

			case 3:
				vm.ServiceStatus = value
			case 4:
				vm.ServiceId = value
			case 5:
				vm.Plan = value
			case 6:
				vm.CreatedAt, _ = time.Parse("Jan/02/2006 15:04 MST", value+" JST")
			case 7:
				vm.DeleteDate, _ = time.Parse("Jan/02/2006 15:04 MST", value+" JST")
			case 8:
				vm.PaymentSpan = value
			}

			c++
		}

		servers = append(servers, vm)
	}

	r.servers = servers

	return nil
}

// --------------------------------

// コントロールパネルのAjaxリクエストと同等
// サーバーのステータス定数を返す
type GetVMStatusJson struct {
	StatusId    string `json:"status_id"`
	StatusName  string `json:"status_name"`
	StatusClass string `json:"status_class"`
}

type vmStatusResult struct {
	VmId   string
	Status ServerStatus
}

func (cmd *Vps) GetVMStatus(id string) (status ServerStatus, err error) {

	if id == "" {
		return StatusUnknown, nil
	}

	r := &vmStatusResult{}
	f := &vmStatusRequest{}

	act := &cpanel.Action{
		Request: f,
		Result:  r,
	}

	values := url.Values{}
	values.Add("evid", id)
	cmd.browser.BrowserInfo.Values = values

	cmd.browser.AddAction(act)

	if err = cmd.browser.Run(); err != nil {
		return StatusUnknown, err
	} else {
		return r.Status, nil
	}
}

type vmStatusRequest struct {
}

func (r *vmStatusRequest) NewRequest(values url.Values) (*http.Request, error) {
	u, err := url.Parse("https://cp.conoha.jp/Service/VPS/GetVMStatus.aspx?" + values.Encode())
	if err != nil {
		return nil, err
	}

	return http.NewRequest("GET", u.String(), nil)
}

func (r *vmStatusResult) Populate(resp *http.Response) error {

	j := &GetVMStatusJson{}
	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(j)
	if err != nil {
		r.Status = StatusUnknown
		return err
	}

	switch j.StatusName {
	case "Running":
		r.Status = StatusRunning
	case "Offline":
		r.Status = StatusOffline
	case "In-use":
		r.Status = StatusInUse
	case "In-formulation":
		r.Status = StatusInFormulation
	default:
		r.Status = StatusUnknown
	}
	return nil
}

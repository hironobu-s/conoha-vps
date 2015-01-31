package command

// VPSの一覧を取得する
// https://cp.conoha.jp/Service/VPS/ のスクレイパー

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Vmを取得する
// 引数のIDのVmが見つかった場合はその構造体を、見つからない場合はnilを返す。
func (cmd *Vps) Vm(vmId string) *Vm {
	var target *Vm

	servers, err := cmd.List(false)
	if err != nil {
		return nil
	}

	for id, vps := range servers {
		if id == vmId {
			target = vps
		}
	}

	return target
}

// VPSの一覧を取得して、IDをキー、Vm構造体のポインタを値としたスライスを返す
// 引数のdeepCrawlをtrueにすると、VMのステータスも取得する
func (cmd *Vps) List(deep bool) (servers map[string]*Vm, err error) {

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
	servers map[string]*Vm
}

func (r *listResult) Populate(resp *http.Response, doc *goquery.Document) error {

	// VPSの一覧を取得する
	sel := doc.Find("#gridServiceList TR")

	servers := map[string]*Vm{}
	for i := range sel.Nodes {
		tr := sel.Eq(i)
		tds := tr.Find("TD")

		if len(tds.Nodes) == 0 {
			continue
		}

		// Vm構造体を準備
		vm := &Vm{}

		// VMの各要素を取得
		c := 0
		for j := range tds.Nodes {

			value := strings.Trim(tds.Eq(j).Text(), " \t\r\n")
			switch c {
			case 1:
				// GetVMStatus()で設定するのでここでは無視する
			case 2:
				vm.Label = value

				// VPSのIDを取得
				href, exists := tds.Eq(j).Find("A").Attr("href")
				if !exists {
					return errors.New(`Nothing Attribute "href". Could not detect the VmID.`)
				}
				sp := strings.Split(href, "/")
				vm.Id = sp[2]

			case 3:
				vm.ServiceStatus = value
			case 4:
				vm.ServiceId = value
			case 5:
				vm.Plan = value
			case 6:
				vm.CreatedAt, _ = time.Parse("2006/01/02 15:04", value)
			case 7:
				vm.DeleteDate, _ = time.Parse("2006/01/02 15:04", value)
			case 8:
				vm.PaymentSpan = value
			}

			c++
		}

		servers[vm.Id] = vm
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
	case "起動中":
		r.Status = StatusRunning
	case "停止":
		r.Status = StatusOffline
	case "設定中":
		r.Status = StatusInUse
	case "構築中":
		r.Status = StatusInFormulation
	default:
		r.Status = StatusUnknown
	}
	return nil
}

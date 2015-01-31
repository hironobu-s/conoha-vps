package command

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Vmの詳細を取得する
func (cmd *Vps) Stat(vmId string) (*Vm, error) {
	vm := cmd.Vm(vmId)
	if vm == nil {
		msg := fmt.Sprintf("VPS not found(id=%s).", vmId)
		return nil, errors.New(msg)
	}

	act := &cpanel.Action{
		Request: &statRequest{
			vm: vm,
		},
		Result: &statResult{
			vm: vm,
		},
	}

	cmd.browser.AddAction(act)
	if err := cmd.browser.Run(); err != nil {
		return nil, err
	}

	return vm, nil
}

type statRequest struct {
	vm *Vm
}

func (r *statRequest) NewRequest(values url.Values) (*http.Request, error) {
	rawurl := "https://cp.conoha.jp/Service/VPS/Control/Console/" + r.vm.Id
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	return http.NewRequest("GET", u.String(), nil)
}

type statResult struct {
	vm *Vm
}

func (r *statResult) Populate(resp *http.Response, doc *goquery.Document) error {
	subbox := doc.Find("#subCtrlBox .subCtrlList TD")

	// VPS詳細
	for i := range subbox.Nodes {
		td := subbox.Eq(i)

		// ラベルを削除
		spans := td.Find("SPAN")
		for j := range spans.Nodes {
			spans.Eq(j).Remove()
		}

		switch i {
		case 0:
			r.vm.NumCpuCore = td.Text()
		case 1:
			r.vm.Memory = td.Text()
		case 2:
			r.vm.Disk1Size = td.Text()
		case 3:
			r.vm.Disk2Size = td.Text()
		case 5:
			r.vm.IPv4 = td.Text()
		case 6:
			r.vm.IPv4netmask = td.Text()
		case 7:
			r.vm.IPv4gateway = td.Text()
		case 8:
			r.vm.IPv4dns1 = td.Text()
		case 9:
			r.vm.IPv4dns2 = td.Text()
		case 10:
			tmp := strings.Split(td.Text(), "\n")
			for i := 0; i < len(tmp); i++ {
				if ipv6 := strings.Trim(tmp[i], " \r\n\t"); ipv6 != "" {
					r.vm.IPv6 = append(r.vm.IPv6, ipv6)
				}
			}
		case 11:
			r.vm.IPv6prefix = td.Text()
		case 12:
			r.vm.IPv6gateway = td.Text()
		case 13:
			r.vm.IPv6dns1 = td.Text()
		case 14:
			r.vm.IPv6dns2 = td.Text()
		case 15:
			r.vm.House = td.Text()
		case 16:
			r.vm.CommonServerId = td.Text()
		}
	}

	// ------------------------
	var err error
	if err = r.populateDate(doc); err != nil {
		return err
	}

	if err = r.populateUploadHost(doc); err != nil {
		return err
	}
	return nil
}

func (r *statResult) populateDate(doc *goquery.Document) error {
	var body string
	var reg *regexp.Regexp
	var matches [][]string
	var err error
	var date time.Time

	// 利用開始日
	body = doc.Find("#subCtrlBoxNav .startData").Text()

	reg = regexp.MustCompile("利用開始日:([0-9/]*)")
	matches = reg.FindAllStringSubmatch(body, -1)

	date, err = time.Parse("2006/01/02 MST", matches[0][1]+" JST")
	if err != nil {
		return err
	}
	r.vm.CreatedAt = date

	// 削除予定日
	body = doc.Find("#subCtrlBoxNav .endData").Text()

	reg = regexp.MustCompile("削除予定日:([0-9/]*)")
	matches = reg.FindAllStringSubmatch(body, -1)

	date, err = time.Parse("2006/01/02 MST", matches[0][1]+" JST")
	if err == nil {
		r.vm.DeleteDate = date
	}
	return nil
}

func (r *statResult) populateUploadHost(doc *goquery.Document) error {
	// ISOアップロード先とシリアルコンソール接続先
	body := doc.Find("DL.listStyle01").Text()
	reg := regexp.MustCompile("接続先：(.+)／")

	matches := reg.FindAllStringSubmatch(body, -1)
	if len(matches) != 2 || len(matches[0]) != 2 || len(matches[1]) != 2 {
		// パースエラー
		return errors.New("Parse error. Can't detect ISO upload host or serial console host.")
	}

	r.vm.SerialConsoleHost = matches[0][1]
	r.vm.IsoUploadHost = matches[1][1]
	return nil
}

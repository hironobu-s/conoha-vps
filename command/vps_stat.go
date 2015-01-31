package command

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	flag "github.com/ogier/pflag"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VpsStat struct {
	*Vps
	vmId    string
	incIPv6 bool
}

func NewVpsStat() *VpsStat {
	return &VpsStat{
		Vps: NewVps(),
	}
}

func (cmd *VpsStat) parseFlag() error {

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)

	fs.StringVarP(&cmd.vmId, "id", "i", "", "VPS-ID or Label")
	fs.BoolVarP(&cmd.incIPv6, "include-ipv6", "v", false, "Including IPv6 informations.")

	fs.Parse(os.Args[1:])

	if cmd.vmId == "" {
		// コマンドライン引数で指定されていない場合は、標準入力から受け付ける
		if err := cmd.scanf(); err != nil {
			return err
		}
	}
	return nil
}

// 標準入力からVpsIdを読み込む
func (cmd *VpsStat) scanf() error {
	var n int
	var err error

	println("Please enter VPS-ID or Label in order to get the status.")
	print("ID or Label: ")
	n, err = fmt.Scanf("%s", &cmd.vmId)
	if n != 1 || err != nil {
		return err
	}

	return nil
}

func (cmd *VpsStat) Run() error {
	var err error
	if err = cmd.parseFlag(); err != nil {
		return err
	}

	vm, err := cmd.Stat(cmd.vmId)
	if err != nil {
		return err
	}

	var lines []string = []string{}

	padding := 20
	format := "%-" + strconv.Itoa(padding) + "s %s"

	lines = append(lines, fmt.Sprintf(format, "VPS ID", vm.Id))
	lines = append(lines, fmt.Sprintf(format, "ServerStatus", vm.ServerStatus))
	lines = append(lines, fmt.Sprintf(format, "Label", vm.Label))
	lines = append(lines, fmt.Sprintf(format, "ServiceStatus", vm.ServiceStatus))
	lines = append(lines, fmt.Sprintf(format, "Service ID", vm.ServiceId))
	lines = append(lines, fmt.Sprintf(format, "Plan", vm.Plan))
	lines = append(lines, fmt.Sprintf(format, "Created At", vm.CreatedAt.Format(time.RFC3339)))
	lines = append(lines, fmt.Sprintf(format, "Delete Date", vm.DeleteDate))
	lines = append(lines, fmt.Sprintf(format, "Payment Span", vm.PaymentSpan))
	lines = append(lines, fmt.Sprintf(format, "CPU", vm.NumCpuCore))
	lines = append(lines, fmt.Sprintf(format, "Memory", vm.Memory))
	lines = append(lines, fmt.Sprintf(format, "Disk1", vm.Disk1Size))
	lines = append(lines, fmt.Sprintf(format, "Disk2", vm.Disk2Size))

	lines = append(lines, fmt.Sprintf(format, "IPv4 Address", vm.IPv4))
	lines = append(lines, fmt.Sprintf(format, "IPv4 Netmask", vm.IPv4netmask))
	lines = append(lines, fmt.Sprintf(format, "IPv4 Gateway", vm.IPv4gateway))
	lines = append(lines, fmt.Sprintf(format, "IPv4 DNS1", vm.IPv4dns1))
	lines = append(lines, fmt.Sprintf(format, "IPv4 DNS2", vm.IPv4dns2))

	if cmd.incIPv6 {
		for i := 0; i < len(vm.IPv6); i++ {
			if i == 0 {
				lines = append(lines, fmt.Sprintf(format, "IPv6 Address", vm.IPv6[i]))
			} else {
				lines = append(lines, fmt.Sprintf(format, "", vm.IPv6[i]))
			}
		}
		lines = append(lines, fmt.Sprintf(format, "IPv6 Gateway", vm.IPv6gateway))
		lines = append(lines, fmt.Sprintf(format, "IPv6 DNS1", vm.IPv6dns1))
		lines = append(lines, fmt.Sprintf(format, "IPv6 DNS2", vm.IPv6dns2))
	}

	lines = append(lines, fmt.Sprintf(format, "Host Server", vm.House))
	lines = append(lines, fmt.Sprintf(format, "Common Server ID", vm.CommonServerId))
	lines = append(lines, fmt.Sprintf(format, "Serial Console(SSH)", vm.SerialConsoleHost))
	lines = append(lines, fmt.Sprintf(format, "ISO Upload(SFTP)", vm.IsoUploadHost))

	fmt.Println(strings.Join(lines, "\n"))

	return nil
}

// Vmの詳細を取得する
func (cmd *VpsStat) Stat(vmId string) (*Vm, error) {
	vpsList := NewVpsList()
	vm := vpsList.Vm(vmId)
	if vm == nil {
		var msg string
		if vmId == "" {
			msg = fmt.Sprintf("VPS not found.")
		} else {
			msg = fmt.Sprintf("VPS not found(id=%s).", vmId)
		}
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

	reg = regexp.MustCompile("Started:([0-9/]*)")
	matches = reg.FindAllStringSubmatch(body, -1)

	if len(matches) > 0 && len(matches[0]) > 1 && matches[0][1] != "" {
		date, err = time.Parse("2006/01/02 MST", matches[0][1]+" JST")
		if err != nil {
			return err
		}
		r.vm.CreatedAt = date
	} else if matches[0][1] == "" {
		// 日付が空欄。何もしない
	} else {
		// パースエラー
		return errors.New("Parse error. Can't detect CreatedAt.")
	}

	// 削除予定日
	body = doc.Find("#subCtrlBoxNav .endData").Text()

	reg = regexp.MustCompile("Scheduled Removal Date:([0-9/]*)")
	matches = reg.FindAllStringSubmatch(body, -1)

	if len(matches) > 0 && len(matches[0]) > 1 && matches[0][1] != "" {
		date, err = time.Parse("2006/01/02 MST", matches[0][1]+" JST")
		if err == nil {
			r.vm.DeleteDate = date
		}
	} else if matches[0][1] == "" {
		// 日付が空欄。何もしない
	} else {
		// パースエラー
		return errors.New("Parse error. Can't detect DeleteDate.")
	}
	return nil
}

func (r *statResult) populateUploadHost(doc *goquery.Document) error {
	// ISOアップロード先とシリアルコンソール接続先
	body := doc.Find("DL.listStyle01").Text()
	reg := regexp.MustCompile("Connect to: (.+)/")

	matches := reg.FindAllStringSubmatch(body, -1)
	if len(matches) != 2 || len(matches[0]) != 2 || len(matches[1]) != 2 {
		// パースエラー
		return errors.New("Parse error. Can't detect ISO upload host or serial console host.")
	}

	r.vm.SerialConsoleHost = matches[0][1]
	r.vm.IsoUploadHost = matches[1][1]
	return nil
}

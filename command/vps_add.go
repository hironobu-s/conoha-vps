package command

// VPSを追加する
// https://cp.conoha.jp/Service/VPS/Add/ のスクレイパー

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	PlanTypeBasic = 1 + iota
	PlanTypeWindows
)

const (
	Plan1G = 1 + iota
	Plan2G
	Plan4G
	Plan8G
	Plan16G
)

const (
	TemplateDefault1 = 1 + iota
	TemplateDefault2
	TemplateDefault3
	TemplateDefault4
)

// 追加するVPSの情報
// VpsAdd.Add()に渡す場合は PlanType, Plan, Template, RootPasswordをセットすれば良い
type VpsAddInformation struct {

	// プラン種別(PlanType*定数)
	PlanType int

	// プラン(Plan*定数)
	Plan int

	// テンプレートイメージ
	Template int

	// rootパスワード
	RootPassword string

	// ----------

	// VpsPlan構造体
	VpsPlan *VpsPlan

	// SSHキーID
	SshKeyId string
}

type VpsPlan struct {
	label  string
	planId string
}

func (cmd *Vps) Add(info *VpsAddInformation) error {

	if info.PlanType != PlanTypeBasic && info.PlanType != PlanTypeWindows {
		return errors.New("Undefined plan type.")
	}

	if 1 <= info.Plan && info.Plan <= 5 {
		// OK
	} else {
		return errors.New("Undefined plan.")
	}

	var act *cpanel.Action
	act = &cpanel.Action{
		Request: &addFormRequest{},
		Result: &addFormResult{
			info: info,
		},
	}
	cmd.browser.AddAction(act)

	act = &cpanel.Action{
		Request: &addConfirmRequest{
			info: info,
		},
		Result: &addConfirmResult{},
	}
	cmd.browser.AddAction(act)

	act = &cpanel.Action{
		Request: &addSubmitRequest{},
		Result:  &addSubmitResult{},
	}
	cmd.browser.AddAction(act)

	if err := cmd.browser.Run(); err != nil {
		return err
	}
	return nil

	// VPS追加を実行
	// var err error
	// err = v.addFormPrepare(info)
	// if err != nil {
	// 	return err
	// }

	// err = v.addFormConfirm(info)
	// if err != nil {
	// 	return err
	// }

	// err = v.addFormSubmit(info)
	// if err != nil {
	// 	return err
	// }
	return nil
}

// ---------------------- form --------------------

// フォームのHTMLを取得してパラメータを処理する
type addFormRequest struct {
}

func (r *addFormRequest) NewRequest(values url.Values) (*http.Request, error) {
	// フォームを取得
	return http.NewRequest("GET", "https://cp.conoha.jp/Service/VPS/Add/", nil)
}

type addFormResult struct {
	info *VpsAddInformation
}

func (r *addFormResult) Populate(resp *http.Response, doc *goquery.Document) error {
	// プラン決定する
	plans, err := r.detectPlans(doc, r.info.PlanType)
	if err != nil {
		return err
	}
	r.info.VpsPlan = plans[r.info.Plan-1]

	// SSHキーIDを決定する
	r.info.SshKeyId = r.sshKeyId(doc)

	return nil
}

// VPS追加フォームのHTMLからプラン一覧を作る
// 返り値の1GB, 2GB, 4GB, 8GB, 16GBの5要素であることが保証されます。
func (r *addFormResult) detectPlans(doc *goquery.Document, planType int) (plans []*VpsPlan, err error) {
	// Linux Plan
	plans = []*VpsPlan{}

	var sel *goquery.Selection
	if planType == PlanTypeBasic {
		sel = doc.Find("#trLinuxPlan LI")
	} else if planType == PlanTypeWindows {
		sel = doc.Find("#trWindowsPlan LI")
	} else {
		return nil, errors.New("Undefined plan type.")
	}

	i := 1
	for n := range sel.Nodes {
		node := sel.Eq(n)

		var planId, label string
		planId, _ = node.Find("INPUT").Attr("value")
		label = node.Text()

		// プラン名のメモリ容量をチェックする
		if strings.Index(label, strconv.Itoa(i)+"GB") < 0 {
			msg := fmt.Sprintf("Wrong plan name. [%s]", label)
			return nil, errors.New(msg)
		}

		p := &VpsPlan{
			label:  label,
			planId: planId,
		}
		plans = append(plans, p)

		i *= 2
	}

	if len(plans) != 5 {
		return nil, errors.New("The number of Linux plans is not 5.")
	}

	return plans, nil
}

// VPS追加フォームのHTMLからSSH公開鍵のIDを取得する
func (r *addFormResult) sshKeyId(doc *goquery.Document) string {
	sshKeyId, _ := doc.Find("#ContentPlaceHolder1_ContentPlaceHolder1_rbKey_0").Attr("value")
	return sshKeyId
}

// ---------------------- confirm --------------------

// Confirmページのフォームを埋めてPOSTする
type addConfirmRequest struct {
	info *VpsAddInformation
}

func (r *addConfirmRequest) NewRequest(values url.Values) (*http.Request, error) {

	info := r.info

	// プラン種別
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$rbPlanCategory", strconv.Itoa(info.PlanType))

	// プラン
	if info.PlanType == PlanTypeBasic {
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$rbLinuxPlan", info.VpsPlan.planId)
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$rbWindowsPlan", "2364")
	} else {
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$rbLinuxPlan", "")
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$rbWindowsPlan", info.VpsPlan.planId)
	}

	// 支払い(固定値)
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$UnitMonth", "selectCredit1")

	// テンプレートイメージ
	if info.PlanType == PlanTypeBasic {
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$selLinuxOS", "default/"+strconv.Itoa(info.Template))
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$selWindowsOS", "default/"+strconv.Itoa(TemplateDefault3))
	} else {
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$selLinuxOS", "default/"+strconv.Itoa(TemplateDefault1))
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$selWindowsOS", "default/"+strconv.Itoa(info.Template))
	}

	// rootパスワード(標準プランのみ)
	if info.PlanType == PlanTypeBasic {
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$txtRootPassword", info.RootPassword)
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$txtConfirmPassword", info.RootPassword)

	} else {
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$txtRootPassword", "")
		values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$txtConfirmPassword", "")
	}

	// SSHキー
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$rbKey", info.SshKeyId)

	// ほか(適当な固定値でかまわない)
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$btnConfirm", "確認")
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$hfCpu", "2")
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$hfDisk", "100")
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$hfMemory", "1")
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$hfInital", "0円")
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$hfRunning", "507円")

	req, err := http.NewRequest("POST", "https://cp.conoha.jp/Service/VPS/Add/", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req, nil

}

type addConfirmResult struct {
}

func (r addConfirmResult) Populate(resp *http.Response, doc *goquery.Document) error {
	var sel *goquery.Selection

	// rootパスワード不備などのフォームエラー
	sel = doc.Find(".errorMsg")
	for i := range sel.Nodes {
		node := sel.Eq(i)
		return errors.New(strings.Trim(node.Text(), "\r\n \t"))
	}

	// 追加ボタンが存在しない場合はエラー
	sel = doc.Find("#ContentPlaceHolder1_ContentPlaceHolder1_btnExecute")
	v, _ := sel.Attr("value")
	if v == "" {
		return errors.New("Server returned invalid html(Submit button is not included).")
	}

	return nil
}

// ---------------------- submit --------------------

type addSubmitRequest struct {
}

func (r *addSubmitRequest) NewRequest(values url.Values) (*http.Request, error) {
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$btnExecute", "決定")

	req, err := http.NewRequest("POST", "https://cp.conoha.jp/Service/VPS/Add/Confirm.aspx", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

type addSubmitResult struct {
}

func (r *addSubmitResult) Populate(resp *http.Response, doc *goquery.Document) error {
	// 追加に成功するとBodyに通知メッセージが含まれている
	sel := doc.Find("#ltInfoMessage")
	if sel.Text() != "" {
		return nil
	} else {
		msg := fmt.Sprintf("Server returned the invalid body(Info Message is not include).", resp.StatusCode)
		return errors.New(msg)
	}
}

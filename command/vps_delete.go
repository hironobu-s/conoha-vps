package command

// VPSを削除する
// https://cp.conoha.jp/Service/VPS/Del/* のスクレイパー

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"net/http"
	"net/url"
	"strings"
)

func (cmd *Vps) Delete(vmId string) error {
	// 削除対象のVMを特定する
	vm := cmd.Vm(vmId)
	if vm == nil {
		msg := fmt.Sprintf("VPS not found(id=%s).", vmId)
		return errors.New(msg)
	}

	// 削除実行
	var act *cpanel.Action
	act = &cpanel.Action{
		Request: &deleteFormRequest{
			vm: vm,
		},
		Result: &deleteFormResult{},
	}
	cmd.browser.AddAction(act)

	act = &cpanel.Action{
		Request: &deleteConfirmRequest{},
		Result:  &deleteConfirmResult{},
	}
	cmd.browser.AddAction(act)

	act = &cpanel.Action{
		Request: &deleteSubmitRequest{},
		Result:  &deleteSubmitResult{},
	}
	cmd.browser.AddAction(act)

	if err := cmd.browser.Run(); err != nil {
		return err
	}

	return nil
}

type deleteFormRequest struct {
	vm *Vm
}

func (r *deleteFormRequest) NewRequest(values url.Values) (*http.Request, error) {
	// VPSのIDを指定
	values.Set("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$gridServiceList$tr-"+r.vm.Id+"$ctl01", "on")

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

type deleteFormResult struct{}

func (r *deleteFormResult) Populate(resp *http.Response, doc *goquery.Document) error {
	// 確認ボタンが表示されていることを確認
	sel := doc.Find("#ContentPlaceHolder1_ContentPlaceHolder1_btnConfirm")
	v, _ := sel.Attr("value")
	if v == "" {
		return errors.New("Server returned the invalid body(Confirm button is not included).")
	}
	return nil
}

// ---------------------------

type deleteConfirmRequest struct{}

func (r *deleteConfirmRequest) NewRequest(values url.Values) (*http.Request, error) {
	values.Set("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$btnConfirm", "確認")

	// フォームを取得
	req, err := http.NewRequest("POST", "https://cp.conoha.jp/Service/VPS/Del/Default.aspx", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

type deleteConfirmResult struct{}

func (r *deleteConfirmResult) Populate(resp *http.Response, doc *goquery.Document) error {
	// 決定ボタンが表示されていることを確認
	sel := doc.Find("#ContentPlaceHolder1_ContentPlaceHolder1_btnConfirm")
	v, _ := sel.Attr("value")
	if v == "" {
		return errors.New("Server returned the invalid body(Confirm button is not included).")
	}
	return nil
}

// ---------------------------

type deleteSubmitRequest struct{}

func (r *deleteSubmitRequest) NewRequest(values url.Values) (*http.Request, error) {
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

type deleteSubmitResult struct{}

func (r *deleteSubmitResult) Populate(resp *http.Response, doc *goquery.Document) error {
	// 削除に成功するとBodyに通知メッセージが含まれている
	sel := doc.Find("#ltInfoMessage")
	if sel.Text() != "" {
		return nil
	} else {
		msg := fmt.Sprintf("Server returned the invalid body(Info Message is not include).")
		return errors.New(msg)
	}
}

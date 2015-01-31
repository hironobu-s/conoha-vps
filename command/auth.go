package command

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"github.com/howeyc/gopass"
	"github.com/k0kubun/pp"
	flag "github.com/ogier/pflag"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Auth struct {
	account  string
	password string

	*Command
}

func (cmd *Auth) parseFlag() error {

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)

	fs.StringVarP(&cmd.account, "account", "a", "", "ConoHa Account")
	fs.StringVarP(&cmd.password, "password", "p", "", "ConoHa Password")

	fs.Parse(os.Args[1:])

	if cmd.account == "" || cmd.password == "" {

		// コマンドライン引数で指定されていない場合は、標準入力から受け付ける
		if err := cmd.inputAccountInfo(); err != nil {
			return errors.New("Not enough arguments.")
		}
	}

	return nil
}

// 認証を実行してログイン状態を返す
func (cmd *Auth) Auth() (loggedIn bool, err error) {

	if err = cmd.parseFlag(); err != nil {
		return false, err
	}

	cmd.config.Account = cmd.account
	cmd.config.Password = cmd.password

	var act *cpanel.Action

	act = &cpanel.Action{
		Request: &loginFormRequest{},
		Result:  &loginFormResult{},
	}
	cmd.browser.AddAction(act)

	act = &cpanel.Action{
		Request: &loginDoRequest{
			account:  cmd.account,
			password: cmd.password,
		},
		Result: &loginDoResult{},
	}
	cmd.browser.AddAction(act)

	if err := cmd.browser.Run(); err != nil {
		return false, err
	}

	return cmd.LoggedIn()
}

// 標準入力からアカウントとパスワードを読み込む
func (cmd *Auth) inputAccountInfo() error {
	var n int
	var err error

	println("Please input ConoHa accounts.")
	print("ConoHa Account: ")
	n, err = fmt.Scanf("%s", &cmd.account)
	if n != 1 || err != nil {
		return err
	}

	print("Password: ")
	cmd.password = string(gopass.GetPasswd())

	return nil
}

type loginFormRequest struct {
}

func (r *loginFormRequest) NewRequest(values url.Values) (*http.Request, error) {
	return http.NewRequest("GET", "https://cp.conoha.jp/Login.aspx", nil)
}

type loginFormResult struct {
}

func (r *loginFormResult) Populate(resp *http.Response, doc *goquery.Document) error {
	pp.Printf("loginForm done.\n")
	return nil
}

// ---------------------

type loginDoRequest struct {
	account  string
	password string
}

func (r *loginDoRequest) NewRequest(values url.Values) (req *http.Request, err error) {

	values.Set("ctl00$ContentPlaceHolder1$txtConoHaLoginID", r.account)
	values.Set("ctl00$ContentPlaceHolder1$txtConoHaLoginPW", r.password)
	values.Set("ctl00$ContentPlaceHolder1$btnLogin", "ログイン")

	req, err = http.NewRequest("POST", "https://cp.conoha.jp/Login.aspx", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://cp.conoha.jp/Login.aspx")

	return req, nil
}

type loginDoResult struct {
}

func (r *loginDoResult) Populate(resp *http.Response, doc *goquery.Document) error {
	pp.Printf("loginDo done.\n")
	return nil
}

// ログイン状態を返す。ログインしていればtrue していなければfalseが返る。
// トップページを取得して、ヘッダー部にアカウントが含まれているかをチェックする
func (cmd *Auth) LoggedIn() (loggedIn bool, err error) {

	r := &loggedInResult{}
	act := &cpanel.Action{
		Request: &loggedInRequest{},
		Result:  r,
	}

	cmd.browser.AddAction(act)
	if err := cmd.browser.Run(); err != nil {
		return false, err
	} else {
		return r.LoggedIn, nil
	}
}

type loggedInRequest struct {
}

func (r *loggedInRequest) NewRequest(values url.Values) (*http.Request, error) {
	return http.NewRequest("GET", "https://cp.conoha.jp/", nil)
}

type loggedInResult struct {
	LoggedIn bool
}

func (r *loggedInResult) Populate(resp *http.Response, doc *goquery.Document) error {
	accountId := doc.Find("#divLoginUser").Text()

	if accountId != "" {
		r.LoggedIn = true
	} else {
		r.LoggedIn = false
	}
	return nil
}

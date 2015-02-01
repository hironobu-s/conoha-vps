package command

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"github.com/hironobu-s/conoha-vps/lib"
	"github.com/howeyc/gopass"
	flag "github.com/ogier/pflag"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func NewLogin() *Login {
	return &Login{
		Command: NewCommand(),
	}
}

type Login struct {
	account  string
	password string

	*Command
}

func (cmd *Login) parseFlag() error {
	var help bool

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	fs.BoolVarP(&help, "help", "h", false, "help")
	fs.StringVarP(&cmd.account, "account", "a", "", "ConoHa Account")
	fs.StringVarP(&cmd.password, "password", "p", "", "ConoHa Password")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		return err
	}

	if help {
		fs.Usage()
		return &ShowUsageError{}
	}

	if cmd.account == "" || cmd.password == "" {

		// コマンドライン引数で指定されていない場合は、標準入力から受け付ける
		if err := cmd.inputAccountInfo(); err != nil {
			return errors.New("Not enough arguments.")
		}
	}

	return nil
}

func (cd *Login) Usage() {
	fmt.Println(`Usage: conoha login [OPTIONS]

DESCRIPTION
    Authenticate an account.
    If account or password not set, you can input interactively.

OPTIONS
    -a: --account:   ConoHa Account.
    -p: --password:  Password.
    -h: --help:      Show usage.  
`)
}

func (cmd *Login) Run() error {
	log := lib.GetLogInstance()

	var err error
	if err = cmd.parseFlag(); err != nil {
		return err
	}

	cmd.config.Account = cmd.account
	cmd.config.Password = cmd.password

	loggedIn, err := cmd.Login()
	if err != nil {
		return err
	}

	if loggedIn {
		log.Infof("Login Successfully.")
	} else {
		log.Infof("Login failed. Enter correct ConoHa account ID and password.")
	}
	return nil
}

// 認証を実行してログイン状態を返す
func (cmd *Login) Login() (loggedIn bool, err error) {
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
func (cmd *Login) inputAccountInfo() error {
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
	return nil
}

// ログイン状態を返す。ログインしていればtrue していなければfalseが返る。
// トップページを取得して、ヘッダー部にアカウントが含まれているかをチェックする
func (cmd *Login) LoggedIn() (loggedIn bool, err error) {

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

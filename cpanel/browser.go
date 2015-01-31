package cpanel

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strings"
)

// コントロールパネル上のアクション
type Action struct {
	Request ActionRequester
	Result  ActionResulter
}

// アクションへのリクエストを作成する
type ActionRequester interface {
	// HTTPリクエストを作成する
	NewRequest(values url.Values) (*http.Request, error)
}

// アクションの結果を格納する
// HTMLの場合とJSONの場合、二つのインターフェイスを用意してある。
type ActionResulter interface {
}

type HtmlActionResulter interface {
	// HTTPレスポンスをパースして、結果オブジェクトを作成する
	Populate(resp *http.Response, doc *goquery.Document) error
}

type JsonActionResulter interface {
	// HTTPレスポンスをパースして、結果オブジェクトを作成する
	Populate(resp *http.Response) error
}

func (act *Action) Run(bi *BrowserInfo) (err error) {

	if act.Request == nil || act.Result == nil {
		return errors.New("Some Struct fields of cpanel.Action undefined.")
	}

	// リクエストを作成
	req, err := act.Request.NewRequest(bi.Values)
	if err != nil {
		return err
	}

	// HTTPヘッダをセット
	for key, value := range bi.headers {
		req.Header.Set(key, value)
	}

	// HTTPリクエスト実行
	cli := &http.Client{Jar: bi.cookiejar}
	resp, err := cli.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//_, _ = ioutil.ReadAll(resp.Body)
	//println(string(body))

	dump, _ := httputil.DumpRequest(req, true)
	println(string(dump))

	// if req.URL.String() == "https://cp.conoha.jp/Service/VPS/Del/Confirm.aspx" && req.Method == "POST" {
	// 	dump, _ = httputil.DumpResponse(resp, true)
	// 	println(string(dump))
	// }

	switch r := act.Result.(type) {
	case HtmlActionResulter:
		var doc *goquery.Document
		doc, err = goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return err
		}
		// hiddenパラメータを取得
		bi.Values = act.hiddenParams(doc)

		// パース結果を返す
		return r.Populate(resp, doc)

	case JsonActionResulter:
		return r.Populate(resp)

	default:
		return errors.New("Undefined Result type.")
	}
}

// BrowserInfoにHTMLフォームに共通する "__" で始まるhidden要素を抽出してバインドする
func (act Action) hiddenParams(doc *goquery.Document) url.Values {

	values := url.Values{}

	inputs := doc.Find("INPUT[type=hidden]")
	for i := range inputs.Nodes {
		n := inputs.Eq(i)
		name, exists := n.Attr("name")
		if !exists || strings.Index(name, "__") != 0 {
			//if !exists {
			continue
		}

		value, _ := n.Attr("value")

		values.Add(name, value)
	}
	return values
}

const (
	COOKIE_URL   = "https://cp.conoha.jp/"
	SESSION_NAME = "ASP.NET_SessionId"
)

// Browserの設定情報
type BrowserInfo struct {
	// CookieJar
	cookiejar *cookiejar.Jar

	// ブラウザが送るHTTPヘッダ
	headers map[string]string

	// リクエストに付与されるURL/POSTパラメータ
	Values url.Values
}

func (b *BrowserInfo) InitializeDefault() {
	b.headers = map[string]string{
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.10; rv:34.0) Gecko/20100101 Firefox/34.0",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "ja,en-us;q=0.7,en;q=0.3",
	}
	b.Values = url.Values{}
	b.cookiejar, _ = cookiejar.New(nil)
}

func (b *BrowserInfo) cookieUrl() *url.URL {
	u, _ := url.Parse(COOKIE_URL)
	return u
}

func (b *BrowserInfo) Sid() string {
	url := b.cookieUrl()
	for _, cookie := range b.cookiejar.Cookies(url) {
		if cookie.Name == SESSION_NAME {
			return cookie.Value
		}

	}
	return ""
}

func (b *BrowserInfo) FixSid(sid string) {
	url := b.cookieUrl()

	cookie := &http.Cookie{
		Name:  SESSION_NAME,
		Value: sid,
	}

	b.cookiejar.SetCookies(url, []*http.Cookie{
		cookie,
	})
}

// Webブラウザの代わりにコントロールパネルへアクセスする
type Browser struct {
	// BrowserInfo
	BrowserInfo *BrowserInfo

	// 実行するリクエストのスライス
	actions []*Action
}

func NewBrowser() *Browser {

	info := &BrowserInfo{}
	info.InitializeDefault()

	b := &Browser{
		BrowserInfo: info,
	}
	return b
}

// アクションを追加する
func (b *Browser) AddAction(act *Action) {
	b.actions = append(b.actions, act)
}

// アクションをすべて削除する
func (b *Browser) ClearAction() {
	b.actions = []*Action{}
}

func (b *Browser) Run() error {
	for _, act := range b.actions {

		err := act.Run(b.BrowserInfo)
		if err != nil {
			b.ClearAction()
			return err
		}
	}

	b.ClearAction()
	return nil
}

package command

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"github.com/hironobu-s/conoha-vps/lib"
	flag "github.com/ogier/pflag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type PrivateKey string

type SshKey struct {
	*Vps
	destPath string
	sshKeyNo int
}

func NewSshKey() *SshKey {
	return &SshKey{
		Vps: NewVps(),
	}
}

func (cmd *SshKey) parseFlag() error {
	var help bool

	fs := flag.NewFlagSet("conoha-vps", flag.ContinueOnError)
	fs.Usage = cmd.Usage

	fs.BoolVarP(&help, "help", "h", false, "help")
	fs.StringVarP(&cmd.destPath, "path", "f", "", ``)
	fs.IntVarP(&cmd.sshKeyNo, "sshkey-no", "s", 1, ``)

	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		return err
	}

	if help {
		fs.Usage()
		return &ShowUsageError{}
	}

	if cmd.destPath == "" {
		// デフォルト
		cmd.destPath = "conoha-" + cmd.config.Account + "-" + strconv.Itoa(cmd.sshKeyNo) + ".key"
	}

	return nil
}

func (cd *SshKey) Usage() {
	fmt.Println(`Usage: conoha ssh-key <file> [OPTIONS]

DESCRIPTION
    Download and store SSH Private key.


OPTIONS
    -f: --file:       Local filename the private key is stored.
                      Default is "conoha-{AccountID}.key"

    -s: --sshkey-no:  SSH Key number. Default is 1.
                      If the number of keys one, It wil be ignored.

    -h: --help:       Show usage.      
`)
}

func (cmd *SshKey) Run() error {
	log := lib.GetLogInstance()

	var err error
	if err = cmd.parseFlag(); err != nil {
		return err
	}

	err = cmd.DownloadSshKey(cmd.destPath)
	if err == nil {
		log.Infof(`Download is complete. A private key is stored in "%s".`, cmd.destPath)
		return nil
	} else {
		return err
	}
}

// SSH秘密鍵をダウンロードする
func (cmd *SshKey) DownloadSshKey(destPath string) error {
	var err error
	destPath, err = filepath.Abs(destPath)
	if err != nil {
		return err
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}

	key, err := cmd.SshKey()
	if err != nil {
		return err
	}

	// パーミッションを0600にセットする
	os.Chmod(destPath, 0600)

	if _, err = file.WriteString(string(key)); err != nil {
		return err
	}

	return nil
}

// SSH秘密鍵を取得する
func (cmd *SshKey) SshKey() (PrivateKey, error) {
	var err error

	var act *cpanel.Action

	// 秘密鍵一覧ページを取得して鍵の一覧を取得する
	rt := &sshDownloadFormResult{
		sshKeyNo: cmd.sshKeyNo,
	}
	act = &cpanel.Action{
		Request: &sshDownloadFormRequest{},
		Result:  rt,
	}
	cmd.browser.AddAction(act)

	rtd := &sshDownloadKeyResult{}
	act = &cpanel.Action{
		Request: &sshDownloadKeyRequest{
			formResult: rt,
		},
		Result: rtd,
	}
	cmd.browser.AddAction(act)

	if err = cmd.browser.Run(); err != nil {
		return "", err
	}
	return rtd.SshKey, nil
}

type sshDownloadFormRequest struct {
}

func (r *sshDownloadFormRequest) NewRequest(values url.Values) (*http.Request, error) {
	return http.NewRequest("GET", "https://cp.conoha.jp/Service/VPS/keyPair/", nil)
}

type sshDownloadFormResult struct {
	sshKeyNo   int
	SshKeyName string
}

func (r *sshDownloadFormResult) Populate(resp *http.Response, doc *goquery.Document) error {
	var sel *goquery.Selection
	sel = doc.Find("#ContentPlaceHolder1_ContentPlaceHolder1_gridSSHKeyList .btnIconPrivateKeyDL02")

	i := 0
	for n := range sel.Nodes {
		if i != r.sshKeyNo-1 {
			i++
			continue
		}

		node := sel.Eq(n)
		name, exists := node.Attr("name")
		if exists {
			r.SshKeyName = name
		}
		break
	}

	if r.SshKeyName == "" {
		return errors.New("SSH Key not found.")
	}

	return nil
}

type sshDownloadKeyRequest struct {
	formResult *sshDownloadFormResult
}

func (r *sshDownloadKeyRequest) NewRequest(values url.Values) (*http.Request, error) {
	values.Add(r.formResult.SshKeyName, "Private Key Download")
	values.Add("ctl00$ctl00$ContentPlaceHolder1$ContentPlaceHolder1$hfTargetKey", "")

	req, err := http.NewRequest("POST", "https://cp.conoha.jp/Service/VPS/keyPair/", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

type sshDownloadKeyResult struct {
	SshKey PrivateKey
}

func (r *sshDownloadKeyResult) Populate(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	r.SshKey = PrivateKey(body)
	return nil
}

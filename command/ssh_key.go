package command

import (
	"fmt"
	"github.com/hironobu-s/conoha-vps/cpanel"
	"github.com/hironobu-s/conoha-vps/lib"
	flag "github.com/ogier/pflag"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type PrivateKey string

type SshKey struct {
	*Vps
	destPath string
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
		cmd.destPath = "conoha-" + cmd.config.Account + ".key"
	}

	return nil
}

func (cd *SshKey) Usage() {
	fmt.Println(`Usage: conoha ssh-key <file> [OPTIONS]

DESCRIPTION
    Download and store SSH Private key.


OPTIONS
    -f: --file:  Local filename the private key is stored.
                 Default is "conoha-{AccountID}.key"

    -h: --help:  Show usage.      
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
func (cmd *Vps) DownloadSshKey(destPath string) error {
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
func (cmd *Vps) SshKey() (PrivateKey, error) {

	var vm *Vm
	var err error

	vpsList := NewVpsList()
	servers, err := vpsList.List(false)
	if err != nil {
		return "", err
	}

	for _, v := range servers { // これ他に良いやり方無いかな？？？
		vm = v
		break
	}

	var act *cpanel.Action

	rt := &sshDownloadKeyResult{}
	act = &cpanel.Action{
		Request: &sshDownloadKeyRequest{
			vm: vm,
		},
		Result: rt,
	}
	cmd.browser.AddAction(act)

	if err = cmd.browser.Run(); err != nil {
		return "", err
	}

	return rt.SshKey, nil
}

type sshDownloadKeyRequest struct {
	vm *Vm
}

func (r *sshDownloadKeyRequest) NewRequest(values url.Values) (*http.Request, error) {
	return http.NewRequest("GET", "https://cp.conoha.jp/Service/VPS/Control/DownloadPrivateKey/"+r.vm.Id, nil)
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

package command

import (
	"github.com/hironobu-s/conoha-vps/cpanel"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type PrivateKey string

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

	servers, err := cmd.List(false)
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

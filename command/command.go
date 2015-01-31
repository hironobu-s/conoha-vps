package command

import (
	"github.com/hironobu-s/conoha-vps/cpanel"
	"github.com/hironobu-s/conoha-vps/lib"
)

type ExitCode int

const (
	ExitCodeOK = iota
	ExitCodeNG
)

type Commander interface {
	// コマンドライン引数を処理する
	parseFlag() error

	// コマンド終了時の処理を記述する
	Shutdown()
}

type Command struct {
	config  *lib.Config
	browser *cpanel.Browser
}

// Commandの実行が完了したときに呼ばれる関数。忘れずdeferすること。
// ブラウザのセッションIDを設定ファイルに記録する。
func (c *Command) Shutdown() {
	log := lib.GetLogInstance()

	c.config.Sid = c.browser.BrowserInfo.Sid()
	c.config.Write()

	log.Debug("write: " + c.browser.BrowserInfo.Sid())
}

func NewCommand(commandName string) (Commander, error) {

	log := lib.GetLogInstance()
	log.Debug("start " + commandName)

	var err error

	// Configを作成
	c := &lib.Config{}
	if err = c.Read(); err != nil {
		return nil, err
	}

	// ブラウザを作成してセッションIDをセットする
	browser := cpanel.NewBrowser()
	browser.BrowserInfo.FixSid(c.Sid)

	// コマンドを作成する
	command := &Command{
		config:  c,
		browser: browser,
	}

	var cmd Commander
	switch commandName {
	case "auth":
		cmd = &Auth{
			Command: command,
		}
	case "vps-list":
		fallthrough
	case "vps-stat":
		fallthrough
	case "vps-add":
		fallthrough
	case "ssh-key":
		fallthrough
	case "vps-delete":
		cmd = &Vps{
			Command: command,
		}
	}

	return cmd, nil
}

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

	// // コマンドを実行する
	Run() error

	// コマンドのUsageを表示する
	Usage()

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

func NewCommand() *Command {
	// Configを作成
	c := &lib.Config{}
	c.Read()

	// ブラウザを作成してセッションIDをセットする
	browser := cpanel.NewBrowser()
	browser.BrowserInfo.FixSid(c.Sid)

	// コマンドを作成する
	cmd := &Command{
		config:  c,
		browser: browser,
	}
	return cmd
}

// USage()を表示するだけの場合でもErrorを返すことになるので、
// この場合は専用のエラーを返すようにする。
type ShowUsageError struct {
	s string
}

func (e ShowUsageError) Error() string {
	return e.s
}

package lib

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

const (
	CONFIGFILE = ".conoha-vps"
)

type Config struct {
	Account  string
	Password string
	Sid      string
}

func (c *Config) ConfigFilePath() string {
	user, err := user.Current()
	if err == nil {
		path := user.HomeDir + string(filepath.Separator) + CONFIGFILE
		return filepath.Clean(path)
	} else {
		// ここに来ることはなさそうだが、そのときはカレントディレクトリに作成する
		return CONFIGFILE
	}
}

func (c *Config) Remove() {
	var err error

	path := c.ConfigFilePath()
	if _, err = os.Stat(path); err != nil {
		// ファイルが存在しない場合は何もしない
		return
	}
	os.Remove(path)
}

func (c *Config) Read() {
	var err error

	path := c.ConfigFilePath()

	if _, err = os.Stat(path); err != nil {
		// ファイルが存在しない場合は何もしない
		return
	}

	file, err := os.Open(path)
	if err != nil {
		// 設定ファイルを開けない場合は何もしない
		return
	}

	dec := json.NewDecoder(file)
	if err := dec.Decode(&c); err != nil {
		// 設定ファイルが正しくない
		return
	}
}

func (c *Config) Write() error {
	path := c.ConfigFilePath()
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	if err = os.Chmod(path, 0600); err != nil {
		return err
	}

	enc := json.NewEncoder(file)
	err = enc.Encode(c)
	if err != nil {
		return err
	}
	return nil
}

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

func (c *Config) Read() error {
	var err error

	path := c.ConfigFilePath()

	if _, err = os.Stat(path); err != nil {
		// ファイルが存在しない場合は何もしない
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(file)
	if err := dec.Decode(&c); err != nil {
		return err
	}

	return nil
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

package lib

import (
	"encoding/json"
	"github.com/mitchellh/go-homedir"
	"os"
	"path/filepath"
)

var Version string

const (
	CONFIGFILE = ".conoha-vps"
)

type Config struct {
	Account  string
	Password string
	Sid      string
}

func (c *Config) ConfigFilePath() (string, error) {
	homedir, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	return homedir + string(filepath.Separator) + CONFIGFILE, nil
}

func (c *Config) Remove() {
	var err error

	path, _ := c.ConfigFilePath()

	if _, err = os.Stat(path); err != nil {
		// ファイルが存在しない場合は何もしない
		return
	}

	os.Remove(path)
}

func (c *Config) Read() {
	var err error

	path, _ := c.ConfigFilePath()

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
	var err error
	path, err := c.ConfigFilePath()
	if err != nil {
		return err
	}

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

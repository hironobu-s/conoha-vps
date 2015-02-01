package command

import (
	"github.com/hironobu-s/conoha-vps/lib"
)

type Version struct {
	*Command
}

func NewVersion() *Version {
	return &Version{
		Command: NewCommand(),
	}
}

func (cmd *Version) parseFlag() error {
	return nil
}

func (cmd *Version) Usage() {
}

func (cmd *Version) Run() error {
	println(lib.VERSION)
	return nil
}

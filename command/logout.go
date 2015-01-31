package command

type Logout struct {
	*Command
}

func NewLogout() *Logout {
	return &Logout{
		Command: NewCommand(),
	}
}

func (cmd *Logout) parseFlag() error {
	return nil
}

func (cmd *Logout) Run() error {
	cmd.config.Remove()
	return nil
}

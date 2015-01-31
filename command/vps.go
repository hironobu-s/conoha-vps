package command

import (
	"time"
)

func NewVps() *Vps {
	return &Vps{
		Command: NewCommand(),
	}
}

type Vps struct {
	info *VpsAddInformation
	*Command
}

func (cmd *Vps) parseFlag() error {
	return nil
}

func (cmd *Vps) Run() error {
	return nil
}

// VPSのステータス
type ServerStatus int

const (
	StatusRunning       = 1  // 稼働中
	StatusOffline       = 4  // 停止
	StatusInUse         = 6  // 取得中
	StatusInFormulation = 8  // サービス準備中
	StatusNoinformation = 98 // 未取得
	StatusUnknown       = 99
)

func (s ServerStatus) String() string {
	switch s {
	case StatusRunning:
		return "Running"
		//return "稼働中"
	case StatusOffline:
		return "Offline"
		//return "停止"
	case StatusInUse:
		return "No status"
		//return "取得中"
	case StatusInFormulation:
		return "Preparing"
		//return "サービス準備中"
	case StatusNoinformation:
		return "-"
		//return "未取得"
	case StatusUnknown:
		fallthrough
	default:
		return "Unknown"
		//return "不明"
	}
}

// 単一VPSを表す構造体
// ServiceStatusとServerStatusは別物であることに注意
type Vm struct {
	Id            string
	ServerStatus  ServerStatus
	Label         string
	ServiceStatus string
	ServiceId     string
	Plan          string
	CreatedAt     time.Time
	DeleteDate    time.Time
	PaymentSpan   string

	// 詳細情報
	NumCpuCore        string
	Memory            string
	Disk1Size         string
	Disk2Size         string
	IPv4              string
	IPv4netmask       string
	IPv4gateway       string
	IPv4dns1          string
	IPv4dns2          string
	IPv6              []string
	IPv6prefix        string
	IPv6gateway       string
	IPv6dns1          string
	IPv6dns2          string
	House             string
	CommonServerId    string
	SerialConsoleHost string
	IsoUploadHost     string
}

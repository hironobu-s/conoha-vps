package command

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Vps struct {
	info *VpsAddInformation
	*Command
}

func (cmd *Vps) parseFlag() error {
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

func (vm *Vm) String() string {

	padding := 14

	r := reflect.Indirect(reflect.ValueOf(vm))
	t := r.Type()

	for i := 0; i < r.NumField(); i++ {
		f := t.Field(i)
		if len(f.Name) > padding {
			padding = len(f.Name)
		}
	}

	format := "%" + strconv.Itoa(padding) + "s: "

	lines := []string{}

	for i := 0; i < r.NumField(); i++ {
		name := t.Field(i).Name

		if value, ok := r.Field(i).Interface().([]string); ok {
			for j := 0; j < len(value); j++ {
				if j == 0 {
					lines = append(lines, fmt.Sprintf(format+"%s", name, value[j]))
				} else {
					lines = append(lines, fmt.Sprintf("%"+strconv.Itoa(padding)+"s  %s", "", value[j]))
				}
			}

		} else if value, ok := r.Field(i).Interface().(time.Time); ok {
			if value.Year() < 2000 {
				lines = append(lines, fmt.Sprintf(format+"%s", name, "-"))
			} else {
				lines = append(lines, fmt.Sprintf(format+"%s", name, value.Format(time.RFC1123)))
			}

		} else {
			lines = append(lines, fmt.Sprintf(format+"%s", name, r.Field(i).String()))
		}
	}

	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

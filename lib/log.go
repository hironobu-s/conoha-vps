package lib

import (
	"bytes"
	"fmt"
	"github.com/Sirupsen/logrus"
	"sort"
)

func init() {
}

var instance *logrus.Logger

func GetLogInstance() *logrus.Logger {
	if instance == nil {
		// ログレベルの設定
		instance = logrus.New()

		instance.Level = logrus.InfoLevel
		instance.Formatter = &SimpleFormatter{}
	}
	return instance
}

type SimpleFormatter struct {
}

func (f *SimpleFormatter) appendKeyValue(b *bytes.Buffer, key, value interface{}) {
	switch value.(type) {
	case string, error:
		fmt.Fprintf(b, "%v=%q ", key, value)
	default:
		fmt.Fprintf(b, "%v=%v ", key, value)
	}
}

func (f *SimpleFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	var keys []string
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := &bytes.Buffer{}

	fmt.Fprintf(b, "%s", entry.Message)

	b.WriteByte('\n')
	return b.Bytes(), nil

}

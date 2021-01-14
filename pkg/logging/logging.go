package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"path"
	"runtime"
)

// Logger global object for logging across the pkg/
var Logger = logrus.New()

func init() {

	Logger.SetLevel(logrus.TraceLevel)
	formatter := &logrus.TextFormatter{
		ForceColors:   false,
		FullTimestamp: true,
	}
	if Logger.GetLevel() >= logrus.DebugLevel {
		Logger.SetReportCaller(true)
		formatter.CallerPrettyfier = callerPrettyfier
	}
	Logger.Formatter = formatter

}

func callerPrettyfier(f *runtime.Frame) (string, string) {
	filename := path.Base(f.File)
	return "", fmt.Sprintf(" [%s:%d]", filename, f.Line)
}

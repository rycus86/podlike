package connect

import (
	"log"
	"os"
)

const (
	LogLevel_DEBUG int = iota
	LogLevel_INFO  int = iota
	LogLevel_WARN  int = iota
	LogLevel_ERROR int = iota
	LogLevel_NONE  int = iota
)

var (
	level  = LogLevel_INFO
	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
)

func SetLogLevel(newLevel int) {
	level = newLevel
}

func (cp *connectionPair) debug(v ...interface{}) {
	if level <= LogLevel_DEBUG {
		cp.emitLog("[DEBUG]", v)
	}
}

func (cp *connectionPair) info(v ...interface{}) {
	if level <= LogLevel_INFO {
		cp.emitLog("[INFO ]", v)
	}
}

func (cp *connectionPair) warn(v ...interface{}) {
	if level <= LogLevel_WARN {
		cp.emitLog("[WARN ]", v)
	}
}

func (cp *connectionPair) error(v ...interface{}) {
	if level <= LogLevel_ERROR {
		cp.emitLog("[ERROR]", v)
	}
}

func (cp *connectionPair) emitLog(prefix string, v []interface{}) {
	logger.Println(append([]interface{}{cp.logPrefix, prefix}, v...)...)
}

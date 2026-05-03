package log

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var logFile *os.File
var Verbose bool

// -----------------------------------------
// public

func Close() error {
	if logFile == nil {
		return nil
	}
	return logFile.Close()
}

func Command(name string, args ...string) *exec.Cmd {
	cmdStr := name + " " + strings.Join(args, " ")
	fmt.Fprintf(logFile, "[cmd] %s\n", cmdStr)
	return exec.Command(name, args...)
}

func Debug(format string, args ...interface{}) {
	if !Verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
}

func GetLogWriter() io.Writer {
	return logFile
}

func OpenLog() error {
	logPath := filepath.Join("/var/log", "declpac.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = f
	writeTimestamp()
	return nil
}

func Write(msg []byte) {
	logFile.Write(msg)
}

// -----------------------------------------
// private

func writeTimestamp() {
	ts := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf("\n--- %s ---\n", ts)
	logFile.Write([]byte(header))
}

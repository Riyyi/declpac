package state

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var logFile *os.File

func OpenLog() error {
	logPath := filepath.Join("/var/log", "declpac.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = f
	return nil
}

func GetLogWriter() io.Writer {
	return logFile
}

func Write(msg []byte) {
	PrependWithTimestamp(logFile, msg)
}

func Close() error {
	if logFile == nil {
		return nil
	}
	return logFile.Close()
}

func PrependWithTimestamp(w io.Writer, msg []byte) {
	ts := time.Now().Format("2006-01-02 15:04:05")
	header := fmt.Sprintf("\n--- %s ---\n", ts)
	w.Write([]byte(header))
	w.Write(msg)
}

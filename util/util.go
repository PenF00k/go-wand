package util

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"strings"
)

func Close(c io.Closer, outFile string) {
	err := c.Close()
	if err != nil {
		log.Errorf("failed to close file %s", outFile)
	}
}

func CloseWatcher(w *fsnotify.Watcher) {
	err := w.Close()
	if err != nil {
		log.Errorf("failed to close watcher")
	}
}

func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().String()
		idx := strings.LastIndex(localAddr, ":")
		return localAddr[0:idx], nil
	}

	return "", err
}

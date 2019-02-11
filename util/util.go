package util

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"io"
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

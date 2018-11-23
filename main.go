package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
)

func main() {
	var src string
	var output string
	var goOut string

	flag.StringVar(&src, "src", "", "directory name")
	flag.StringVar(&output, "jsOut", "", "js output directory")
	flag.StringVar(&goOut, "goOut", "", "go output directory")
	devMode := flag.Bool("dev", false, "do not use gomobile")
	jsPort := flag.Int("jsDevPort", 9009, "port for js")
	watchMode := flag.Bool("watch", false, "watch mode")

	flag.Parse()

	goutput := path.Join(goOut, "mobile")
	if _, err := os.Stat(goutput); os.IsNotExist(err) {
		os.Mkdir(goutput, os.ModePerm)
	}

	if !*watchMode {
		if len(src) > 0 {
			Parse(src, output, goutput, *devMode, *jsPort)
		}
	} else {
		if len(src) > 0 {
			Parse(src, output, goutput, *devMode, *jsPort)
			watchGo(src, output, goutput, *devMode, *jsPort)
		}
	}
}

func watchGo(src string, output string, goutput string, devMode bool, port int) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					Parse(src, output, goutput, devMode, port)

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(src)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

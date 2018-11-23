package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	var config string

	// flag.Bool("init", false, "init config file")
	devMode := flag.Bool("dev", false, "do not use gomobile")
	jsPort := flag.Int("jsDevPort", 9009, "port for js")
	watchMode := flag.Bool("watch", false, "watch mode")
	init := flag.Bool("init", false, "init")

	flag.StringVar(&config, "config", "auto.yaml", "directory name")

	flag.Parse()

	if *init {
		log.Printf("New config stored.")
		storeConfig()
		return
	}

	readConfig(config, *watchMode, *jsPort, *devMode)
}

type Deploy struct {
	MobileSource string
	Js           string
	AutoGin      bool
	Remote       bool
	Watch        bool
	Port         int16
}

type RunConfiguration struct {
	Source string
	Deploy Deploy
}

func storeConfig() {
	configuration := RunConfiguration{}
	configuration.Source = "./"
	configuration.Deploy.Js = "./js"
	configuration.Deploy.MobileSource = "./mobile"
	configuration.Deploy.AutoGin = true
	configuration.Deploy.Watch = true
	configuration.Deploy.Remote = true
	configuration.Deploy.Port = 3000

	out, _ := yaml.Marshal(configuration)
	// Use os.Create to create a file for writing.
	f, err := os.Create("auto.yaml")
	if err != nil {
		log.Errorf("not able to create auto.yaml")
		return
	}

	defer f.Close()
	// Create a new writer.
	w := bufio.NewWriter(f)

	// Write a string to the file.
	w.Write(out)
	// Flush.
	w.Flush()

	{
		// Use os.Create to create a file for writing.
		shell, err := os.Create(".run.sh")
		if err != nil {
			log.Errorf("not able to create auto.yaml")
			return
		}

		defer shell.Close()
		// Create a new writer.
		wsh := bufio.NewWriter(shell)

		// Write a string to the file.
		wsh.Write([]byte("#/bin/bash\ngin -a 9009 run call.go\n"))
		// Flush.
		wsh.Flush()
	}

	os.Chmod(".run.sh", 0777)
}

func readConfig(config string, watchMode bool, port int, dev bool) {
	data, err := ioutil.ReadFile(config)
	if err != nil {
		log.Errorf("Can not open config file")
		return
	}

	configuration := RunConfiguration{}
	err = yaml.Unmarshal([]byte(data), &configuration)
	if err != nil {
		log.Errorf("Can not open config file")

		return
	}

	src := configuration.Source
	output := configuration.Deploy.Js
	remote := configuration.Deploy.Remote || dev
	goutput := configuration.Deploy.MobileSource

	goFullOutDir := path.Join(goutput, "mobile")
	if _, err := os.Stat(goFullOutDir); os.IsNotExist(err) {
		os.Mkdir(goFullOutDir, os.ModePerm)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		os.Mkdir(output, os.ModePerm)
	}

	Parse(src, output, goutput, remote, port)

	if configuration.Deploy.AutoGin && (configuration.Deploy.Watch || watchMode) {
		log.Printf("running gin")
		cmd := exec.Command("open", "-a", "Terminal", "`pwd`")
		// cmd.Dir = goutput
		cmd.Start()
	}

	if configuration.Deploy.Watch || watchMode {
		watchGo(src, output, goFullOutDir, remote, port)
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

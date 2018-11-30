package config

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v2"
)

type Source struct {
	Package string
}

type Js struct {
	Path string
}

type Wrapper struct {
	Package string
	Port    int16
}

type Configuration struct {
	Source  Source
	Wrapper Wrapper
	Js      Js
}

func requestString(name string) string {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Scans a line from Stdin(Console)
		scanner.Scan()
		// Holds the string that scanned
		text := scanner.Text()
		if len(text) != 0 {
			clean := strings.Trim(text, " \t,.")
			if clean != "" {
				return clean
			}
		}
	}
}

func StoreConfig() {
	configuration := Configuration{}

	configuration.Source.Package = requestString("Your go package to wrap: ")
	configuration.Js.Path = requestString("Target JS directory: ")
	configuration.Wrapper.Package = requestString("Wrapper's package name: ")
	configuration.Wrapper.Port = 9009

	out, _ := yaml.Marshal(configuration)
	// Use os.Create to create a file for writing.
	f, err := os.Create("wand.yaml")
	if err != nil {
		log.Errorf("Failed to create gocall.yaml")
		return
	}

	defer f.Close()
	// Create a new writer.
	w := bufio.NewWriter(f)
	// Write a string to the file.
	w.Write(out)
	// Flush.
	w.Flush()
}

func ReadConfig(config string) (*Configuration, error) {
	data, err := ioutil.ReadFile(config)
	if err != nil {
		log.Errorf("Can not open config file")
		return nil, err
	}

	configuration := Configuration{}
	err = yaml.Unmarshal([]byte(data), &configuration)
	if err != nil {
		log.Errorf("Can not open config file")
		return nil, err
	}

	// goFullOutDir := path.Join(goutput, "mobile")
	// if _, err := os.Stat(goFullOutDir); os.IsNotExist(err) {
	// 	os.Mkdir(goFullOutDir, os.ModePerm)
	// }

	// if _, err := os.Stat(output); os.IsNotExist(err) {
	// 	os.Mkdir(output, os.ModePerm)
	// }

	// Parse(src, output, goFullOutDir, remote, port)

	// if configuration.Deploy.AutoGin && (configuration.Deploy.Watch || watchMode) {
	// 	log.Printf("running gin")
	// 	cmd := exec.Command("open", "-a", "iterm", "`pwd`")
	// 	// cmd.Dir = goutput
	// 	cmd.Start()
	// }

	// if configuration.Deploy.Watch || watchMode {
	// 	watchGo(src, output, goFullOutDir, remote, port)
	// }

	return &configuration, nil
}

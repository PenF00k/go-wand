package main

import (
	"go/build"
	"os"
	"path"

	"gitlab.vmassive.ru/gocallgen/config"
	"gitlab.vmassive.ru/gocallgen/generator"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/fsnotify/fsnotify"
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "gocall.yaml",
			Usage: "configuration file",
		},
	}

	app.Name = "gocallgen"
	app.Usage = "link beetween go and js"
	app.Action = func(c *cli.Context) error {
		runApplication(c.String("config"), true)
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:        "init",
			Aliases:     []string{"i"},
			Usage:       "create new config files",
			Description: "This is how we describe describeit the function",
			Action: func(c *cli.Context) error {
				config.StoreConfig()
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}

	return gopath
}

func runApplication(configName string, dev bool) {
	configuration, err := config.ReadConfig(configName)
	if err != nil {
		return
	}

	goPath := getGoPath()

	fullGoSourcePath := path.Join(goPath, "src", configuration.Source.Package)
	targetGoCallPath := path.Join(goPath, "src", configuration.Wrapper.Package)

	createDirectory(targetGoCallPath)
	createDirectory(configuration.Js.Path)

	pathMap := generator.PathMap{
		Source: fullGoSourcePath,
		Target: targetGoCallPath,
		Js:     configuration.Js.Path,
	}

	goPackageName := configuration.Wrapper.Package
	if dev {
		goPackageName = "main"
	}

	codeList := &generator.CodeList{
		Package:       goPackageName,
		Dev:           dev,
		Port:          configuration.Wrapper.Port,
		SourcePackage: configuration.Source.Package,
		PathMap:       pathMap,
		Config:        configuration,
	}

	watchGo(codeList)
}

func watchGo(codeList *generator.CodeList) {
	Parse(codeList)

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
					Parse(codeList)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(codeList.PathMap.Source)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func createDirectory(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, os.ModePerm)
	}
}

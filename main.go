package main

import (
	"gitlab.vmassive.ru/wand/generator"
	"gitlab.vmassive.ru/wand/parser"
	"gitlab.vmassive.ru/wand/util"
	"go/build"
	"os"
	"os/signal"
	"path"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"gitlab.vmassive.ru/wand/config"
	"gitlab.vmassive.ru/wand/reload"

	"github.com/fsnotify/fsnotify"
)

//go-assets-builder  ./templates -o ./assets/assets.go -p assets
//protoc --dart_out="./generated" client.proto timestamp.proto

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "wand.yaml",
			Usage: "configuration file",
		},
		cli.BoolFlag{
			Name:  "release",
			Usage: "contruct release",
		},
		cli.BoolFlag{
			Name:  "genDartDto",
			Usage: "generate only dart dto classes (for instance for shared library)",
		},
	}

	app.Name = "wand"
	app.Usage = "magic link between go and js"
	app.Action = func(c *cli.Context) error {
		runApplication(c.String("config"), !c.Bool("release"), c.Bool("genDartDto"))
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

func runApplication(configName string, dev bool, genDartDto bool) {
	configuration, err := config.ReadConfig(configName)
	if err != nil {
		return
	}

	goPath := getGoPath()

	fullGoSourcePath := path.Join(goPath, "src", configuration.Source.Package)
	targetGoCallPath := path.Join(goPath, "src", configuration.Wrapper.Package)

	createDirectory(targetGoCallPath)
	createDirectory(configuration.Js.Path)

	protoPackageName := "proto_client"

	protoPath := path.Join(targetGoCallPath, protoPackageName)
	protoRelPath := path.Join(configuration.Wrapper.Package, protoPackageName)
	createDirectory(protoPath)

	workingDir, err := os.Getwd()
	if err != nil {
		log.Errorf("couldn't get current working dir")
	}

	flutterGeneratedPathRelative := configuration.Flutter.GeneratedProtoPath
	flutterGeneratedPath := path.Join(workingDir, flutterGeneratedPathRelative)
	createDirectory(flutterGeneratedPath)

	javaGeneratedPathRelative := configuration.Java.GeneratedProtoPath
	javaGeneratedPath := path.Join(workingDir, javaGeneratedPathRelative)
	createDirectory(javaGeneratedPath)

	pathMap := generator.PathMap{
		Source:              fullGoSourcePath,
		Target:              targetGoCallPath,
		Js:                  configuration.Js.Path,
		FlutterGenerated:    flutterGeneratedPath,
		FlutterGeneratedRel: flutterGeneratedPathRelative,
		Proto:               protoPath,
		JavaGenerated:       javaGeneratedPath,
		ProtoRel:            protoRelPath,
	}

	packageMap := generator.PackageMap{
		ProtoPackageName:  protoPackageName,
		FlutterAppPackage: configuration.Flutter.AppPackage,
	}

	goPackageName := configuration.Wrapper.Package
	if dev {
		goPackageName = "main"
	}

	codeList := &generator.CodeList{
		Package:       goPackageName,
		Dev:           dev,
		ServerIp:      getIP(),
		Port:          configuration.Wrapper.Port,
		SourcePackage: configuration.Source.Package,
		Config:        configuration,
		PathMap:       pathMap,
		PackageMap:    packageMap,
	}

	if genDartDto {
		parser.Parse(codeList, genDartDto)
		return
	}

	if dev {
		watchGo(codeList)
	} else {
		parser.Parse(codeList, false)
	}
}

func watchGo(codeList *generator.CodeList) {
	parser.Parse(codeList, false)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer util.CloseWatcher(watcher)

	rel, err := reload.New(codeList)
	if err != nil {
		return
	}

	shutdown(rel)

	go rel.Run()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					done <- true
					return
				}
				ext := path.Ext(event.Name)
				if ext != ".go" {
					continue
				}

				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Remove == fsnotify.Remove ||
					event.Op&fsnotify.Rename == fsnotify.Rename ||
					event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("modified file:", event.Name)
					parser.Parse(codeList, false)
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
		e := os.MkdirAll(dir, os.ModePerm)
		if e != nil {
			log.Errorf("error while creating dir for path '%s': %v", dir, e)
		}
	}
}

func shutdown(runner *reload.LiveReload) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		log.Println("Got signal: ", s)
		err := runner.Kill()
		if err != nil {
			log.Print("Error killing: ", err)
		}
		os.Exit(1)
	}()
}

func getIP() string {
	ip, err := util.GetOutboundIP()
	if err != nil {
		log.Errorf("couldn't get server ip address: %v", err)
	}

	return ip
}

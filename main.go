package main

import (
	"flag"
	"os"
	"path"
)

func main() {
	var src string
	var output string
	var goOut string

	flag.StringVar(&src, "src", "", "directory name")
	flag.StringVar(&output, "jsOut", "", "js output directory")
	flag.StringVar(&output, "goOut", "", "js output directory")

	flag.Parse()

	goutput := path.Join(goOut, "mobile")
	if _, err := os.Stat(goutput); os.IsNotExist(err) {
		os.Mkdir(goutput, os.ModePerm)
	}

	if len(src) > 0 {
		Parse(src, output, goutput)
	}
}

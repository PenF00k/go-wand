package reload

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	"gitlab.vmassive.ru/gocallgen/generator"
)

type LiveReload struct {
	bin       string
	stop      chan struct{}
	codeList  *generator.CodeList
	command   *exec.Cmd
	writer    io.Writer
	starttime time.Time
	errors    string
}

func New(codeList *generator.CodeList) (*LiveReload, error) {
	bin := path.Join(codeList.PathMap.Target, "livecall")

	rerload := LiveReload{
		bin:       bin,
		stop:      make(chan struct{}),
		codeList:  codeList,
		starttime: time.Now(),
		writer:    ioutil.Discard,
	}

	return &rerload, nil
}

func (reload *LiveReload) Build() error {
	args := append([]string{"go", "build", "-o", reload.bin})

	var command *exec.Cmd
	command = exec.Command(args[0], args[1:]...)

	command.Dir = reload.codeList.PathMap.Target
	output, err := command.CombinedOutput()

	if command.ProcessState.Success() {
		reload.errors = ""
	} else {
		reload.errors = string(output)
	}

	if len(reload.errors) > 0 {
		return fmt.Errorf(reload.errors)
	}

	return err
}

func (reload *LiveReload) Stop() {
	reload.stop <- struct{}{}
}

func (reload *LiveReload) buildAndRun() {
	err := reload.Build()
	if err == nil {
		reload.RunBuild()
	} else {
		log.Errorln("build failed")
		fmt.Println(reload.errors)
		buildErrors := strings.Split(reload.errors, "\n")
		for _, err := range buildErrors {
			log.Errorf(err)
		}
	}

}

func (reload *LiveReload) Run() error {
	reload.SetWriter(os.Stdout)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	reload.buildAndRun()

	log.Printf("start walking")
	err = filepath.Walk(reload.codeList.PathMap.Source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				if path != reload.codeList.PathMap.Source {
					watcher.Add(path)
				}
			}

			return nil
		})

	if err != nil {
		return err
	}

	watcher.Add(reload.codeList.PathMap.Target)

	for {
		select {
		case <-reload.stop:
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			log.Println("event:", event)

			fi, err := os.Stat(event.Name)
			if err == nil {

				if fi.Mode().IsDir() {
					if event.Op&fsnotify.Rename == fsnotify.Rename ||
						event.Op&fsnotify.Create == fsnotify.Create {
						watcher.Add(event.Name)
					}

					continue
				}
			}

			ext := path.Ext(event.Name)
			if ext != ".go" {
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					watcher.Remove(event.Name)
				}

				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Rename == fsnotify.Rename ||
				event.Op&fsnotify.Create == fsnotify.Create {
				reload.buildAndRun()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Println("error:", err)
		}
	}
}

func (r *LiveReload) runBin() error {
	r.command = exec.Command(r.bin)
	stdout, err := r.command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := r.command.StderrPipe()
	if err != nil {
		return err
	}

	err = r.command.Start()
	if err != nil {
		return err
	}

	r.starttime = time.Now()

	go io.Copy(r.writer, stdout)
	go io.Copy(r.writer, stderr)
	go r.command.Wait()

	return nil
}

func (r *LiveReload) SetWriter(writer io.Writer) {
	r.writer = writer
}

func (r *LiveReload) Info() (os.FileInfo, error) {
	return os.Stat(r.bin)
}

func (r *LiveReload) needsRefresh() bool {
	info, err := r.Info()
	if err != nil {
		return false
	} else {
		return info.ModTime().After(r.starttime)
	}
}

func (r *LiveReload) Exited() bool {
	return r.command != nil && r.command.ProcessState != nil && r.command.ProcessState.Exited()
}

func (r *LiveReload) RunBuild() (*exec.Cmd, error) {
	if r.needsRefresh() {
		r.Kill()
	}

	if r.command == nil || r.Exited() {
		err := r.runBin()
		if err != nil {
			log.Print("Error running: ", err)
		}
		time.Sleep(250 * time.Millisecond)
		return r.command, err
	} else {
		return r.command, nil
	}
}

func (r *LiveReload) Kill() error {
	if r.command != nil && r.command.Process != nil {
		done := make(chan error)
		go func() {
			r.command.Wait()
			close(done)
		}()

		//Trying a "soft" kill first
		if runtime.GOOS == "windows" {
			if err := r.command.Process.Kill(); err != nil {
				return err
			}
		} else if err := r.command.Process.Signal(os.Interrupt); err != nil {
			return err
		}

		//Wait for our process to die before we return or hard kill after 3 sec
		select {
		case <-time.After(3 * time.Second):
			if err := r.command.Process.Kill(); err != nil {
				log.Println("failed to kill: ", err)
			}
		case <-done:
		}
		r.command = nil
	}

	return nil
}

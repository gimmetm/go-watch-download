package fileworker

import (
	"context"
	"fmt"
	log "github.com/gimmetm/go-run-download/pkg/logging"
	"gopkg.in/fsnotify.v1"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Fswatcher struct {
	Watcher *fsnotify.Watcher
	FileMap map[string]chan fsnotify.Op
}

const WAIT_TIME = 3 * time.Second

func New() *Fswatcher {

	fileMap := make(map[string]chan fsnotify.Op)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Logger.Fatal(err)
	}

	fs := &Fswatcher{
		Watcher: watcher,
		FileMap: fileMap,
	}
	homeDir, err := os.UserHomeDir()
	downloadPath := fmt.Sprintf("%s/Downloads", homeDir)
	if err := fs.Watcher.Add(downloadPath); err != nil {
		log.Logger.Fatal(err)
	} else {
		log.Logger.Debugf("%s Added.", downloadPath)
	}
	return fs
}

func (fs *Fswatcher) Start(ctx context.Context, wg *sync.WaitGroup) {

	log.Logger.Infoln("Start FsWatcher")
	go func(wg *sync.WaitGroup) {

		defer fs.Watcher.Close()
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				log.Logger.Infoln("Stop FsWatcher : context kill signal was sent")
				return
			case evt, ok := <-fs.Watcher.Events:
				if !ok {
					return
				}
				// filter file(.ica)
				if !strings.HasSuffix(evt.Name, ".ica") {
					continue
				}

				log.Logger.Debugf("[FileEvent:%s]%s", evt.Op.String(), evt.Name)

				switch op := evt.Op; {
				case op&fsnotify.Create == fsnotify.Create:
					fi, err := os.Stat(evt.Name)
					if err != nil {
						log.Logger.Errorf("[FE:Create] %v\n", err.Error())
						continue
					}
					if fi.Mode().IsDir() {
						if err := fs.Watcher.Add(evt.Name); err != nil {
							log.Logger.Errorln(err)
						} else {
							log.Logger.Debugf("%s Added.", evt.Name)
						}
					} else {
						// 파일 Create 타이머 시작
						newCh := make(chan fsnotify.Op)
						fs.FileMap[evt.Name] = newCh
						go fs.FileTimer(evt.Name, newCh)
					}
				case op&fsnotify.Write == fsnotify.Write:
					fi, err := os.Stat(evt.Name)
					if err != nil {
						log.Logger.Errorf("[FE:Write] %v\n", err.Error())
						continue
					}
					if !fi.Mode().IsDir() {
						if ch, ok := fs.FileMap[evt.Name]; ok {
							ch <- fsnotify.Write
						} else {
							// 파일 Write 타이머 시작
							newCh := make(chan fsnotify.Op)
							fs.FileMap[evt.Name] = newCh
							go fs.FileTimer(evt.Name, newCh)
						}
					}
				case op&fsnotify.Remove == fsnotify.Remove || op&fsnotify.Rename == fsnotify.Rename:
					if ch, ok := fs.FileMap[evt.Name]; ok {
						ch <- op
					}
				default:
					log.Logger.Debugf("Ignore[%s : %s]", evt.Name, evt.Op.String())
				}
			case err, ok := <-fs.Watcher.Errors:
				if !ok {
					break
				}
				log.Logger.Error("error:", err)
			default:
			}
		}
		log.Logger.Infoln("Stop FsWatcher : Loop Out")
	}(wg)

}

func (fs *Fswatcher) FileTimer(filepath string, ch chan fsnotify.Op) {

	defer func() {
		if v := recover(); v != nil {
			err, _ := v.(error)
			log.Logger.Errorf("[Painc] on FileTimer Error :  %v", err.Error())
		}
	}()

	log.Logger.Debugf("[File Timer Start] %s", filepath)
	timer := time.NewTimer(WAIT_TIME)
LOOP:
	for {
		select {
		case op := <-ch:
			switch op {
			case fsnotify.Write:
				log.Logger.Debugf("[File Timer Reset] %s", filepath)
				timer.Reset(WAIT_TIME)
			case fsnotify.Rename:
				fallthrough
			case fsnotify.Remove:
				log.Logger.Debugf("[File Removed] %s", filepath)
				timer.Stop()
				close(ch)
				delete(fs.FileMap, filepath)
				break LOOP
			}
		case <-timer.C:
			log.Logger.Debugf("[File Timer End] %s", filepath)
			cmd := exec.Command("open", filepath)
			err := cmd.Start()
			if err != nil {
				log.Logger.Fatal(err)
			}

			close(ch)
			delete(fs.FileMap, filepath)
			break LOOP
		}
	}
}

func (fs *Fswatcher) AddPath(root string) {
	log.Logger.Infof("Add Watch directory and subdirectories : %s", root)

	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fi, err := os.Stat(path)
			if err != nil {
				return err
			}
			if fi.Mode().IsDir() {
				if err := fs.Watcher.Add(path); err != nil {
					return err
				} else {
					log.Logger.Debugf("%s Added.", path)
				}
			}
			return nil
		})
	if err != nil {
		log.Logger.Error(err)
	}
}

func (fs *Fswatcher) DelPath(root string) {
	log.Logger.Infof("[Directory Removed] %s\n", root)

	err := filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			fi, err := os.Stat(path)
			if err != nil {
				return err
			}
			if fi.Mode().IsDir() {
				if err := fs.Watcher.Remove(path); err != nil {
					return err
				} else {
					log.Logger.Debugf("[Stop Watching] %s", path)
				}
			}
			return nil
		})
	if err != nil {
		log.Logger.Error(err)
	}
}

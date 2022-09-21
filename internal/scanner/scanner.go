package scanner

import (
	"context"
	"fmt"
	"folder-scan/internal"
	"github.com/enriquebris/goconcurrentqueue"
	"log"
	"os"
	"path"
	"sync/atomic"
	"time"
)

type Scanner struct {
	cfg         Config
	counter     int64
	takeQueue   chan *internal.FolderInfo
	bufferQueue goconcurrentqueue.Queue

	done chan bool

	workersRunning int32
}

func (app *Scanner) scanFolder(folder *internal.FolderInfo) {
	atomic.AddInt32(&app.workersRunning, 1)

	files, err := os.ReadDir(folder.Data.Path)
	if err != nil {
		log.Println("Failed to scan folder by path=", folder.Data.Path)
		log.Println(err.Error())
	}

	for _, file := range files {
		filePath := path.Join(folder.Data.Path, file.Name())
		if file.IsDir() {
			subFolderInfo := internal.FolderInfo{
				Data: &internal.FSData{
					Name:   file.Name(),
					Path:   filePath,
					Parent: folder,
				},
				Files: make([]*internal.FileInfo, 0, 2), // TODO test capacities
			}
			folder.Subs = append(folder.Subs, &subFolderInfo)

			err := app.bufferQueue.Enqueue(&subFolderInfo)
			if err != nil {
				panic(err)
			}
		} else {
			osFileInfo, err := file.Info()
			if err == nil && osFileInfo != nil {
				fileInfo := internal.FileInfo{
					Name:   file.Name(),
					Path:   filePath,
					Size:   osFileInfo.Size(),
					Parent: folder,
				}

				folder.Files = append(folder.Files, &fileInfo)
				folder.FilesSize += fileInfo.Size
			} else {
				log.Println("Failed to get info about file by path=", filePath)
				if err != nil {
					log.Println(err.Error())
				}
			}
		}
	}

	atomic.AddInt32(&app.workersRunning, -1)
}

func (app *Scanner) runWorker(ctx context.Context) {
	for {
		select {
		case folder := <-app.takeQueue:
			app.scanFolder(folder)
		case <-ctx.Done():
			return
		case <-app.done:
			return
		}
	}
}

func (app *Scanner) ManageQueues() {
	for {
		e, err := app.bufferQueue.DequeueOrWaitForNextElement()
		if err != nil {
			panic(err)
		}
		if e == nil {
			return
		}

		app.takeQueue <- e.(*internal.FolderInfo)
	}
}

func (app *Scanner) WaitForScanning() {
	cnt := 0
	for {
		time.Sleep(1000)
		if app.bufferQueue.GetLen() == 0 && atomic.LoadInt32(&app.workersRunning) == 0 {
			cnt++
			if cnt == 3 {
				app.done <- true

				// to stop ManageQueues()
				err := app.bufferQueue.Enqueue(nil)
				if err != nil {
					panic(err)
				}

				return
			}
		} else {
			cnt = 0
		}
	}
}

func (app *Scanner) Start(ctx context.Context) *internal.FolderInfo {
	fmt.Println("Starting scan.")
	fmt.Println("Root dir=", app.cfg.RootDir)
	fmt.Println("Save report to=", app.cfg.SaveTo)
	fmt.Printf("Using %d workers\n", app.cfg.Workers)

	for i := 0; i < app.cfg.Workers; i++ {
		go app.runWorker(ctx)
	}

	rootInfo := internal.FolderInfo{
		Data: &internal.FSData{
			Name: path.Base(app.cfg.RootDir),
			Path: app.cfg.RootDir,
		},
	}
	err := app.bufferQueue.Enqueue(&rootInfo)
	if err != nil {
		panic(err)
	}

	go app.ManageQueues()

	app.WaitForScanning()

	return &rootInfo
}

func NewScanner(cfg Config) *Scanner {
	return &Scanner{
		cfg:         cfg,
		counter:     0,
		takeQueue:   make(chan *internal.FolderInfo, 128),
		bufferQueue: goconcurrentqueue.NewFIFO(),
		done:        make(chan bool),
	}
}

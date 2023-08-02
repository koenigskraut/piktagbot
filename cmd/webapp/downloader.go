package main

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"log"
	"os"
	"strconv"
	"sync"
)

type downloadedMap struct {
	m map[int64]bool
	sync.RWMutex
}

func (dm *downloadedMap) init(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		n, err := strconv.ParseInt(entry.Name(), 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		dm.add(n)
	}
}

func (dm *downloadedMap) add(documentID int64) {
	dm.Lock()
	dm.m[documentID] = true
	dm.Unlock()
}

func (dm *downloadedMap) peek(documentID int64) bool {
	dm.RLock()
	defer dm.RUnlock()
	return dm.m[documentID]
}

var downloaded = downloadedMap{m: make(map[int64]bool)}

type receiveFiles struct {
	files []*tg.InputDocumentFileLocation
	ch    chan string
}

var downloadChan = make(chan *receiveFiles)

func downloadLoop(ch chan *receiveFiles, ctx context.Context, tgClient *tg.Client) {
	d := downloader.NewDownloader()
	for r := range ch {
		go func(rr *receiveFiles) {
			for _, file := range rr.files {
				if !downloaded.peek(file.ID) {
					_, err := d.Download(tgClient, file).
						ToPath(ctx, fmt.Sprintf("%s/%d", stickerPath, file.ID))
					if err != nil {
						rr.ch <- "404"
						continue
					}
					downloaded.add(file.ID)
				}
				rr.ch <- strconv.FormatInt(file.ID, 10)
			}
			close(rr.ch)
		}(r)
	}
}

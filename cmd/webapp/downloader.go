package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
	"io"
	"log"
	"os"
	"os/exec"
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

type InputDocumentMimeTyped struct {
	mimeType database.MimeType
	doc      *tg.InputDocumentFileLocation
}

type receiveFiles struct {
	files []InputDocumentMimeTyped
	ch    chan string
}

var downloadChan = make(chan *receiveFiles)

func downloadLoop(ch chan *receiveFiles, ctx context.Context, tgClient *tg.Client) {
	d := downloader.NewDownloader()
	for r := range ch {
		go downloadAll(ctx, tgClient, d, r)
	}
}

const prefix = `slv`

func downloadAll(ctx context.Context, api *tg.Client, d *downloader.Downloader, r *receiveFiles) {
	defer close(r.ch)
	for _, file := range r.files {
		if !downloaded.peek(file.doc.ID) {
			webpPath := fmt.Sprintf("%s/%d", stickerPath, file.doc.ID)
			switch file.mimeType {
			case database.MimeTypeWebp:
				_, err := d.Download(api, file.doc).ToPath(ctx, webpPath)
				if err != nil {
					r.ch <- "404"
					continue
				}
			case database.MimeTypeWebm:
				webmPath := fmt.Sprintf("%s/webm/%d", stickerPath, file.doc.ID)
				_, err := d.Download(api, file.doc).ToPath(ctx, webmPath)
				if err != nil {
					r.ch <- "404"
					continue
				}
				// yeah.. i know
				err = exec.Command("ffmpeg", "-y", "-vcodec", "libvpx-vp9", "-i", webmPath,
					"-vframes", "1", "-f", "webp", webpPath).Run()
				if err != nil {
					r.ch <- "404"
					continue
				}
			case database.MimeTypeTgs:
				tgsPath := fmt.Sprintf("%s/tgs/%d.tgs", stickerPath, file.doc.ID)
				_, err := d.Download(api, file.doc).ToPath(ctx, tgsPath)
				if err != nil {
					r.ch <- "404"
					continue
				}
				fileIn, err := os.Open(tgsPath)
				if err != nil {
					r.ch <- "404"
					continue
				}
				reader, err := gzip.NewReader(fileIn)
				if err != nil {
					r.ch <- "404"
					continue
				}
				jsonPath := fmt.Sprintf("%s/%d", stickerPath, file.doc.ID)
				fileOut, err := os.Create(jsonPath)
				if err != nil {
					r.ch <- "404"
					continue
				}
				if _, err := io.Copy(fileOut, reader); err != nil {
					r.ch <- "404"
					continue
				}
				fileOut.Close()
				reader.Close()
				fileIn.Close()
			}
			downloaded.add(file.doc.ID)
		}
		r.ch <- string(prefix[file.mimeType]) + strconv.FormatInt(file.doc.ID, 10)
	}
}

package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gobwas/ws/wsutil"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
)

type downloadedMap struct {
	m map[int64]bool
	sync.RWMutex
}

func (dm *downloadedMap) init(filePath string) {
	for _, suffix := range [][2]string{{"", ".webp"}, {"json", ".json"}} {
		entries, err := os.ReadDir(path.Join(filePath, suffix[0]))
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasSuffix(name, suffix[1]) {
				continue
			}
			n, err := strconv.ParseInt(name[:len(name)-len(suffix[1])], 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			dm.add(n)
		}
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
	files  []InputDocumentMimeTyped
	output *wsutil.Writer
	ch     chan error
}

var downloadChan = make(chan *receiveFiles)

func downloadLoop(ch chan *receiveFiles, ctx context.Context, tgClient *tg.Client) {
	d := downloader.NewDownloader()
	for r := range ch {
		go downloadAll(ctx, tgClient, d, r)
	}
}

const webpPathFmt = "%s/%d.webp"
const webmPathFmt = "%s/webm/%d.webm"
const tgsPathFmt = "%s/tgs/%d.tgs"
const jsonPathFmt = "%s/json/%d.json"

var pathFmtInitial = []string{webpPathFmt, tgsPathFmt, webmPathFmt}
var pathFmtFinal = []string{webpPathFmt, jsonPathFmt, webpPathFmt}

func writeHeader(w io.Writer, file InputDocumentMimeTyped) error {
	if _, err := w.Write([]byte{byte(file.mimeType)}); err != nil {
		return err
	}
	return binary.Write(w, binary.LittleEndian, file.doc.ID)
}

func downloadWebp(ctx context.Context, api *tg.Client,
	doc *tg.InputDocumentFileLocation, d *downloader.Downloader,
	path string, out io.Writer,
) error {
	fileWebp, err := os.Create(path)
	if err != nil {
		return err
	}
	writerWebp := io.MultiWriter(out, fileWebp)
	_, err = d.Download(api, doc).Stream(ctx, writerWebp)
	return err
}

func downloadTgs(ctx context.Context, api *tg.Client,
	doc *tg.InputDocumentFileLocation, d *downloader.Downloader,
	pathInitial, pathFinal string, out io.Writer,
) error {
	bufTgs := &bytes.Buffer{}
	bufTgsWriter := bufio.NewWriterSize(bufTgs, 256*1024)
	fileTgs, err := os.Create(pathInitial)
	if err != nil {
		return err
	}
	tgsWriter := io.MultiWriter(bufTgsWriter, fileTgs)
	if _, err := d.Download(api, doc).Stream(ctx, tgsWriter); err != nil {
		return err
	}
	if err := bufTgsWriter.Flush(); err != nil {
		return err
	}
	if err := fileTgs.Close(); err != nil {
		return err
	}

	gzReader, err := gzip.NewReader(bufTgs)
	if err != nil {
		return err
	}
	fileJson, err := os.Create(pathFinal)
	if err != nil {
		return err
	}
	jsonWriter := io.MultiWriter(out, fileJson)
	if _, err := io.Copy(jsonWriter, gzReader); err != nil {
		return err
	}
	if err := fileJson.Close(); err != nil {
		return err
	}
	return nil
}

func downloadWebm(ctx context.Context, api *tg.Client,
	doc *tg.InputDocumentFileLocation, d *downloader.Downloader,
	pathInitial, pathFinal string, out io.Writer,
) error {
	_, err := d.Download(api, doc).ToPath(ctx, pathInitial)
	if err != nil {
		return err
	}
	// yeah.. i know
	err = exec.Command("ffmpeg", "-y", "-vcodec", "libvpx-vp9", "-i", pathInitial,
		"-vframes", "1", "-f", "webp", pathFinal).Run()
	if err != nil {
		return err
	}
	fileWebp, err := os.Open(pathFinal)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, fileWebp)
	return err
}

func downloadAll(ctx context.Context, api *tg.Client, d *downloader.Downloader, r *receiveFiles) {
	defer close(r.ch)
	writer := r.output
	for _, file := range r.files {
		var err error

		if err := writeHeader(writer, file); err != nil {
			r.ch <- err
			return
		}

		finalPath := fmt.Sprintf(pathFmtFinal[file.mimeType], stickerPath, file.doc.ID)
		// if file exist, write to channel writer
		if downloaded.peek(file.doc.ID) {
			localFile, err := os.Open(finalPath)
			if err != nil {
				log.Println(err)
				continue
			}
			if _, err := io.Copy(writer, localFile); err != nil {
				r.ch <- err
				return
			}
			_ = localFile.Close()
			if err := writer.Flush(); err != nil {
				r.ch <- err
				return
			}
			continue
		}

		initialPath := fmt.Sprintf(pathFmtInitial[file.mimeType], stickerPath, file.doc.ID)
		switch file.mimeType {
		case database.MimeTypeWebp:
			err = downloadWebp(ctx, api, file.doc, d, initialPath, writer)
		case database.MimeTypeTgs:
			err = downloadTgs(ctx, api, file.doc, d, initialPath, finalPath, writer)
		case database.MimeTypeWebm:
			err = downloadWebm(ctx, api, file.doc, d, initialPath, finalPath, writer)
		default:
			err = errors.New("unknown mime type")
		}
		if err != nil {
			log.Println(err)
			continue
		}
		downloaded.add(file.doc.ID)
		if err := writer.Flush(); err != nil {
			r.ch <- err
			return
		}
	}
	r.ch <- nil
}

package origin

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/hls-on-the-fly/internal/manifest"
	"github.com/timohahaa/hls-on-the-fly/internal/mp4"
)

func (o *Origin) masterM3U8(w http.ResponseWriter, r *http.Request) {
	var (
		fileName       = chi.URLParam(r, "filename")
		isEncrypted, _ = strconv.ParseBool(r.URL.Query().Get("encrypt"))
	)

	assets, err := o.storage.GetAllAssets(fileName)
	if err != nil {
		log.Errorf("[origin] (filename=%v) %v", fileName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m3u, err := manifest.Master(assets, o.domain+fileName, isEncrypted)
	if err != nil {
		log.Errorf("[origin] (filename=%v) %v", fileName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/vnd.apple.mpegurl")
	w.WriteHeader(http.StatusOK)
	w.Write(m3u)
}

func (o *Origin) mediaM3U8(w http.ResponseWriter, r *http.Request) {
	var (
		quality        = chi.URLParam(r, "quality")
		fileName       = chi.URLParam(r, "filename")
		isEncrypted, _ = strconv.ParseBool(r.URL.Query().Get("encrypt"))
	)

	asset, err := o.storage.GetFileAsset(fileName, quality)
	if err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var (
		fileURL           = o.domain + fileName + "/" + quality + "/chunk.mp4"
		keyURL            = o.domain + "key"
		reader  io.Reader = asset
	)

	// encryption is done on-the-fly
	if isEncrypted {
		pipeR, pipeW := io.Pipe()
		reader = pipeR

		go func() {
			defer pipeW.Close()

			err := mp4.Encrypt(asset, pipeW, mp4.EncryptParams{
				KeyID:  o.kid,
				Key:    o.key,
				IVHex:  o.ivHex,
				Scheme: "cbcs",
			})
			if err != nil {
				log.Errorf("[origin] (filename=%v, quality=%v) encrypt: %v", fileName, quality, err)
				pipeW.CloseWithError(err) // pipeR will see the error
				return
			}

		}()
	}

	m3u, err := manifest.Media(reader, manifest.MediaParams{
		FileURL:     fileURL,
		KeyURL:      keyURL,
		IvHex:       o.ivHex,
		IsEncrypted: isEncrypted,
	})
	if err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/vnd.apple.mpegurl")
	w.WriteHeader(http.StatusOK)
	w.Write(m3u)
}

func (o *Origin) chunk(w http.ResponseWriter, r *http.Request) {
	var (
		quality        = chi.URLParam(r, "quality")
		fileName       = chi.URLParam(r, "filename")
		isEncrypted, _ = strconv.ParseBool(r.URL.Query().Get("encrypt"))
	)

	from, to, err := parseRange(r.Header.Get("Range"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	asset, err := o.storage.GetFileAsset(fileName, quality)
	if err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if isEncrypted {
		tmpDir := filepath.Join(os.TempDir(), fileName, quality, strconv.FormatInt(from, 10))

		if err := os.MkdirAll(filepath.Dir(tmpDir), os.ModePerm); err != nil {
			log.Errorf("[origin] (filename=%v, quality=%v) encrypt: %v", fileName, quality, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tmpFile, err := os.Create(tmpDir)
		if err != nil {
			log.Errorf("[origin] (filename=%v, quality=%v) encrypt: %v", fileName, quality, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = mp4.Encrypt(asset, tmpFile, mp4.EncryptParams{
			KeyID:  o.kid,
			Key:    o.key,
			IVHex:  o.ivHex,
			Scheme: "cbcs",
		})
		if err != nil {
			log.Errorf("[origin] (filename=%v, quality=%v) encrypt: %v", fileName, quality, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		asset = tmpFile
	}

	info, err := asset.Stat()
	if err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) stat: %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = asset.Seek(from, io.SeekStart); err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) seek: %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Add("Content-Length", strconv.FormatInt(to-from+1, 10))
	w.Header().Add("Content-Range", fmt.Sprintf("%v-%v/%v", from, to, info.Size()))

	_, _ = io.CopyN(w, asset, to-from+1)
}

func (o *Origin) serveKey(w http.ResponseWriter, _ *http.Request) {
	b, _ := o.key.MarshalBinary()
	w.Write(b)
}

func parseRange(hdr string) (from, to int64, err error) {
	hdr = strings.TrimPrefix(hdr, "bytes=")
	parts := strings.Split(hdr, "-")

	if len(parts) != 2 {
		return 0, 0, errors.New("invalid range header")
	}

	from, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}

	to, err = strconv.ParseInt(parts[1], 10, 64)
	return
}

package origin

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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

	m3u, err := manifest.Master(assets, "http://"+o.server.Addr+"/"+fileName, isEncrypted)
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
		fileURL = "http://" + o.server.Addr + "/" + fileName + "/" + quality + "/chunk.mp4"
		keyURL  = "http://" + o.server.Addr + "/key"
	)
	m3u, err := manifest.Media(asset, manifest.MediaParams{
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
		q              = r.URL.Query()
		from, _        = strconv.ParseInt(q.Get("from"), 10, 64)
		size, _        = strconv.ParseInt(q.Get("size"), 10, 64)
		isEncrypted, _ = strconv.ParseBool(r.URL.Query().Get("encrypt"))
	)

	asset, err := o.storage.GetFileAsset(fileName, quality)
	if err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if isEncrypted {
		tmpDir := filepath.Join(os.TempDir(), fileName)
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

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Add("Content-Length", strconv.FormatInt(size, 10))
	_, _ = asset.Seek(from, io.SeekStart)
	_, _ = io.CopyN(w, asset, size)
}

func (o *Origin) serveKey(w http.ResponseWriter, _ *http.Request) {
	w.Write(o.key[:])
}

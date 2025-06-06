package origin

import (
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

func (o *Origin) mediaM3U8(w http.ResponseWriter, r *http.Request) {
	var (
		quality  = chi.URLParam(r, "quality")
		fileName = chi.URLParam(r, "filename")
	)

	_, err := o.storage.GetFileAsset(fileName, quality)
	if err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Origin) chunk(w http.ResponseWriter, r *http.Request) {
	var (
		quality  = chi.URLParam(r, "quality")
		fileName = chi.URLParam(r, "filename")
		q        = r.URL.Query()
		from, _  = strconv.ParseInt(q.Get("from"), 10, 64)
		to, _    = strconv.ParseInt(q.Get("to"), 10, 64)

		chunkSize = to - from
	)

	asset, err := o.storage.GetFileAsset(fileName, quality)
	if err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Add("Content-Length", strconv.FormatInt(chunkSize, 10))
	_, _ = asset.Seek(from, io.SeekStart)
	_, _ = io.CopyN(w, asset, chunkSize)
}

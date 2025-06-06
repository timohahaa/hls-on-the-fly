package origin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

func (o *Origin) mediaM3U8(w http.ResponseWriter, r *http.Request) {
	var (
		quality  = chi.URLParam(r, "quality")
		fileName = chi.URLParam(r, "filename")
	)

	asset, err := o.storage.GetFileAsset(fileName, quality)
	if err != nil {
		log.Errorf("[origin] (filename=%v, quality=%v) %v", fileName, quality, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Origin) chunk(w http.ResponseWriter, r *http.Request) {

}

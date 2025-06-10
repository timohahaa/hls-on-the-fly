package origin

import (
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/hls-on-the-fly/internal/storage"
)

type Origin struct {
	mux     *chi.Mux
	server  http.Server
	signal  chan os.Signal
	storage *storage.Storage

	// hardcoded, because this is a proof-of-concept
	key   uuid.UUID
	kid   uuid.UUID
	ivHex string
}

func New(addr string) (*Origin, error) {
	var (
		mux = chi.NewMux()
		o   = &Origin{
			mux: mux,
			server: http.Server{
				Addr:    addr,
				Handler: mux,
			},
			signal: make(chan os.Signal),
			key:    uuid.MustParse("4047b82f-a25e-4a58-8c2b-116dbcf81660"),
			kid:    uuid.MustParse("613aa5a4-cb22-491f-89b8-583a5432046a"),
			ivHex:  strings.ReplaceAll(uuid.MustParse("9184bb2f-cf97-4226-b3a3-8ed8f8b3fe2e").String(), "-", ""),
		}
		err error
	)

	if o.storage, err = storage.New(); err != nil {
		return nil, err
	}

	o.route()

	return o, nil
}

func (o *Origin) route() {
	o.mux.Use(cors.AllowAll().Handler)
	o.mux.Route("/{filename}", func(mux chi.Router) {
		mux.Get("/master.m3u8", o.masterM3U8)
		mux.Route("/{quality:(360|480|720|1080|audio)}", func(mux chi.Router) {
			mux.Get("/media.m3u8", o.mediaM3U8)
			mux.Get("/chunk.mp4", o.chunk)
		})
	})

	o.mux.Get("/key", o.serveKey)
}

func (o *Origin) Run() error {
	var (
		signals = []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGKILL,
		}
	)

	log.Infof("[origin] HTTP server listening on %s", o.server.Addr)
	go func() {
		log.Fatal(o.server.ListenAndServe())
	}()

	signal.Notify(o.signal, signals...)
	signal := <-o.signal
	log.Infof("[origin] got signal: %s", signal)

	o.server.Close()

	return nil

}

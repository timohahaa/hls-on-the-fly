package origin

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/hls-on-the-fly/internal/storage"
)

type Origin struct {
	mux     *chi.Mux
	server  http.Server
	signal  chan os.Signal
	storage *storage.Storage
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
	o.mux.Route("/{filename}/{quality:(360|480|720|1080)}", func(mux chi.Router) {
		mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})
		mux.Get("/media.m3u8", o.mediaM3U8)
		mux.Get("/chunk.mp4", o.chunk)
	})
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

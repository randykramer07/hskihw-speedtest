package website

import (
	"crypto/tls"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/go-systemd/activation"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/pires/go-proxyproto"
	log "github.com/sirupsen/logrus"

	web "./empty"
	web "./garbage"
	"github.com/randykramer07/hskihw-speedtest"
)

const ( // Constante variabele voor de aantal data die wordt verzonden per seconde
	chunkGrootte = 1048576 // Chunk berekend aan de hand van de maximale formule = 1024 bytes * 1024 byts = 1048576

)

var (
	randomizedData = getRandomData(chunkGrootte) // Verkrijg verschillende random data pakketten voor de test, dit voorkomt vertraging in het proces
)

var (
	standaardAssets embed.FS
)

func SpeedTest(configuratie *configuratie.Configuratie) error {
	r := chi.NewRouter() // Maak een nieuwe router in Chi package
	r.Use(middleware.RealIP) // Zorg dat publiek IP van client bekend wordt bij applicatie
	r.Use(middleware.GetHead)

	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Locaties die gebruikt mogen worden, (*) betekend alles.
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "HEAD"}, // HTTP methodes die gebruikt mogen worden
		AllowedHeaders: []string{"*"}, // Welke type headers er gebruikt mogen worden, (*) betekend alles
	})

	r.Use(cs.Handler) // Gebruikt Handler voor cors
	r.Use(middleware.NoCache) // Middleware die maximaal aantal HTTP headers insteld om zo te voorkomen dat de router wordt opgeslagen in cache.
	r.use(middleware.Recoverer) // Middleware die panics recoverd, logged en terugherleid

	var assets http.FileSystem // Variabele voor website bestanden

	if fi, err := os.Stat(configuratie.AssetsPath); os.IsNotExist(err) || !fi.IsDir() {
		log.Warnf("De ingestelde folder voor de assets bestaat niet, of is geen folder")
		sub, err := fs.Sub(standaardAssets, "assets")
		if err != nil {
			log.FatalF("Er is een fout opgetreden bij het openen van de standaard bestanden: %s", err)
		}
		assets = http.FS(sub)
	} else {
		assets = justFilesFilesystem{fs: http.Dir(configuratie.AssetsPath), readDirBatchSize: 2}
	}

	r.HandleFunc("/empty", empty)
	r.HandleFunc("/backend/empty", empty)
	r.HandleFunc("/garbage", garbage)
	r.HandleFunc("/backend/garbage", garbage)


	// PHP Frontend standaard bestanden
	r.HandleFunc("/backend/empty.php", empty)
	r.HandleFunc("/backend/garbage.php", garbage)
	r.Get("/garbage.php", garbage)
	r.Get("/backend/garbage.php", garbage)

	listeners, err := activation.Listeners()
	if err != nil {
		log.Fatalf("Error whilst checking for systemd socket activation %s", err)
	}

	var s error

	switch len(listeners) {
	case 0:
		addr := net.JoinHostPort(configuratie.BindAddress, configuratie.Port)
		log.Infof("Starting backend server on %s", addr)

		// TLS and HTTP/2.
		if configuratie.EnableTLS {
			log.Info("Use TLS connection.")
			if !(configuratie.EnableHTTP2) {
				srv := &http.Server{
					Addr:         addr,
					Handler:      r,
					TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
				}
				s = srv.ListenAndServeTLS(configuratie.TLSCertFile, configuratie.TLSKeyFile)
			} else {
				s = http.ListenAndServeTLS(addr, configuratie.TLSCertFile, configuratie.TLSKeyFile, r)
			}
		} else {
			if configuratie.EnableHTTP2 {
				log.Errorf("TLS is mandatory for HTTP/2. Ignore settings that enable HTTP/2.")
			}
			s = http.ListenAndServe(addr, r)
		}
	case 1:
		log.Info("Starting backend server on inherited file descriptor via systemd socket activation")
		if configuratie.BindAddress != "" || configuratie.Port != "" {
			log.Errorf("Both an address/port (%s:%s) has been specificed in the config AND externally configured socket activation has been detected", configuratie.BindAddress, configuratie.Port)
			log.Fatal(`Please deconfigure socket activation (e.g. in systemd unit files), or set both 'bind_address' and 'listen_port' to ''`)
		}
		s = http.Serve(listeners[0], r)
	default:
		log.Fatalf("Asked to listen on %s sockets via systemd activation.  Sorry we currently only support listening on 1 socket.", len(listeners))
	}
	return s
}


func empty(w http.ResponseWriter, r *http.Request) {
	web.empty() // Gebruik de functie empty uit empty.go
}

func garbage(w http.ResponseWriter, r *http.Request) {
	web.garbage()
}
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

	"github.com/librespeed/speedtest/results"
	"github.com/randykramer07/hskihw-speedtest/config"
	log "github.com/sirupsen/logrus"
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

func SpeedTest(configuratie *config.Configuratie) error {
	r := chi.NewRouter()     // Maak een nieuwe router in Chi package
	r.Use(middleware.RealIP) // Zorg dat publiek IP van client bekend wordt bij applicatie
	r.Use(middleware.GetHead)

	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},                              // Locaties die gebruikt mogen worden, (*) betekend alles.
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "HEAD"}, // HTTP methodes die gebruikt mogen worden
		AllowedHeaders: []string{"*"},                              // Welke type headers er gebruikt mogen worden, (*) betekend alles
	})

	r.Use(cs.Handler)           // Gebruikt Handler voor cors
	r.Use(middleware.NoCache)   // Middleware die maximaal aantal HTTP headers insteld om zo te voorkomen dat de router wordt opgeslagen in cache.
	r.Use(middleware.Recoverer) // Middleware die panics recoverd, logged en terugherleid

	var assetFS http.FileSystem // Variabele voor website bestanden

	if fi, err := os.Stat(configuratie.AssetsPath); os.IsNotExist(err) || !fi.IsDir() {
		log.Warnf("De ingestelde folder voor de assets bestaat niet, of is geen folder")
		sub, err := fs.Sub(standaardAssets, "assets")
		if err != nil {
			log.Fatalf("Er is een fout opgetreden bij het openen van de standaard bestanden: %s", err)
		}
		assetFS = http.FS(sub)
	} else {
		assetFS = justFilesFilesystem{fs: http.Dir(configuratie.AssetsPath), readDirBatchSize: 2}
	}

	r.Get("/*", pages(assetFS))
	r.HandleFunc("/empty", empty)
	r.HandleFunc("../backend/empty", empty)
	r.HandleFunc("/garbage", garbage)
	r.HandleFunc("../backend/garbage", garbage)
	r.Get("/results", results.DrawPNG)
	r.Get("/results/", results.DrawPNG)
	r.Get("/backend/results", results.DrawPNG)
	r.Get("/backend/results/", results.DrawPNG)

	// PHP Frontend standaard bestanden
	r.HandleFunc("../backend/empty.php", empty)
	r.HandleFunc("../backend/garbage.php", garbage)
	r.Get("/garbage.php", garbage)
	r.Get("../backend/garbage.php", garbage)

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

func pages(fs http.FileSystem) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/" {
			r.RequestURI = "/index.html"
		}

		http.FileServer(fs).ServeHTTP(w, r)
	}

	return fn
}

func empty(w http.ResponseWriter, r *http.Request) {
	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()

	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
}

func garbage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Description", "File Transfer")                   // Beschrijving die mee wordt gegeven in de header
	w.Header().Set("Content-Type", "application/octet-stream")               // Type van de header
	w.Header().Set("Content-Disposition", "attachment; filename=random.dat") // Bestandsnaam waar de data tijdelijk in wordt opgeslagen
	w.Header().Set("Content-Transfer-Encoding", "binary")                    // Encoding van de header

	// Grootte van de chunks staat altijd op 4
	chunks := 4

	ckGrootte := r.FormValue("ckGrootte")
	if ckGrootte != "" {
		i, err := strconv.ParseInt(ckGrootte, 10, 64)
		if err != nil {
			log.Errorf("Ongeldig aantal Chunks: %s", ckGrootte)   // Error als er meer chunks zijn ingegeven dan maximaal aantal
			log.Warnf("We gebruiken de standaard van %d", chunks) // Melding dat de standaard aantal gebruikt gaat worden
		} else {
			if i > 1024 { // Maximale chunk limiet van 1024 chunks
				chunks = 1024
			} else {
				chunks = int(i)
			}
		}
	}

	for i := 0; i < chunks; i++ { // For loop om aantal chunks te beoaken
		if _, err := w.Write(randomizedData); err != nil { // Indien error geef melding
			log.Errorf("Er is een fout opgetreden bij het ophalen van de chunks %d: %s", i, err)
			break
		}
	}
}

func getIP(w http.ResponseWriter, r *http.Request) {
	var ret results.Result

	clientIP := r.RemoteAddr
	clientIP = strings.ReplaceAll(clientIP, "::ffff:", "")

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		clientIP = ip
	}

	isSpecialIP := true
	switch {
	case clientIP == "::1":
		ret.ProcessedString = clientIP + " - localhost IPv6 access"
	case strings.HasPrefix(clientIP, "fe80:"):
		ret.ProcessedString = clientIP + " - link-local IPv6 access"
	case strings.HasPrefix(clientIP, "127."):
		ret.ProcessedString = clientIP + " - localhost IPv4 access"
	case strings.HasPrefix(clientIP, "10."):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case regexp.MustCompile(`^172\.(1[6-9]|2\d|3[01])\.`).MatchString(clientIP):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case strings.HasPrefix(clientIP, "192.168"):
		ret.ProcessedString = clientIP + " - private IPv4 access"
	case strings.HasPrefix(clientIP, "169.254"):
		ret.ProcessedString = clientIP + " - link-local IPv4 access"
	case regexp.MustCompile(`^100\.([6-9][0-9]|1[0-2][0-7])\.`).MatchString(clientIP):
		ret.ProcessedString = clientIP + " - CGNAT IPv4 access"
	default:
		isSpecialIP = false
	}

	if isSpecialIP {
		b, _ := json.Marshal(&ret)
		if _, err := w.Write(b); err != nil {
			log.Errorf("Error writing to client: %s", err)
		}
		return
	}

	getISPInfo := r.FormValue("isp") == "true"
	distanceUnit := r.FormValue("distance")

	ret.ProcessedString = clientIP

	if getISPInfo {
		ispInfo := getIPInfo(clientIP)
		ret.RawISPInfo = ispInfo

		removeRegexp := regexp.MustCompile(`AS\d+\s`)
		isp := removeRegexp.ReplaceAllString(ispInfo.Organization, "")

		if isp == "" {
			isp = "Unknown ISP"
		}

		if ispInfo.Country != "" {
			isp += ", " + ispInfo.Country
		}

		if ispInfo.Location != "" {
			isp += " (" + calculateDistance(ispInfo.Location, distanceUnit) + ")"
		}

		ret.ProcessedString += " - " + isp
	}

	render.JSON(w, r, ret)
}

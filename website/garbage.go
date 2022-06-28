package garbage

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"github.com/randykramer07/hskihw-speedtest/website"
)

func garbage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=random.dat")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	// Aantal chunks om te gebruiken voor de Test, standaard = 4
	chunks := 4

	ckGrootte := r.FormValue("ckGrootte")
	if ckSize != "" {
		i, err := strconv.ParseInt(ckGrootte, 10, 64)
		if err != nil {
			log.Errorf("Ongeldig aantal chunks: %s", ckGrootte)
			log.Warnf("Er wordt gebruik gemaakt van het standaard aantal chunks: %d", chunks)
		} else {
			// Zet de maximale chunk size op 1024
			if i > 1024 {
				chunks = 1024
			} else {
				chunks = int(i)
			}
		}
	}

	for i := 0; i < chunks; i++ {
		if _, err := w.Write(randomizedData); err != nil {
			log.Errorf("Error writing back to client at chunk number %d: %s", i, err)
			break
		}
	}
}
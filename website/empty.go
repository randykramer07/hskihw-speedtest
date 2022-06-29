package empty

import (
	"io"
	"io/ioutil"
	"net/http"
)

func empty(w http.ResponseWriter, r *http.Request) {
	_, error := io.Copy(ioutil.Discard, r.Body)
	if error != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = r.Body.Close()
	w.Header().Set("Connection", "keep-alive") // Permanente HTTP verbindings header
	w.WriteHeader(http.StatusOK)
}

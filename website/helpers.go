package web

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/umahmood/haversine"

	"github.com/randykramer07/hskihw-speedtest/config"
)

var (
	serverCoord haversine.Coord
)

func getRandomData(length int) []byte {
	data := make([]byte, length)
	if _, err := rand.Read(data); err != nil {
		log.Fatalf("Het is niet gelukt om een random dataset te genereren: %s", err)
	}
	return data
}

func SetServerLocation(configuratie *configuratie.Configuratie) { // Stelt de serverlocatie in aan de hand van vooraf ingestelde coördinaten
	if configuratie.ServerLatitude != 0 || configuratie.ServerLongtitude != 0 {
		log.Infof("Ingestelde Server: %.6f, %.6f", configuratie.ServerLatitude, configuratie.ServerLongtitude)
		serverCoord.Lat = configuratie.ServerLatitude
		serverCoord.Lon = configuratie.ServerLongtitude
		return
	}

	var ret results.IPInfoResponse
	resp, err := http.DefaultClient.Get(getIPInfoURL(""))
	if err != nil {
		log.Errorf("Error getting repsonse from ipinfo.io: %s", err)
		return
	}
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response from ipinfo.io: %s", err)
		return
	}
	defer resp.Body.Close()

	if err := json.Unmarshal(raw, &ret); err != nil {
		log.Errorf("Error parsing response from ipinfo.io: %s", err)
		return
	}

	if ret.Location != "" {
		serverCoord, err = parseLocationString(ret.Location)
		if err != nil {
			log.Errorf("Cannot get server coordinates: %s", err)
			return
		}
	}

	log.Infof("Fetched server coordinates: %.6f, %.6f", serverCoord.Lat, serverCoord.Lon)
}

func parseLocationString(location string) (haversine.Coord, error) {
	var coord haversine.Coord

	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		err := fmt.Errorf("unknown location format: %s", location)
		log.Error(err)
		return coord, err
	}

	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		log.Errorf("Error parsing latitude: %s", parts[0])
		return coord, err
	}

	lng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		log.Errorf("Error parsing longitude: %s", parts[0])
		return coord, err
	}

	coord.Lat = lat
	coord.Lon = lng

	return coord, nil
}
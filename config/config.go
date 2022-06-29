package config

import (
	log "github.com/sirupsen/logrus" //Logsysteem voor applicatie
	"github.com/spf13/viper"         // Package die goed te gebruiken is bij het gebruik van config files zoals JSON, TOML, TAML, ENV en INI
)

type Configuratie struct { // Type .. struct wordt gebruikt om een simple configuratie om te zetten in bruikbare configuratie voor de applicatie

	// Main Applicatie configuratie, waarop communiceert de applicatie en waar zijn de testservers.
	BindAddress      string  `yaml:"bind.address"`      // BindAddress is voor de keuze van de interface van de applicatie
	Port             string  `yaml:"listen.port"`       // Poort waarop de applicatie luistert
	ServerLatitude   float64 `yaml:"server.latitude"`   // Latitude GPS Locatie van de testserver (Om afstand te bepalen)
	ServerLongtitude float64 `yaml:"server.longtitude"` // Longtitude GPS Locatie van de testserver (Om afstand te bepalen)
	IPInfoAPIKey     string  `yaml:"ipinfo.apikey"`     // API Key om gebruik te maken van https://ipinfo.io
	//Website gerelateerde dingen
	AssetsPath string `yaml:"assets.path"` // Locatie waar de HTML bestanden staan van de website

}

var (
	configuratieBestand string
	loadedConfiguratie  *Configuratie = nil // Stelt configuratie in van het geladen configuratiebestand , anders wordt alles op 0 gezet.
)

func init() { // Stel voor viper de standaard gegevens in (Deze worden aangepast aan de hand van de settings.yaml)
	viper.SetDefault("bind.address", "127.0.0.1") // Standaard interface
	viper.SetDefault("listen.port", "9100")       //Standaard poort
	viper.SetDefault("download_chunks", 4)
	viper.SetDefault("enable_cors", false)
	viper.SetDefault("enable_tls", false)
	viper.SetDefault("enable_http2", false)
	viper.SetConfigName("settings") // Naam van het bestand met aangepaste settings
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // Locatie van het settings bestand
}

func Load(configPath string) Configuratie { // Functie om configuratiebestand te laden
	var conf Configuratie

	configuratieBestand = configPath
	viper.SetConfigFile(configPath)

	if error := viper.ReadInConfig(); error != nil { // Errorhandler wanneer er een fout zit in het configuratiebestand.
		if _, ok := error.(viper.ConfigFileNotFoundError); ok {
			log.Warnf("Er is geen configuratiebestand gevonden. standaardwaardes worden gebruikt.") // Log een probleem dat er geen configuratiebestand is gevonden
		} else {
			log.Fatalf("Fout opgetreden bij het lezen van de volgende configuratie: %s", error) // Log een fatale error door 1 van de configuraties
		}
	}

	loadedConfiguratie = &conf
	return conf
}

func LoadedConfig() *Configuratie { // Gebruik de eerder geladen Configuratie in de string configuratieBestand
	if loadedConfiguratie == nil { // Als loadedConfig == nil (oftewel er is geen configuratiebestand geladen), laad dan configuratieBestand
		Load(configuratieBestand)
	}
	return loadedConfiguratie // Veranderd loadedConfig
}

package main 

import (
	"flag"
	"github.com/randykramer07/hskihw-speedtest/config" // Importeer Configuratiemap
	"github.com/randykramer07/hskihw-speedtest/website" // Importeer map met Website gerelateerde bestanden

	log "github.com/sirupsen/logrus" // Importeer Logsysteem
)

var ( // Variabele voor aanwezigheid configuratie, melding voor aanpassing
	optioneleConfiguratie = flag.String("c", "", "Er is een configuratiebestand om te gebruiken, gegevens worden ingesteld aan de hand van settings.yaml bestand")
)

func main() {
	flag.Parse{}
	configuratie := configuratie.Load(*optioneleConfiguratie) // Zet de geladen configuratie om in een lokale variabele
	website.SetServerLocation(&configuratie) // Lees de serverlocatie uit de configuratie
}
![HSKIHW Logo](https://github.com/randykramer07/hskihw-speedtest/blob/main/logo/HSKIHW-wit.png?raw=true)

# HSKIHW SpeedTest
Hoe Snel Kan Ik Hier Weg SpeedTest geschreven in GO en PHP.

Lightweight SpeedTest geimplementeerd met onder andere JavaScript, GO en PHP

**Compatibility**

Alle moderne browsers worden ondersteund waaronder: IE11, Edge, Chrome, Firefox, Safari. Werkt ook op mobiel.

**Wat kan de SpeedTest?**

   - Downloadsnelheid meten
   - Uploadsnelheid meten
   - Ping meten
   - IP Adres, ISP, afstand van de server (optional)

**Server requirements**

   Alle systemen die GO ondersteunen

**Installatie**

Je hebt de nieuwste Go versie nodig om de applicatie te compilen.

    Installeer Go 1.18.3

    $ go get golang.org/dl/go1.18.3
    # Uitgaande van standaard GOPATH map (~/go), Go 1.18.3 wordt geinstalleerd in ~/go/bin
    $ ~/go/bin/go1.18.3 version
    go version go1.18.3 linux/amd64

    Clone mijn repository:

    $ git clone github.com/randykramer07/hskihw-speedtest

    Build

    # Verander je working directory naar de map van de repository
    $ cd hskihw-speedtest
    # Compile
    $ go build -ldflags "-w -s" -trimpath -o speedtest main.go

    Zet de assets folder, settings.yaml en de gecompilde speedtest binary in een losse map

    Verander settings.yaml naar jouw gewenste instellingen:

    # bind address, use empty string to bind to all interfaces
    bind_address: 127.0.0.1
    # backend listen port, default is 9100
    listen_port: 9100
    # Server location, laat op 0 staan om automatisch te bepalen
    server.latitude: 0
    server_longtitude: 0
    # ipinfo.io API key, if applicable
    ipinfo.apikey: ""
    
    assets.path="./assets"
    
    tls: false
    http2: false

    # tls_cert_file: "cert.pem"
    # tls_key_file: "privkey.pem"

**License**

Copyright (C) 2022 Randy Kramer

This program is free software: you can redistribute it and/or modify it under the terms of the GNU Lesser General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU Lesser General Public License along with this program. If not, see https://www.gnu.org/licenses/lgpl.

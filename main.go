package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/oschwald/geoip2-golang"
)

// STRUCT FINAL COM TODOS OS CAMPOS
type Response struct {
	IP                      string  `json:"ip"`
	City                    string  `json:"city"`
	Region                  string  `json:"region"`
	Country                 string  `json:"country"`
	CountryCode             string  `json:"country_code"`
	Latitude                float64 `json:"latitude"`
	Longitude               float64 `json:"longitude"`
	Timezone                string  `json:"timezone"`
	PostalCode              string  `json:"postal_code"`
	AsNumber                uint    `json:"as_number"`
	AsOrganization          string  `json:"as_organization"`
	ISP                     string  `json:"isp"`
	Organization            string  `json:"org"`
	Error                   string  `json:"error,omitempty"`
}

// Função auxiliar para pegar o nome em pt-BR e, se não existir, cair para 'en'
func getName(names map[string]string) string {
	if name, ok := names["pt-BR"]; ok && name != "" {
		return name
	}
	if name, ok := names["en"]; ok && name != "" {
		return name
	}
	return ""
}

func main() {
	// 1. ABRE O BANCO DE DADOS CITY
	dbCity, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal("Erro ao abrir GeoLite2-City:", err)
	}
	defer dbCity.Close()

	// 2. ABRE O BANCO DE DADOS ASN
	dbASN, err := geoip2.Open("GeoLite2-ASN.mmdb")
	if err != nil {
		log.Fatal("Erro ao abrir GeoLite2-ASN:", err)
	}
	defer dbASN.Close()

	// 3. ABRE O BANCO DE DADOS ISP
	dbISP, err := geoip2.Open("GeoLite2-ISP.mmdb")
	if err != nil {
		log.Fatal("Erro ao abrir GeoLite2-ISP:", err)
	}
	defer dbISP.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.URL.Query().Get("ip")
		if ipStr == "" {
			http.Error(w, "Por favor forneça um IP na query string: /?ip=XXX", 400)
			return
		}

		ip := net.ParseIP(ipStr)

		// CONSULTAS AOS TRÊS BANCOS
		recordCity, errCity := dbCity.City(ip)
		recordASN, _ := dbASN.ASN(ip) // ASN não falha a nível de IP inválido, mas pode falhar no arquivo.
		recordISP, _ := dbISP.ISP(ip) // ISP não falha a nível de IP inválido.

		// Verifica erro fundamental (falha total na busca City)
		if errCity != nil && errCity.Error() == "ip address not found" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{IP: ipStr, Error: "IP não encontrado na base de dados."})
			return
		}
		
		subdivisionName := ""
		if errCity == nil && len(recordCity.Subdivisions) > 0 {
			subdivisionName = getName(recordCity.Subdivisions[0].Names)
		}

		resp := Response{
			IP:          ipStr,
			City:        getName(recordCity.City.Names),
			Region:      subdivisionName,
			Country:     getName(recordCity.Country.Names),
			CountryCode: recordCity.Country.IsoCode,
			Latitude:    recordCity.Location.Latitude,
			Longitude:   recordCity.Location.Longitude,
			Timezone:    recordCity.Location.TimeZone,
			PostalCode:  recordCity.Postal.Code,
			
			// Dados ASN
			AsNumber: recordASN.AutonomousSystemNumber,
			AsOrganization: recordASN.AutonomousSystemOrganization,
			
			// Dados ISP
			ISP: recordISP.Isp,
			Organization: recordISP.Organization,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Servidor rodando na porta " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

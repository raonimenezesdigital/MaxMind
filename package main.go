package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/oschwald/geoip2-golang"
)

type Response struct {
	IP          string `json:"ip"`
	City        string `json:"city"`
	Region      string `json:"region"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

func main() {
	// Abre o banco de dados que estará na pasta do container
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Pega o IP da URL (ex: /?ip=187.45...) ou usa o IP da conexão
		ipStr := r.URL.Query().Get("ip")
		if ipStr == "" {
			http.Error(w, "Por favor forneça um IP na query string: /?ip=XXX", 400)
			return
		}

		ip := net.ParseIP(ipStr)
		record, err := db.City(ip)
		if err != nil {
			http.Error(w, "Erro ao buscar IP", 500)
			return
		}

		resp := Response{
			IP:          ipStr,
			City:        record.City.Names["pt-BR"], // Pega nome em Português
			Region:      record.Subdivisions[0].Names["pt-BR"],
			Country:     record.Country.Names["pt-BR"],
			CountryCode: record.Country.IsoCode,
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

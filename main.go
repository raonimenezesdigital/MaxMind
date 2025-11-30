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
	Error       string `json:"error,omitempty"` // Novo campo para debug
}

func main() {
	// Abre o banco de dados (Linha 20)
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.URL.Query().Get("ip")
		if ipStr == "" {
			http.Error(w, "Por favor forneça um IP na query string: /?ip=XXX", 400)
			return
		}

		ip := net.ParseIP(ipStr)
		record, err := db.City(ip)

		// Verifica se deu erro na busca (Linha 39)
		if err != nil {
			resp := Response{
				IP:    ipStr,
				Error: "Erro interno: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
			return
		}
        
        // --- NOVO TRATAMENTO PARA REGIÃO (Linhas 50-57) ---
		subdivisionName := ""
		if len(record.Subdivisions) > 0 {
			// Acessa apenas se a lista NÃO estiver vazia
			subdivisionName = record.Subdivisions[0].Names["pt-BR"]
		}
        // ---------------------------------------------------

		resp := Response{
			IP:          ipStr,
			City:        record.City.Names["pt-BR"],
			Region:      subdivisionName, // Usa a variável segura
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

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
    Error       string `json:"error,omitempty"`
}

// Função auxiliar para pegar o nome em pt-BR e, se não existir, cair para 'en'
func getName(names map[string]string) string {
    // 1. Tenta Português
    if name, ok := names["pt-BR"]; ok && name != "" {
        return name
    }
    // 2. Tenta Inglês
    if name, ok := names["en"]; ok && name != "" {
        return name
    }
    return "" // Retorna vazio se nada for encontrado
}

func main() {
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

        if err != nil {
            resp := Response{
                IP:    ipStr,
                Error: "Erro interno ao buscar IP: " + err.Error(),
            }
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError)
            json.NewEncoder(w).Encode(resp)
            return
        }
        
        // --- NOVO TRATAMENTO DE REGIÃO E CIDADE ---
        subdivisionName := ""
        if len(record.Subdivisions) > 0 {
            subdivisionName = getName(record.Subdivisions[0].Names)
        }
        // ------------------------------------------

        resp := Response{
            IP:          ipStr,
            City:        getName(record.City.Names),
            Region:      subdivisionName,
            Country:     getName(record.Country.Names),
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

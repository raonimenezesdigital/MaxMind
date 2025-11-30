# Estágio de Build
FROM golang:1.21-alpine as builder

WORKDIR /app

# Instala ferramentas necessárias
RUN apk add --no-cache curl tar

# Copia os arquivos
COPY . .

# Baixa dependências e compila
RUN go mod tidy
RUN go build -o main .

# ARGUMENTO DE BUILD (Sua chave entra aqui)
ARG MAXMIND_LICENSE_KEY

# 1. Baixa o banco de dados CITY (Cidade, Estado, Coordenadas)
RUN curl -L "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz" -o GeoLite2-City.tar.gz \
    && tar -xzf GeoLite2-City.tar.gz \
    && mv GeoLite2-City_*/*.mmdb .

# 2. Baixa o banco de dados ASN (AS Number e Organização do AS)
RUN curl -L "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz" -o GeoLite2-ASN.tar.gz \
    && tar -xzf GeoLite2-ASN.tar.gz \
    && mv GeoLite2-ASN_*/*.mmdb .
    
# 3. Baixa o banco de dados ISP (ISP e Nome da Organização)
RUN curl -L "https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ISP&license_key=${MAXMIND_LICENSE_KEY}&suffix=tar.gz" -o GeoLite2-ISP.tar.gz \
    && tar -xzf GeoLite2-ISP.tar.gz \
    && mv GeoLite2-ISP_*/*.mmdb .

# Estágio Final (Imagem leve)
FROM alpine:latest
WORKDIR /root/
# Copia o binário e os 3 arquivos MMDB
COPY --from=builder /app/main .
COPY --from=builder /app/GeoLite2-City.mmdb .
COPY --from=builder /app/GeoLite2-ASN.mmdb .
COPY --from=builder /app/GeoLite2-ISP.mmdb .

CMD ["./main"]

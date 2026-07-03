# Multi-stage build. Context build = ROOT repo (karena go.mod ada di root).
# Build:  docker build -f 17-studi-kasus-microservices/inventory-service/Dockerfile -t inventory .
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -o /out/inventory ./17-studi-kasus-microservices/inventory-service

# Image akhir minimal (statis, tanpa OS penuh).
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/inventory /inventory
EXPOSE 50052
ENTRYPOINT ["/inventory"]

# ---- Stage 1: build ----
# Context build = ROOT repo (karena go.mod di root).
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0 -> binari statis; -ldflags menyuntik versi & mengecilkan ukuran.
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/app ./30-deployment

# ---- Stage 2: runtime ----
# distroless: TANPA shell/OS penuh -> image kecil & permukaan serangan minimal.
# :nonroot -> berjalan sebagai user non-root (keamanan).
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]

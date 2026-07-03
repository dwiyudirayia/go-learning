# Multi-stage build. Context build = ROOT repo.
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -o /out/order ./17-studi-kasus-microservices/order-service

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/order /order
EXPOSE 3000
ENTRYPOINT ["/order"]

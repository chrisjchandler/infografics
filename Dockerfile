# Runtime image for the single-process Infografics server.
# Development remains documented in README.md and uses the local toolchain.
FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags='-s -w' -o /out/infograf ./cmd/infograf

FROM alpine:3.20

RUN addgroup -S infograf && adduser -S -G infograf infograf \
    && mkdir -p /data \
    && chown -R infograf:infograf /data
WORKDIR /app
COPY --from=build /out/infograf /app/infograf

ENV INFOGRAF_ADDR=:8080 \
    INFOGRAF_DATA_DIR=/data
EXPOSE 8080
VOLUME ["/data"]
USER infograf
ENTRYPOINT ["/app/infograf"]

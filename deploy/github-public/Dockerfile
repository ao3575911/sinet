# syntax=docker/dockerfile:1
FROM golang:1.27.3-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=secret,id=proxy_ca \
    if [ -f /run/secrets/proxy_ca ]; then SSL_CERT_FILE=/run/secrets/proxy_ca go mod download; else go mod download; fi
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /connector-github ./cmd/connector-github
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /connector-github /connector-github
USER 65532:65532
EXPOSE 8090
ENTRYPOINT ["/connector-github"]

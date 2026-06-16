FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o /app/yt-transcript-server ./cmd/server

FROM scratch
COPY --from=builder /app/yt-transcript-server /usr/local/bin/yt-transcript-server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV PORT=8080
ENV SERVER_HOST=0.0.0.0
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/yt-transcript-server"]

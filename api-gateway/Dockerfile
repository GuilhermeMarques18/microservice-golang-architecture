FROM golang:1.26.3-alpine AS builder
LABEL authors="guilherme"

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /gateway \
    ./api-gateway

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /gateway /gateway

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/gateway"]

# syntax=docker/dockerfile:1
FROM golang:1.20-alpine AS build

WORKDIR /src
# Copy go.mod + go.sum first for caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /src/cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server .

# Final minimal image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /out/server /usr/local/bin/internal-transfers
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/internal-transfers"]

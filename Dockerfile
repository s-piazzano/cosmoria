FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cosmoria ./cmd/cosmoria/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /cosmoria /cosmoria
COPY --from=builder /src/db/migrations /db/migrations
EXPOSE 8080
CMD ["/cosmoria"]

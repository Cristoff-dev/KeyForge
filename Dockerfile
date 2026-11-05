FROM golang:alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o keyforge ./cmd/keyforge

FROM gcr.io/distroless/static-debian12

USER nonroot:nonroot

COPY --from=builder /app/keyforge /keyforge

EXPOSE 8080

ENTRYPOINT ["/keyforge"]
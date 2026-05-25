FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy root module files and pre-download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy full source tree
COPY . .

# Build the faucet server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/faucet-server ./faucet-app/server

FROM alpine:3.23.3

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1001 -S faucet
RUN adduser -S faucet -u 1001 -G faucet

USER faucet

COPY --from=builder /build/bin/faucet-server /bin/faucet-server

EXPOSE 8080

CMD ["faucet-server"]

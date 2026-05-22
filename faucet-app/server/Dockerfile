FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy root module files first (the replace directive in faucet-app/server points here)
COPY go.mod go.sum ./

# Copy faucet-app/server module files and pre-download dependencies
COPY faucet-app/server/go.mod faucet-app/server/go.sum ./faucet-app/server/
RUN cd faucet-app/server && go mod download

# Copy full source tree so the replace directive can resolve
COPY . .

# Build the faucet server binary
RUN cd faucet-app/server && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /build/bin/faucet-server .

FROM alpine:3.23.3

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1001 -S faucet
RUN adduser -S faucet -u 1001 -G faucet

USER faucet

COPY --from=builder /build/bin/faucet-server /bin/faucet-server

EXPOSE 8080

CMD ["faucet-server"]

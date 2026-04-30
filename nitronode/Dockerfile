FROM golang:1.25-alpine AS builder

WORKDIR /build
ENV CGO_ENABLED=1
ARG VERSION="n/a"

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./bin/nitronode -ldflags "-X main.Version=${VERSION}" ./nitronode

FROM alpine:3.23.3

RUN addgroup -g 1001 -S nitronode
RUN adduser -S nitronode -u 1001 -G nitronode

USER nitronode

COPY --from=builder /build/bin /bin

EXPOSE 7824 4242

CMD ["nitronode"]

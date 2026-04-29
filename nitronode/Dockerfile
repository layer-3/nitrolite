FROM golang:1.25-alpine AS builder

WORKDIR /build
ENV CGO_ENABLED=1
ARG VERSION="n/a"

RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./bin/clearnode -ldflags "-X main.Version=${VERSION}" ./clearnode

FROM alpine:3.23.3

RUN addgroup -g 1001 -S clearnode
RUN adduser -S clearnode -u 1001 -G clearnode

USER clearnode

COPY --from=builder /build/bin /bin

EXPOSE 7824 4242

CMD ["clearnode"]

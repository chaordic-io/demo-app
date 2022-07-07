FROM golang:1.18.3 AS builder
ENV CGO_ENABLED=0
WORKDIR /build
COPY . .
# RUN go build -ldflags "-X 'github.com/chaordic-io/demo-app/internal.Version=$VERSION' -X 'github.com/chaordic-io/demo-app/internal.BuildDate=$(date)'" -o demo-app cmd/main.go
RUN go build -o demo-app cmd/main.go

FROM alpine:3.16.0

RUN apk update && apk add git tzdata
RUN adduser -S demo

WORKDIR /app

COPY --from=builder /build/demo-app /app/demo-app

USER demo

CMD ["./demo-app"]
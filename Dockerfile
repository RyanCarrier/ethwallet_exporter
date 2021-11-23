FROM golang:1.17.3-alpine as build

RUN apk add build-base

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o /ethwallet_exporter

FROM alpine:3.15

COPY --from=build /ethwallet_exporter /

EXPOSE 9887

ENTRYPOINT [ "/ethwallet_exporter" ]
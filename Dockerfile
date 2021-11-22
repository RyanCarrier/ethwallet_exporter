FROM golang:1.17.3-alpine3.14

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /ethwallet_exporter

EXPOSE 8084

CMD [ "/ethwallet_exporter" ]
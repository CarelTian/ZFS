FROM golang:1.23.5-alpine

WORKDIR /app
COPY .. .

RUN go mod download
RUN go build

EXPOSE 9000

CMD ["./ZFS"]
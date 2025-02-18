# Instructions for Golang docker image: https://hub.docker.com/_/golang
FROM golang:1.24

WORKDIR /usr/src/app/

COPY go.mod go.sum /usr/src/app/
RUN go mod download

COPY . .
RUN go build -o ./ -v ./...

ENTRYPOINT [ "./fragments" ]
CMD [ "./fragments", "debug"]
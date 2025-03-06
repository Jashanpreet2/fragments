# Instructions for Golang docker image: https://hub.docker.com/_/golang
# Image optimization guidance found at: 
# https://medium.com/code-beyond/dockerizing-golang-apps-a-step-by-step-guide-to-reducing-docker-image-size-306898e7359e

FROM golang:1.24.0-alpine3.21@sha256:2d40d4fc278dad38be0777d5e2a88a2c6dee51b0b29c97a764fc6c6a11ca893c AS buildstage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o ./ -v ./...

ENTRYPOINT [ "./fragments" ]
CMD [ "./fragments", "debug"]

FROM alpine:latest

WORKDIR /app

COPY --from=buildstage /app/fragments /app/
COPY --from=buildstage /app/.env.debug /app/
COPY --from=buildstage /app/testProfiles.csv /app/

CMD ["./fragments", "debug"]
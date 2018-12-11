FROM golang:alpine as build
RUN apk add --no-cache git
WORKDIR /src
ADD main.go .
RUN go get -d -v ./...
RUN go build -v -o switchd

FROM alpine
COPY --from=build /src/switchd /usr/local/bin/switchd
EXPOSE 8000
CMD switchd

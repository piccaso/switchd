FROM golang:alpine as build
RUN apk add --no-cache git
WORKDIR /src
ADD switchd.go .
RUN go get -d -v ./...
RUN go build -v -o switchd switchd.go

FROM alpine
COPY --from=build /src/switchd /usr/local/bin/switchd
EXPOSE 8000
CMD switchd

FROM alpine:edge AS build
RUN apk update
RUN apk upgrade
RUN apk add --update go gcc g++ musl-dev
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
RUN GOOS=linux go build -o /go-app -tags musl

FROM golang:1.18-alpine

COPY --from=build /go-app /go-app
CMD [ "/go-app" ]


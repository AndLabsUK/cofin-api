FROM golang:1.20-alpine

WORKDIR /app

COPY . .

COPY ./crontab /etc/crontabs/root

RUN apk add make

RUN make all

CMD crond && ./bin/api


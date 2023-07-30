FROM golang:1.20-alpine

WORKDIR /app

COPY ./bin ./bin

COPY ./crontab /etc/crontabs/root

CMD crond && ./bin/api-amd64-linux

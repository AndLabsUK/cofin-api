FROM alpine:latest

RUN apk add build-base gcompat

WORKDIR /app

COPY ./bin ./bin

COPY ./crontab /etc/crontabs/root

EXPOSE 80

CMD crond && /app/bin/api-amd64-linux

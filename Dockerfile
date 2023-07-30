FROM alpine:latest

WORKDIR /app

COPY ./bin ./bin

COPY ./crontab /etc/crontabs/root

EXPOSE 80

CMD crond && /app/bin/api-amd64-linux

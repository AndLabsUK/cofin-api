FROM golang:1.20.5-alpine

WORKDIR /app

COPY ./cofin-api .

EXPOSE 8080

CMD [ "./cofin-api" ]
To start, make sure you have a `.env` file following the example in `.env.example`. Then, you can start the API service locally:
`APP_NAME=api make run`

# API

## Conversational API 
Generate a response.

```
curl --location 'http://localhost:8080/conversation' \
--header 'Authorization: 783bc63759a37295812c9b486da39f9a0bfa53d41d5bd3ee248739f931ba9234' \
--header 'Content-Type: application/json' \
--data '{
    "author": "user",
    "text": "Did the company revenue grow or contract?",
    "ticker": "AAC"
}'
```
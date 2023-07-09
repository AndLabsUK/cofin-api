`make all`
`./bin/api`

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
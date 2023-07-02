To start, make sure you have a `.env` file following the example in `.env.example`. Then, you can start the API service locally:
`APP_NAME=api make run`

# API

## Conversational API 
Generate a response.

```
curl --location --request GET 'http://localhost:8080/conversation' \
--header 'Content-Type: application/json' \
--data '{
    "exchanges": [],
    "user_message": {
        "text": "How much did the company earn?",
        "ticker": "$NET",
        "year": 2023,
        "quarter": 1,
        "source_type": "10-Q"
    }
}'
```
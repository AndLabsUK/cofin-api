package real_stonks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type RealStonks struct{}

type TickerInformation struct {
	Currency    string
	Price       float64
	Change      float64
	TotalVolume float64
}

func (rs RealStonks) GetMarketData(ticker string) (*TickerInformation, error) {
	type dto struct {
		Price            float64 `json:"price"`
		ChangePercentage float64 `json:"change_percentage"`
		TotalVolume      string  `json:"total_vol"`
	}

	url := "https://realstonks.p.rapidapi.com/" + ticker
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", os.Getenv("RAPID_API_KEY"))
	req.Header.Add("X-RapidAPI-Host", "realstonks.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var d *dto
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", string(body), err)
	}

	ti := &TickerInformation{
		Currency:    "USD",
		Price:       d.Price,
		Change:      d.ChangePercentage,
		TotalVolume: convertToFloat(d.TotalVolume),
	}

	return ti, err
}

func convertToFloat(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	suffix := s[len(s)-1:]
	switch suffix {
	case "K", "M", "B":
		modifiedString := s[:len(s)-1]

		value, err := strconv.ParseFloat(modifiedString, 64)
		if err != nil {
			panic(err)
		}

		switch suffix {
		case "K":
			value *= 1000
		case "M":
			value *= 1000000
		case "B":
			value *= 1000000000
		}

		return value
	default:
		value, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(err)
		}
		return value
	}
}

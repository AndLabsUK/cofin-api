package integrations

import (
	"encoding/json"
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
		TotalVolume      string  `json:"total_volume"`
	}

	url := "https://realstonks.p.rapidapi.com/" + ticker
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", os.Getenv("RAPID_API_KEY"))
	req.Header.Add("X-RapidAPI-Host", "realstonks.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var d *dto
	err := json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	ti := &TickerInformation{
		Currency:    "USD", //TODO: Upgrade when non-US stocks are supported
		Price:       d.Price,
		Change:      d.ChangePercentage,
		TotalVolume: convertToFloat(d.TotalVolume),
	}

	return ti, err
}

func convertToFloat(s string) float64 {
	suffix := s[len(s)-1:]
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
}

package sec_api

import (
	"bytes"
	"cofin/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type Exchange string

const (
	NYSE   Exchange = "nyse"
	NASDAQ Exchange = "nasdaq"
)

// The list of exchanges we fetchDocuments companies from. London Stock Exchange
// WHEN?!
var StockExchanges = []Exchange{NYSE, NASDAQ}

// This is the response from the SEC API when we request a list of companies
// traded on an exchange.
type Listing struct {
	Name         string `json:"name,omitempty"`
	Ticker       string `json:"ticker,omitempty"`
	CIK          string `json:"cik,omitempty"`
	CUSIP        string `json:"cusip,omitempty"`
	Exchange     string `json:"exchange,omitempty"`
	IsDelisted   bool   `json:"isDelisted,omitempty"`
	Category     string `json:"category,omitempty"`
	Sector       string `json:"sector,omitempty"`
	Industry     string `json:"industry,omitempty"`
	SIC          string `json:"sic,omitempty"`
	SICSector    string `json:"sicSector,omitempty"`
	SICIndustry  string `json:"sicIndustry,omitempty"`
	FAMASector   string `json:"famaSector,omitempty"`
	FAMAIndustry string `json:"famaIndustry,omitempty"`
	Currency     string `json:"currency,omitempty"`
	Location     string `json:"location,omitempty"`
	ID           string `json:"id,omitempty"`
}

// This is the response from the SEC API when we request a list of filings for a
// company.
type Filing struct {
	ID                                   string               `json:"id,omitempty"`
	AccessionNo                          string               `json:"accessionNo,omitempty"`
	CIK                                  string               `json:"cik,omitempty"`
	Ticker                               string               `json:"ticker,omitempty"`
	CompanyName                          string               `json:"companyName,omitempty"`
	CompanyNameLong                      string               `json:"companyNameLong,omitempty"`
	FormType                             string               `json:"formType,omitempty"`
	Description                          string               `json:"description,omitempty"`
	FiledAt                              string               `json:"filedAt,omitempty"`
	LinkToTxt                            string               `json:"linkToTxt,omitempty"`
	LinkToHtml                           string               `json:"linkToHtml,omitempty"`
	LinkToXbrl                           string               `json:"linkToXbrl,omitempty"`
	LinkToFilingDetails                  string               `json:"linkToFilingDetails,omitempty"`
	Entities                             []Entity             `json:"entities,omitempty"`
	DocumentFormatFiles                  []DocumentFormatFile `json:"documentFormatFiles,omitempty"`
	DataFiles                            []DataFile           `json:"dataFiles,omitempty"`
	SeriesAndClassesContractsInformation []interface{}        `json:"seriesAndClassesContractsInformation,omitempty"`
	PeriodOfReport                       string               `json:"periodOfReport,omitempty"`
}

// This object is embedded in the Filing object.
type DocumentFormatFile struct {
	Sequence    string `json:"sequence,omitempty"`
	Description string `json:"description,omitempty"`
	DocumentURL string `json:"documentUrl,omitempty"`
	Type        string `json:"type,omitempty"`
	Size        string `json:"size,omitempty"`
}

// This object is embedded in the Filing object.
type DataFile struct {
	Sequence    string `json:"sequence,omitempty"`
	Description string `json:"description,omitempty"`
	DocumentURL string `json:"documentUrl,omitempty"`
	Type        string `json:"type,omitempty"`
	Size        string `json:"size,omitempty"`
}

// This object is embedded in the Filing object.
type Entity struct {
	CompanyName          string `json:"companyName,omitempty"`
	CIK                  string `json:"cik,omitempty"`
	IRSNo                string `json:"irsNo,omitempty"`
	StateOfIncorporation string `json:"stateOfIncorporation,omitempty"`
	FiscalYearEnd        string `json:"fiscalYearEnd,omitempty"`
	Type                 string `json:"type,omitempty"`
	Act                  string `json:"act,omitempty"`
	FileNo               string `json:"fileNo,omitempty"`
	FilmNo               string `json:"filmNo,omitempty"`
	Sic                  string `json:"sic,omitempty"`
}

// Get companies traded on an exchange.
func GetTradedCompanies(key string, exchange Exchange) (listings []Listing, err error) {
	const exchangeURLTemplate = "https://api.sec-api.io/mapping/exchange/%v"
	req, err := http.NewRequest("GET", fmt.Sprintf(exchangeURLTemplate, exchange), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", key)

	client := retryablehttp.NewClient()
	client.Logger = nil
	resp, err := client.StandardClient().Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &listings)
	if err != nil {
		return nil, err
	}

	return listings, nil
}

func GetFilingOriginURL(filing Filing) string {
	// Template for paths to the original files on the SEC website.
	const secFileURLTemplate = "https://www.sec.gov/Archives/edgar/data/%v/%v/%v"
	// Get the file name.
	_, fileName := path.Split(filing.LinkToFilingDetails)
	// In the URL, the accession number should have no dashes.
	accessionNumber := strings.ReplaceAll(filing.AccessionNo, "-", "")
	originURL := fmt.Sprintf(secFileURLTemplate, filing.CIK, accessionNumber, fileName)
	return originURL
}

func GetSECArchiveURL(filing Filing) string {
	// Template for downloadable files in the paid SEC API archive.
	const secArchiveURLTemplate = "https://archive.sec-api.io/%v/%v/%v"
	// Get the file name.
	_, fileName := path.Split(filing.LinkToFilingDetails)
	// In the URL, the accession number should have no dashes.
	accessionNumber := strings.ReplaceAll(filing.AccessionNo, "-", "")
	url := fmt.Sprintf(secArchiveURLTemplate, filing.CIK, accessionNumber, fileName)
	return url
}

// Get the filing file from the SEC. Return the origin URL on the SEC website,
// the file bytes, and an error, if there is one.
func GetFilingFile(key string, filing Filing) (file []byte, err error) {
	req, err := http.NewRequest("GET", GetSECArchiveURL(filing), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", key)

	client := retryablehttp.NewClient()
	client.Logger = nil
	resp, err := client.StandardClient().Do(req)
	if err != nil {
		return nil, err
	}

	f, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func GetFilingsSince(key, cik string, kind models.SourceKind, since time.Time, limit int) (filings []Filing, err error) {
	timeStart := since.Format(time.RFC3339)
	timeEnd := time.Now().Format(time.RFC3339)

	var jsonStr = []byte(
		fmt.Sprintf(`{
			"query": {
				"query_string": {
					"query": "formType:\"%v\" AND filedAt:[%v TO %v] AND cik:(%v)",
					"time_zone": "America/New_York"
				}
			},
			"from": "0",
			"size": "%v",
			"sort": [{ "filedAt": { "order": "asc" } }]
		}`, kind, timeStart, timeEnd, cik, limit),
	)

	req, err := http.NewRequest("POST", "https://api.sec-api.io", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", key)
	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Logger = nil
	resp, err := client.StandardClient().Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type Total struct {
		Value    int    `json:"value"`
		Relation string `json:"relation"`
	}

	type Query struct {
		From int `json:"from"`
		Size int `json:"size"`
	}

	type response struct {
		Total   Total    `json:"total"`
		Query   Query    `json:"query"`
		Filings []Filing `json:"filings"`
	}

	var r response
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}

	return r.Filings, nil
}

func ExtractSectionContent(key, originURL string, section models.Section) (string, error) {
	const URLTemplate = "https://api.sec-api.io/extractor?token=%v&url=%v&item=%v&type=text"
	req, err := http.NewRequest("GET", fmt.Sprintf(URLTemplate, key, originURL, section), nil)
	if err != nil {
		return "", err
	}

	client := retryablehttp.NewClient()
	client.Logger = nil
	resp, err := client.StandardClient().Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

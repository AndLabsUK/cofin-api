package main

import (
	"bytes"
	"cofin/core"
	"cofin/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
)

type Match struct {
	ID string `json:"id"`
}

type Matches []Match

type queryResponse struct {
	Matches Matches `json:"matches"`
}

func main() {
	godotenv.Load()

	db, err := core.InitDB()
	if err != nil {
		panic(err)
	}

	err = db.Debug().AutoMigrate(
		&models.User{},
		&models.Company{},
		&models.Document{},
		&models.AccessToken{},
		&models.Message{},
	)
	if err != nil {
		panic(err)
	}

	companies, err := models.GetCompanies(db)
	if err != nil {
		panic(err)
	}

	for i, company := range companies {
		log.Printf("Migrating company %v of %v", i+1, len(companies))

		matches := getMatches(company.ID)
		for _, match := range matches {
			log.Println(match.ID)
			getVector(company.ID, match.ID)
			return
		}
		return
	}
}

func getVector(companyID uint, id string) {
	r, err := http.NewRequest("GET", fmt.Sprintf("https://staging-0344b61.svc.us-west1-gcp-free.pinecone.io/vectors/fetch?ids=%v&namespace=%v", id, companyID), nil)
	if err != nil {
		panic(err)
	}

	r.Header.Add("Api-Key", "cbe242a6-c258-4324-be54-8cefc6e6a1e4")
	r.Header.Add("accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	log.Println(res.StatusCode)

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(b))
}

func getMatches(companyID uint) Matches {
	nullVector := make([]string, 1536)
	for i := range nullVector {
		nullVector[i] = "0"
	}

	v := "[" + strings.Join(nullVector, ",") + "]"

	body := fmt.Sprintf("{\"topK\":10000, \"namespace\": \"%v\", \"vector\": %v}", companyID, v)
	r, err := http.NewRequest("POST", "https://staging-0344b61.svc.us-west1-gcp-free.pinecone.io/query", bytes.NewBuffer([]byte(body)))
	if err != nil {
		panic(err)
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Api-Key", "cbe242a6-c258-4324-be54-8cefc6e6a1e4")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	response := queryResponse{}
	json.Unmarshal(b, &response)
	res.Body.Close()
	return response.Matches
}

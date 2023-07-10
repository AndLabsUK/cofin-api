package google_pki

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type GooglePKI struct{}

func (gpki GooglePKI) GetPublicKeys() ([]GooglePublicKey, error) {
	type dtoPayload struct {
		Keys []GooglePublicKey `json:"keys"`
	}

	url := "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com"
	req, _ := http.NewRequest("GET", url, nil)

	res, _ := http.DefaultClient.Do(req)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var d *dtoPayload
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	return d.Keys, nil
}

func (gpki GooglePKI) GetPublicKeyForKid(kid string) (*GooglePublicKey, error) {
	keys, err := gpki.GetPublicKeys()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.Kid == kid {
			return &key, nil
		}
	}
	return nil, errors.New("key not found")
}

type GooglePublicKey struct {
	Use string `json:"use"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

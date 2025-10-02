package geonames

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type ApiGeonames struct {
	Username string
	BaseURL  string
	Client   *http.Client
}

func NewApiGeonames(username string, client *http.Client) *ApiGeonames {
	return &ApiGeonames{
		Username: username,
		BaseURL:  "http://api.geonames.org/timezoneJSON",
		Client:   client,
	}
}

func (api *ApiGeonames) Lookup(ctx context.Context, lat, lon float64) (string, error) {
	if api.Username == "" {
		log.Default().Printf("[ApiGeonames.Lookup] geonames.Username is empty")
		return "", errors.New("geonames username is empty")
	}

	q := url.Values{}
	q.Set("lat", fmt.Sprintf("%f", lat))
	q.Set("lng", fmt.Sprintf("%f", lon))
	q.Set("username", api.Username)

	u := api.BaseURL + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		log.Println("[ApiGeonames.Lookup] error creating request", err)
		return "", err
	}
	resp, err := api.Client.Do(req)
	if err != nil {
		log.Println("[ApiGeonames.Lookup] error while executing", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	var data response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println("[ApiGeonames.Lookup] error decoding response:", err)
		return "", err
	}

	if data.Status != nil && data.Status.Value != 0 {
		log.Println("[ApiGeonames.Lookup] failure request:", data.Status)
		return "", fmt.Errorf("geonames error (%d): %s", data.Status.Value, data.Status.Message)
	}
	if data.TimeZoneID == "" {
		log.Println("[ApiGeonames.Lookup] failure request: no timezone id found")
		return "", errors.New("empty timezoneId in response")
	}
	return data.TimeZoneID, nil
}

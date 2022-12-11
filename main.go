package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	BASE_URL          = "https://api.princeton.edu:443/student-app/1.0.1"
	REFRESH_TOKEN_URL = "https://api.princeton.edu:443/token"
)

type CampusAPIHelper struct {
	baseUrl     string
	refreshUrl  string
	keyName     string
	secretName  string
	accessToken string
	cache       map[string]*http.Response
	// Route   	func() http.HandlerFunc
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// func NewCampusAPIHelper(consumerKey string, consumerSecret string, baseUrl string, refreshUrl string) *CampusAPIHelper {
// 	helper := &CampusAPIHelper{
// 		baseUrl:    baseUrl,
// 		refreshUrl: refreshUrl,
// 		keyName:    consumerKey,
// 		secretName: consumerSecret,
// 	}
// 	err := helper.refreshAccess()
// 	if err != nil {
// 		// DO ERROR HANDLING
// 	}

// 	return helper
// }

func (s *CampusAPIHelper) refreshAccess() error {
	consumerKey := os.Getenv(s.keyName)
	consumerSecret := os.Getenv(s.secretName)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", s.refreshUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(consumerKey+":"+consumerSecret)))

	response, err := client.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var refreshResponse RefreshTokenResponse
	err = json.Unmarshal(b, &refreshResponse)

	if err != nil {
		return err
	}

	fmt.Println("got new access token: " + refreshResponse.AccessToken)
	s.accessToken = refreshResponse.AccessToken
	return nil
}

func (s *CampusAPIHelper) Request(*http.Request) (*http.Response, error) {
	return nil, nil
}

func main() {
	fmt.Println("HELLO WORLD")
}

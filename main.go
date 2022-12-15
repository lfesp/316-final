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
	"sync"
	"time"

	"github.com/joho/godotenv"
)

const (
	BASE_URL          = "https://api.princeton.edu:443/active-directory/1.0.5"
	REFRESH_TOKEN_URL = "https://api.princeton.edu:443/token"
)

type CampusAPIHelper struct {
	refreshUrl     string
	consumerKey    string
	consumerSecret string
	accessToken    string
	lock           *sync.RWMutex
	// cache       map[string]*http.Response
	// Route   	func() http.HandlerFunc
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type Student struct {
	UniversityId string `json:?universityid"`
	UID          string `json:?uid"`
	Name         string `json:?displayname"`
	Email        string `json:?mail"`
}

func NewCampusAPIHelper(consumerKey string, consumerSecret string, refreshUrl string) *CampusAPIHelper {
	helper := &CampusAPIHelper{
		refreshUrl:     refreshUrl,
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		lock:           &sync.RWMutex{},
	}
	err := helper.refreshAccess(0)
	if err != nil {
		// DO ERROR HANDLING
		// dont yell at me
	}

	return helper
}

func (s *CampusAPIHelper) refreshAccess(i int) error {
	gotLock := s.lock.TryLock()
	if !gotLock {
		fmt.Printf("DID NOT GET LOCK %v \n", i)
		s.lock.Lock()
		s.lock.Unlock()
		fmt.Printf("RETURNED TO ORIGINAL REQUEST WITH NEW TOKEN %v \n", i)
		return nil
	}

	fmt.Println("REFRESHING STARTED")

	defer s.lock.Unlock()

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", s.refreshUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.consumerKey+":"+s.consumerSecret)))

	response, err := client.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	refreshResponse := &RefreshTokenResponse{}
	err = json.Unmarshal(b, refreshResponse)

	if err != nil {
		return err
	}

	// fmt.Println("got new access token: " + refreshResponse.AccessToken)
	s.accessToken = refreshResponse.AccessToken

	fmt.Println("REFRESHING FINISHED")

	return nil
}

func (s *CampusAPIHelper) Do(req *http.Request) (*http.Response, error) {

	// IF THE HTTP REQUEST FAILS
	// 1. try to get the lock (with TryLock)
	// 2a. you GOT the lock, so refresh the token and the unlock
	// 2b. you DID NOT get the lock. wait for the lock to be released and, once it is, immediatly grab and release it
	// 3. make the initial request again, now with a fresh access token
	c := &http.Client{
		Timeout: time.Second * 10, // my default timeout
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	res, err := c.Do(req)
	if res.StatusCode == http.StatusUnauthorized {
		fmt.Println("Authorization is stale")
		err = s.refreshAccess(0)
		if err != nil {
			fmt.Println("Error 1")
			// http.Error(w, "Unable to refresh access token.", http.StatusBadRequest)
			return res, err
		}

		fmt.Println("GOT TO THIS PART OF REFRESH")

		req.Header.Set("Authorization", "Bearer "+s.accessToken)
		res, err = c.Do(req)
		if err != nil {
			fmt.Println("Error 2")
			// http.Error(w, "Unable to retrieve menu data.", http.StatusBadRequest)
			return res, err
		}
	}
	if err != nil {
		fmt.Println("Errored when sending request to the server")
		return res, err
	}

	return res, err
	// remember must close body !! should we make the client do this?
}

func main() {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper := NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL)

	for i := 0; i < 15; i++ {
		go func(i int) {
			testHelper.refreshAccess(i)
		}(i)
	}

	req, err := http.NewRequest(http.MethodGet, BASE_URL+"/users/basic?uid=liame", nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
	}

	res, err := testHelper.Do(req)
	if err != nil {
		fmt.Printf("client: could not execute request: %s\n", err)
	}

	dec := json.NewDecoder(res.Body)
	var s []Student
	err = dec.Decode(&s)
	if err != nil {
		fmt.Printf("client: could not decode json response body: %s\n", err)
	}

	fmt.Printf("%#v \n", s)

	res.Body.Close()

	time.Sleep(10 * time.Second)

	req, err = http.NewRequest(http.MethodGet, BASE_URL+"/users/basic?uid=hvera", nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
	}

	res, err = testHelper.Do(req)
	if err != nil {
		fmt.Printf("client: could not execute request: %s\n", err)
	}

	dec = json.NewDecoder(res.Body)
	err = dec.Decode(&s)
	if err != nil {
		fmt.Printf("client: could not decode json response body: %s\n", err)
	}

	fmt.Printf("%#v \n", s)
}

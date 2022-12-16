package apihelper

import (
	"bufio"
	"bytes"
	"campus-api-helper/cache"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

// A CampusAPIHelper is a wrapper for an HTTP client interacting with Princeton's REST APIs
type CampusAPIHelper struct {
	refreshUrl     string // URL to refresh the access token
	consumerKey    string
	consumerSecret string
	accessToken    string
	lock           *sync.RWMutex
	client         *http.Client
	cache          *cache.LRU
}

type refreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func NewCampusAPIHelper(consumerKey string, consumerSecret string, refreshUrl string, client *http.Client, cacheSize int) (*CampusAPIHelper, error) {
	helper := &CampusAPIHelper{
		refreshUrl:     refreshUrl,
		consumerKey:    consumerKey,
		consumerSecret: consumerSecret,
		lock:           &sync.RWMutex{},
		client:         client,
		cache:          cache.NewLru(cacheSize),
	}

	if helper.client == nil {
		helper.client = http.DefaultClient
	}

	err := helper.refreshAccess(0)
	if err != nil {
		return nil, fmt.Errorf("error obtaining access token: %v", err)
	}

	return helper, nil
}

// Refreshes the access token
func (s *CampusAPIHelper) refreshAccess(i int) error {
	gotLock := s.lock.TryLock()
	if !gotLock {
		// fmt.Printf("DID NOT GET LOCK %v \n", i)
		s.lock.Lock()
		s.lock.Unlock()
		// fmt.Printf("RETURNED TO ORIGINAL REQUEST WITH NEW TOKEN %v \n", i)
		return nil
	}

	// fmt.Println("REFRESHING STARTED")

	defer s.lock.Unlock()

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", s.refreshUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(s.consumerKey+":"+s.consumerSecret)))

	response, err := s.client.Do(req)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	refreshResponse := &refreshTokenResponse{}
	err = json.Unmarshal(b, refreshResponse)

	if err != nil {
		return err
	}

	s.accessToken = refreshResponse.AccessToken

	// fmt.Println("REFRESHING FINISHED")

	return nil
}

// Sends an HTTP request and returns an HTTP response
func (s *CampusAPIHelper) Do(req *http.Request) (*http.Response, error) {
	// If the HTTP Request fails:
	//	 1. try to get the lock (with TryLock)
	// 	 2a. you GOT the lock. refresh the token and then unlock
	// 	 2b. you DID NOT get the lock. wait for the lock to be released and, once it is, immediatly grab and release it
	// 	 3. make the initial request again, now with a fresh access token

	s.lock.RLock()
	defer s.lock.RUnlock()

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	res, err := s.client.Do(req)
	if res.StatusCode == http.StatusUnauthorized {
		err = s.refreshAccess(0)
		if err != nil {
			return res, fmt.Errorf("error refreshing access token %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+s.accessToken)
		res, err = s.client.Do(req)
	}

	return res, err
}

// Issues a GET to the specified URL
func (s *CampusAPIHelper) Get(url string) (*http.Response, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	value, found := s.cache.Get(url)

	if found {
		fmt.Printf("cache hit!")
		reader := bufio.NewReader(bytes.NewReader(value))
		resp, err := http.ReadResponse(reader, nil)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}

	// Make new HTTP request if a cache miss
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.accessToken)

	res, err := s.client.Do(req)
	if res.StatusCode == http.StatusUnauthorized {
		err = s.refreshAccess(0)
		if err != nil {
			return res, fmt.Errorf("error refreshing access token %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+s.accessToken)

		res, err = s.client.Do(req)
	}

	body, err := httputil.DumpResponse(res, true)

	if err != nil {
		return nil, err
	}

	s.cache.Set(url, body)
	// fmt.Println(s.cache.RemainingStorage())

	return res, err
}

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

// CampusAPIHelper is a wrapper for an HTTP client
// interacting with Princeton's REST APIs that
// abstracts away the management of API access tokens.
type CampusAPIHelper struct {
	refreshUrl     string // URL to refresh the access token
	consumerKey    string
	consumerSecret string
	accessToken    string
	lock           *sync.RWMutex
	client         *http.Client
	cache          *cache.LRU
}

// helper struct for unmarshalling access token regeneration responses
type refreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// factory method that instantiates and returns a new CampusAPIHelper struct
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

// refreshes the access token
func (s *CampusAPIHelper) refreshAccess(i int) error {
	// CONCURRENCY LOGIC:
	// If the HTTP Request fails:
	//	 1. try to get the lock (with TryLock)
	// 	 2. case a - you GOT the lock. refresh the token and then unlock
	// 		case b - you DID NOT get the lock. wait for the lock to be released (with Lock)
	//				 and, once it is, immediately release it
	// 	 3. make the initial request again, now with a fresh access token

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

// execute an HTTP request with API access token authentication
func (s *CampusAPIHelper) Do(req *http.Request) (*http.Response, error) {
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

// issues a GET to the specified URL and caches the result.
// if the url results in a cache hit, no HTTP request is issued and the
// cached response body is returned in a new response.
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

	// make new HTTP request if a cache miss
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

	return res, err
}

// issues a HEAD to the specified URL
// adapted from go http package source code:
// https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/net/http/client.go;l=919
func (s *CampusAPIHelper) Head(url string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return s.Do(req)
}

// issues a POST to the specified URL
// adapted from go http package source code:
// https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/net/http/client.go;l=919
func (s *CampusAPIHelper) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return s.Do(req)
}

// issues a POST to the specified URL, with data's keys and
// values URL-encoded as the request body.
// adapted from go http package source code:
// https://cs.opensource.google/go/go/+/refs/tags/go1.19.4:src/net/http/client.go;l=919
func (s *CampusAPIHelper) PostForm(url string, data url.Values) (*http.Response, error) {
	return s.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

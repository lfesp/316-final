package apihelper

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

/******************************************************************************/
/*                                Constants                                   */
/******************************************************************************/

const (
	BASE_URL          = "https://api.princeton.edu:443/active-directory/1.0.5"
	REFRESH_TOKEN_URL = "https://api.princeton.edu:443/token"
)

type Student struct {
	UniversityId string `json:?universityid"`
	UID          string `json:?uid"`
	Name         string `json:?displayname"`
	Email        string `json:?mail"`
}

var netids []string = []string{"liame", "hvera", "sc73", "mtouil", "shmeyer", "cjcheng", "adogra", "cabrooks", "juliacw", "aalevy", "nk5635"}

/******************************************************************************/
/*                                 Helpers                                    */
/******************************************************************************/

// getStudentDo tests CampusAPIHelper's Do() method
func getStudentDo(apiHelper *CampusAPIHelper, t *testing.T) {

	netid := netids[rand.Intn(len(netids))]
	req, err := http.NewRequest(http.MethodGet, BASE_URL+"/users/basic?uid="+netid, nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
	}

	res, err := apiHelper.Do(req)
	if err != nil {
		fmt.Printf("client: could not execute request: %s\n", err)
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)
	var s []Student
	err = dec.Decode(&s)
	if err != nil {
		fmt.Printf("client: could not decode json response body: %s\n", err)
		t.Errorf("Test failed: Router gave non-200 status code: %d", res.StatusCode)
	}

	fmt.Printf("%#v \n", s)
}

// getStudentGet tests CampusAPIHelper's Get() method, which utilizes an LRU cache
func getStudentGet(apiHelper *CampusAPIHelper, t *testing.T) {

	netid := netids[rand.Intn(len(netids))]

	res, err := apiHelper.Get(BASE_URL + "/users/basic?uid=" + netid)
	if err != nil {
		fmt.Printf("client: could not execute request: %s\n", err)
	}
	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)
	var s []Student
	err = dec.Decode(&s)
	if err != nil {
		fmt.Printf("client: could not decode json response body: %s\n", err)
	}

	fmt.Printf("%#v \n", s)

}

/******************************************************************************/
/*                                  Tests                                     */
/******************************************************************************/

//
func TestGet(t *testing.T) {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil, 100000)
	if err != nil {
		log.Fatalln(err)
	}
	// retrieves information for student netIDs
	// intermediate 5 second lag enables observation of cache functionality
	for i := 0; i < 15; i++ {
		go func(helper *CampusAPIHelper, t *testing.T) {
			getStudentDo(helper, t)
		}(testHelper, t)
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("LRU CACHE STATS: %v \n", testHelper.cache.Stats())

	for i := 0; i < 15; i++ {
		go func(helper *CampusAPIHelper, t *testing.T) {
			getStudentDo(helper, t)
		}(testHelper, t)
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("LRU CACHE STATS: %v \n", testHelper.cache.Stats())
}

func TestEviction(t *testing.T) {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil, 10000)
	if err != nil {
		log.Fatalln(err)
	}
	// retrieves information for student netIDs
	// intermediate 5 second lag enables observation of cache functionality
	for i := 0; i < 15; i++ {
		go func(helper *CampusAPIHelper, t *testing.T) {
			getStudentDo(helper, t)
		}(testHelper, t)
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("LRU CACHE STATS: %v \n", testHelper.cache.Stats().Hits)

	for i := 0; i < 15; i++ {
		go func(helper *CampusAPIHelper, t *testing.T) {
			getStudentDo(helper, t)
		}(testHelper, t)
	}

	time.Sleep(5 * time.Second)
	fmt.Printf("LRU CACHE STATS: %v \n", testHelper.cache.Stats().Hits)
}

func TestDo(t *testing.T) {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil, 100000)
	if err != nil {
		log.Fatalln(err)
	}
	// retrieves information for student netIDs
	// intermediate 5 second lag enables observation of cache functionality
	for i := 0; i < 15; i++ {
		go func(helper *CampusAPIHelper, t *testing.T) {
			getStudentDo(helper, t)
		}(testHelper, t)
	}

	time.Sleep(5 * time.Second)

	for i := 0; i < 15; i++ {
		go func(helper *CampusAPIHelper, t *testing.T) {
			getStudentDo(helper, t)
		}(testHelper, t)
	}

	time.Sleep(5 * time.Second)
}

func TestRefresh(t *testing.T) {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil, 100000)
	if err != nil {
		log.Fatalln(err)
	}
	// retrieves information for student netIDs
	// intermediate 5 second lag enables observation of cache functionality
	for i := 0; i < 15; i++ {
		go func(helper *CampusAPIHelper, i int) {
			helper.refreshAccessDebug(i)
		}(testHelper, i)
	}

	time.Sleep(5 * time.Second)
}

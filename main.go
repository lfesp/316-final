package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
	"campus-api-helper/apihelper"

	"github.com/joho/godotenv"
)

const (
	BASE_URL          = "https://api.princeton.edu:443/active-directory/1.0.5"
	REFRESH_TOKEN_URL = "https://api.princeton.edu:443/token"
)

type Student struct {
	UniversityId string `json:?universityid"`
	UID          string `json:?uid"`
}

var netids []string = []string{"liame", "hvera", "sc73", "mtouil", "shmeyer", "cjcheng", "adogra", "cabrooks", "juliacw", "aalevy", "nk5635"}

/******************************************************************************/
/*                                 Helpers                                    */
/******************************************************************************/

// getStudentDo tests CampusAPIHelper's Do() method
func getStudentDo(apiHelper *apihelper.CampusAPIHelper) {

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
	}

	fmt.Printf("%#v \n", s)
}

// getStudentGet tests CampusAPIHelper's Get() method, which utilizes an LRU cache
func getStudentGet(apiHelper *apihelper.CampusAPIHelper) {

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
/*                             Example Code                                   */
/******************************************************************************/

// demonstrate our CampusAPIHelper.Do() method by making 15 concurrent 
// HTTP requests to OIT's Active Directory API for student info
func ShowcaseDo() {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := apihelper.NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil, 100000)
	if err != nil {
		log.Fatalln(err)
	}

	// retrieves information for student netIDs
	// intermediate 5 second lag enables observation of cache functionality
	for i := 0; i < 15; i++ {
		go func(helper *apihelper.CampusAPIHelper) {
			getStudentDo(helper)
		}(testHelper)
	}

	time.Sleep(5 * time.Second)
	fmt.Println()
	fmt.Println("SECOND BURST OF REQUESTS")
	fmt.Println()
	time.Sleep(2 * time.Second)

	for i := 0; i < 15; i++ {
		go func(helper *apihelper.CampusAPIHelper) {
			getStudentDo(helper)
		}(testHelper)
	}

	time.Sleep(5 * time.Second)
}

// demonstrate our CampusAPIHelper.Get() method by making 15 concurrent 
// HTTP requests to OIT's Active Directory API for student info, with a
// limited cache size of 100000 bytes, and logging the cache statistics
func ShowcaseGet() {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := apihelper.NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil, 100000)
	if err != nil {
		log.Fatalln(err)
	}
	// retrieves information for student netIDs
	// intermediate 5 second lag enables observation of cache functionality
	for i := 0; i < 15; i++ {
		go func(helper *apihelper.CampusAPIHelper) {
			getStudentGet(helper)
		}(testHelper)
	}

	time.Sleep(5 * time.Second)
	fmt.Println()
	fmt.Printf("LRU CACHE STATS: %+v \n", *testHelper.Stats())
	fmt.Println()
	time.Sleep(5 * time.Second)

	for i := 0; i < 15; i++ {
		go func(helper *apihelper.CampusAPIHelper) {
			getStudentGet(helper)
		}(testHelper)
	}

	time.Sleep(5 * time.Second)

	fmt.Println()
	fmt.Printf("LRU CACHE STATS: %+v \n", *testHelper.Stats())
}

// demonstrate our CampusAPIHelper.Get() method by making 15 concurrent 
// HTTP requests to OIT's Active Directory API for student info, with a
// limited cache size of 10000 bytes, and logging the cache statistics
func ShowcaseEviction() {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := apihelper.NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil, 10000)
	if err != nil {
		log.Fatalln(err)
	}
	// retrieves information for student netIDs
	// intermediate 5 second lag enables observation of cache functionality
	for i := 0; i < 15; i++ {
		go func(helper *apihelper.CampusAPIHelper) {
			getStudentGet(helper)
		}(testHelper)
	}

	time.Sleep(5 * time.Second)
	fmt.Println()
	fmt.Printf("LRU CACHE STATS: %+v \n", *testHelper.Stats())
	fmt.Println()
	time.Sleep(5 * time.Second)

	for i := 0; i < 15; i++ {
		go func(helper *apihelper.CampusAPIHelper) {
			getStudentGet(helper)
		}(testHelper)
	}

	time.Sleep(5 * time.Second)
	fmt.Println()
	fmt.Printf("LRU CACHE STATS: %+v \n", *testHelper.Stats())
}

func main() {
	fmt.Println("/******************************************************************************/")
	fmt.Println("/*                         CampusAPIHelper.Do()                               */")
	fmt.Println("/******************************************************************************/")
	fmt.Println()
	time.Sleep(5 * time.Second)
	ShowcaseDo()
	time.Sleep(5 * time.Second)
	fmt.Println()

	fmt.Println("/******************************************************************************/")
	fmt.Println("/*                         CampusAPIHelper.Get()                              */")
	fmt.Println("/*                       cache size: 100000 bytes                             */")
	fmt.Println("/******************************************************************************/")
	fmt.Println()
	time.Sleep(5 * time.Second)
	ShowcaseGet()
	time.Sleep(5 * time.Second)
	fmt.Println()

	fmt.Println("/******************************************************************************/")
	fmt.Println("/*                        CampusAPIHelper.Get()                               */")
	fmt.Println("/*                       cache size: 10000 bytes                              */")
	fmt.Println("/******************************************************************************/")
	fmt.Println()
	time.Sleep(5 * time.Second)
	ShowcaseEviction()
	time.Sleep(5 * time.Second)
}

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
	Name         string `json:?displayname"`
	Email        string `json:?mail"`
}

var netids []string = []string{"liame", "hvera", "sc73", "mtouil", "shmeyer", "cjcheng", "adogra", "cabrooks", "juliacw", "aalevy", "nk5635"}

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

func main() {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := apihelper.NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil)
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

	for i := 0; i < 15; i++ {
		go func(helper *apihelper.CampusAPIHelper) {
			getStudentGet(helper)
		}(testHelper)
	}

	time.Sleep(30 * time.Second)


	// for i := 0; i < 15; i++ {
	// 	go func(i int) {
	// 		testHelper.refreshAccess(i)
	// 	}(i)
	// }
}
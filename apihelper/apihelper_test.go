package apihelper

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

const (
	REFRESH_TOKEN_URL = "https://api.princeton.edu:443/token"
)

/******************************************************************************/
/*                                  Tests                                     */
/******************************************************************************/

// test the concurrency behavior of multiple concurrent access token regeneration attempts
func TestRefresh(t *testing.T) {
	godotenv.Load(".env.local")

	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")

	testHelper, err := NewCampusAPIHelper(consumerKey, consumerSecret, REFRESH_TOKEN_URL, nil, 100000)
	if err != nil {
		log.Fatalln(err)
	}

	for i := 0; i < 15; i++ {
		go func(helper *CampusAPIHelper, i int) {
			helper.refreshAccessDebug(i)
		}(testHelper, i)
	}

	time.Sleep(5 * time.Second)
}

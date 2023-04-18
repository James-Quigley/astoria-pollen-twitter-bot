package internal

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/joho/godotenv"
	mastodon "github.com/mattn/go-mastodon"
)

type PollenApiResponsePeriodTrigger struct {
	LGID      int64
	Name      string
	Genus     string
	PlantType string
}

type PollenApiResponsePeriod struct {
	Triggers []PollenApiResponsePeriodTrigger
	Period   string
	Type     string
	Index    float64
}

type PollenApiResponseLocation struct {
	ZIP             string
	City            string
	State           string
	Periods         []PollenApiResponsePeriod `json:"periods"`
	DisplayLocation string
}

type PollenAPIResponse struct {
	Type         string
	ForecastDate string
	Location     PollenApiResponseLocation
}

type Effector func(*http.Request) (*http.Response, error)

func Retry(effector Effector, retries int, delay time.Duration) Effector {
	return func(req *http.Request) (*http.Response, error) {
		for r := 0; ; r++ {
			response, err := effector(req)
			if err == nil || r >= retries {
				return response, err
			}

			log.Printf("Attempt %d failed; retrying in %v", r+1, delay)

			select {
			case <-time.After(delay):
			}
		}
	}
}

func LoadPollenData() (PollenAPIResponse, error) {
	var result PollenAPIResponse
	req, err := http.NewRequest("GET", "https://www.pollen.com/api/forecast/current/pollen/11106", nil)

	if err != nil {
		return result, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Referer", "https://www.pollen.com/forecast/current/pollen/11106")

	client := http.Client{
		Timeout: time.Second * 1,
	}

	retryRequest := Retry(client.Do, 5, time.Millisecond*50)
	resp, err := retryRequest(req)

	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func InitTwitterAPI() *anaconda.TwitterApi {
	twitterAccessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	twitterAccessTokenSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")

	api := anaconda.NewTwitterApiWithCredentials(
		twitterAccessToken,
		twitterAccessTokenSecret,
		twitterConsumerKey,
		twitterConsumerSecret)
	return api
}

func InitMastodonAPI() *mastodon.Client {
	mastodonServerUrl := os.Getenv("MASTODON_SERVER_URL")
	mastodonAccessToken := os.Getenv("MASTODON_ACCESS_TOKEN")

	c := mastodon.NewClient(&mastodon.Config{
		Server:      mastodonServerUrl,
		AccessToken: mastodonAccessToken,
	})

	return c
}

/*
Our pollen levels are on a scale of 12. Low is 0-2.4, Low-Medium is 2.5-4.8, Medium is 4.9-7.2, High-Medium is 7.3-9.6, and High is 9.7-12.0.
*/
func FormatTweet(result PollenAPIResponse) string {
	_, month, day := time.Now().Date()
	idx := strconv.FormatFloat(result.Location.Periods[1].Index, 'f', 2, 64)
	triggers := make([]string, 0)
	for _, trigger := range result.Location.Periods[1].Triggers {
		triggers = append(triggers, trigger.Name)
	}
	scale := "High"
	num, _ := strconv.ParseFloat(idx, 2)
	switch {
	case num < 2.4:
		scale = "Low"
	case num < 4.8:
		scale = "Medium-Low"
	case num < 7.2:
		scale = "Medium"
	case num < 9.6:
		scale = "Medium-High"
	default:
		scale = "High"
	}
	var str string = "Astoria pollen level for " + month.String() + " " + strconv.Itoa(day) + ": " + idx + "/12 (" + scale + ") - " + strings.Join(triggers, ", ") + "\n" + getEmojiScale(math.Ceil(num))
	return str
}

func getEmojiScale(n float64) string {
	scale := ""
	switch n {
	case 1:
		scale = "🟩⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️"
	case 2:
		scale = "🟩🟩⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️"
	case 3:
		scale = "🟩🟩🟩⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️"
	case 4:
		scale = "🟩🟩🟩🟨⬜️⬜️⬜️⬜️⬜️⬜️⬜️⬜️"
	case 5:
		scale = "🟩🟩🟩🟨🟨⬜️⬜️⬜️⬜️⬜️⬜️⬜️"
	case 6:
		scale = "🟩🟩🟩🟨🟨🟨⬜️⬜️⬜️⬜️⬜️⬜️"
	case 7:
		scale = "🟩🟩🟩🟨🟨🟨🟧⬜️⬜️⬜️⬜️⬜️"
	case 8:
		scale = "🟩🟩🟩🟨🟨🟨🟧🟧⬜️⬜️⬜️⬜️"
	case 9:
		scale = "🟩🟩🟩🟨🟨🟨🟧🟧🟧⬜️⬜️⬜️"
	case 10:
		scale = "🟩🟩🟩🟨🟨🟨🟧🟧🟧🟥⬜️⬜️"
	case 11:
		scale = "🟩🟩🟩🟨🟨🟨🟧🟧🟧🟥🟥⬜️"
	case 12:
		scale = "🟩🟩🟩🟨🟨🟨🟧🟧🟧🟥🟥🟥"
	}
	return scale
}

func ValidateEnvVars() error {
	mastodonServerUrl := os.Getenv("MASTODON_SERVER_URL")
	mastodonAccessToken := os.Getenv("MASTODON_ACCESS_TOKEN")

	if mastodonServerUrl == "" || mastodonAccessToken == "" {
		return errors.New("Missing required Mastodon environment variables")
	}

	twitterAccessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	twitterAccessTokenSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")

	if twitterAccessToken == "" || twitterAccessTokenSecret == "" || twitterConsumerKey == "" || twitterConsumerSecret == "" {
		return errors.New("Missing required Twitter environment variables")
	}
	return nil
}

func Handle() error {
	godotenv.Load()

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))
	skipSsm := os.Getenv("SKIP_SSM_PARAMETERS")

	if skipSsm != "true" {
		ssmSvc := ssm.New(sess)
		paramPath := "/astoria-pollen"
		output, err := ssmSvc.GetParametersByPathWithContext(context.TODO(), &ssm.GetParametersByPathInput{
			Path:           &paramPath,
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			log.Fatal(err)
		}

		for _, param := range output.Parameters {
			paramPathParts := strings.Split(*param.Name, "/")
			paramName := paramPathParts[len(paramPathParts)-1]
			err = os.Setenv(paramName, *param.Value)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

	err := ValidateEnvVars()
	if err != nil {
		return err
	}

	result, err := LoadPollenData()
	if err != nil {
		return err
	}

	if result.Location.Periods[1].Index > 1 {
		str := FormatTweet(result)
		log.Println(str)
		if os.Getenv("DRY_RUN") != "true" {
			twitterApi := InitTwitterAPI()
			mastodonClient := InitMastodonAPI()
			_, twitterErr := twitterApi.PostTweet(str, nil)
			_, mastodonErr := mastodonClient.PostStatus(context.TODO(), &mastodon.Toot{
				Status: str,
			})
			if twitterErr != nil {
				return twitterErr
			}
			if mastodonErr != nil {
				return mastodonErr
			}
		}
	}
	return nil
}

package fitbit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type ActivitySummary struct {
	// Activities array of a type I'm not sure of `json:"activities"`
	Goals   Goals   `json:"goals"`
	Summary Summary `json:"summary"`
}

type Goals struct {
	ActiveMinutes int     `json:"activeMinutes"`
	CaloriesOut   int     `json:"caloriesOut"`
	Distance      float64 `json:"distance"`
	Steps         int     `json:"steps"`
}

type Summary struct {
	ActiveScore          int        `json:"activeScore"`
	ActivityCalories     int        `json:"activityCalories"`
	CaloriesBMR          int        `json:"caloriesBMR"`
	CaloriesOut          int        `json:"caloriesOut"`
	Distances            []Distance `json:"distances"`
	FairlyActiveMinutes  int        `json:"fairlyActiveMinutes"`
	LightlyActiveMinutes int        `json:"lightlyActiveMinutes"`
	MarginalCalories     int        `json:"marginalCalories"`
	SedentaryMinutes     int        `json:"sedentaryMinutes"`
	Steps                int        `json:"steps"`
	VeryActiveMinutes    int        `json:"veryActiveMinutes"`
}

type Distance struct {
	Activity string  `json:"activity"`
	Distance float64 `json:"distance"`
}

const (
	BASE_URL   = "https://api.fitbit.com/1"
	USER_AGENT = "go-fitbit-api:v0.0.1"
)

var (
	baseURL, _ = url.Parse(BASE_URL)
)

type Client struct {
	Client  *http.Client
	BaseUrl *url.URL
}

type tokenSource oauth2.Token

func (t *tokenSource) Token() (*oauth2.Token, error) {
	return (*oauth2.Token)(t), nil
}

type ConfigSource struct {
	cfg *oauth2.Config
}

func NewConfigSource(cfg *oauth2.Config) *ConfigSource {
	return &ConfigSource{
		cfg: cfg,
	}
}

func (c *ConfigSource) NewClient(tok *oauth2.Token) *Client {
	// TODO(ttacon): allow the config to have deadlines/timeouts
	// (for the context)?
	return &Client{
		Client:  c.cfg.Client(context.Background(), tok),
		BaseUrl: baseURL,
	}
}

// NewRequest creates an *http.Request with the given method, url and
// request body (if one is passed).
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	// this method is based off
	// https://github.com/google/go-github/blob/master/github/github.go:
	// NewRequest as it's a very nice way of doing this
	_, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// This is useful as this functionality works the same for the actual
	// BASE_URL and the download url (TODO(ttacon): insert download url)
	// this seems to be failing to work not RFC3986 (url resolution)
	//	resolvedUrl := c.BaseUrl.ResolveReference(parsedUrl)
	resolvedUrl, err := url.Parse(c.BaseUrl.String() + urlStr)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if body != nil {
		if err = json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, resolvedUrl.String(), buf)
	if err != nil {
		return nil, err
	}

	// TODO(ttacon): identify which headers we should add
	// e.g. "Accept", "Content-Type", "User-Agent", etc.
	req.Header.Add("User-Agent", USER_AGENT)
	return req, nil
}

// Do "makes" the request, and if there are no errors and resp is not nil,
// it attempts to unmarshal the  (json) response body into resp.
func (c *Client) Do(req *http.Request, respStr interface{}) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return nil, errors.New(fmt.Sprintf("http request failed, resp: %#v", resp))
	}

	// TODO(ttacon): maybe support passing in io.Writer as resp (downloads)?
	if respStr != nil {
		err = json.NewDecoder(resp.Body).Decode(respStr)
	}
	return resp, err
}

// yyyy-MM-dd
func (c *Client) ActivitySummaryForDay(dayString string) (ActivitySummary, error) {
	var summary ActivitySummary
	req, err := c.NewRequest(
		"GET",
		fmt.Sprintf("/user/-/activities/date/%s.json", dayString),
		nil,
	)
	if err != nil {
		return summary, err
	}

	resp, err := c.Do(req, &summary)
	if err != nil {
		return summary, err
	}
	resp.Body.Close()

	return summary, nil
}

func (c *Client) UserProfile() (UserProfile, error) {
	var profile UserProfile
	req, err := c.NewRequest("GET", "/user/-/profile.json", nil)
	if err != nil {
		return profile, err
	}

	resp, err := c.Do(req, &profile)
	if err != nil {
		return profile, err
	}
	resp.Body.Close()

	return profile, nil
}

type UserProfile struct {
	User User `json:"user"`
}

type User struct {
	StrideLengthRunningType string  `json:"strideLengthRunningType"`
	Weight                  float64 `json:"weight"`
	Age                     int     `json:"age"`
	FullName                string  `json:"fullName"`
	Gender                  string  `json:"gender"`
	GlucoseUnit             string  `json:"glucoseUnit"`
	Country                 string  `json:"country"`
	StrideLengthWalking     float64 `json:"strideLengthWalking"`
	Avatar                  string  `json:"avatar"`
	EncodedID               string  `json:"encodedId"`
	StartDayOfWeek          string  `json:"startDayOfWeek"`
	Avatar150               string  `json:"avatar150"`
	Corporate               bool    `json:"corporate"`
	DateOfBirth             string  `json:"dateOfBirth"` // 1970-01-01
	HeightUnit              string  `json:"heightUnit"`
	Locale                  string  `json:"locale"`
	MemberSince             string  `json:"memberSince"` // 2013-06-27
	OffsetFromUTCMillis     int     `json:"offsetFromUTCMillis"`
	AverageDailySteps       int     `json:"averageDailySteps"`
	Timezone                string  `json:"timezone"`
	StrideLengthRunning     float64 `json:"strideLengthRunning"`
	WeightUnit              string  `json:"weightUnit"`
	DistanceUnit            string  `json:"distanceUnit"`
	Height                  float64 `json:"height"`
	StrideLengthWalkingType string  `json:"strideLengthWalkingType"`
	DisplayName             string  `json:"displayName"`

	// topBadges
	// features
}

package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var (
	TWITTER_BEARER_TOKEN string
	TWITTER_USER_NAME    string
)

func init() {
	TWITTER_BEARER_TOKEN = os.Getenv("TWITTER_BEARER_TOKEN")
	TWITTER_USER_NAME = os.Getenv("TWITTER_USER_NAME")
}

type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	UserName        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url"`
	PublicMetrics   struct {
		FollowersCount int `json:"followers_count"`
		FollowingCount int `json:"following_count"`
		TweetCount     int `json:"tweet_count"`
		ListedCount    int `json:"listed_count"`
	} `json:"public_metrics"`
}

func (u *User) Metrics() string {
	return fmt.Sprintf("{followers=%v, following=%v, tweets=%v, listed=%v}", u.PublicMetrics.FollowersCount, u.PublicMetrics.FollowingCount, u.PublicMetrics.TweetCount, u.PublicMetrics.ListedCount)
}

func LoadSelf() (*User, error) {
	url := fmt.Sprintf("https://api.twitter.com/2/users/by/username/%s?user.fields=id,name,profile_image_url,public_metrics,username", TWITTER_USER_NAME)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+TWITTER_BEARER_TOKEN)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	type userResult struct {
		Data User `json:"data"`
	}
	var result userResult
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("status code: %d, body: %s", res.StatusCode, string(b))
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result.Data, nil
}

// type = followers or following
func ListFriends(selfID string, typ string, token string) ([]User, string, error) {
	url := fmt.Sprintf("https://api.twitter.com/2/users/%s/%s?max_results=1000&user.fields=id,name,profile_image_url,public_metrics,username", selfID, typ)
	req, _ := http.NewRequest("GET", url, nil)
	if token != "" {
		p := req.URL.Query()
		p.Add("pagination_token", token)
		req.URL.RawQuery = p.Encode()
	}
	req.Header.Add("Authorization", "Bearer "+TWITTER_BEARER_TOKEN)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	type listResult struct {
		Data []User `json:"data"`
		Meta struct {
			NextToken string `json:"next_token"`
		} `json:"meta"`
	}
	var result listResult
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, "", fmt.Errorf("status code: %d, body: %s", res.StatusCode, string(b))
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, "", err
	}
	return result.Data, result.Meta.NextToken, nil
}

func ListUsers(ids []string) ([]User, error) {
	url := "https://api.twitter.com/2/users?user.fields=id,name,profile_image_url,public_metrics,username&ids=" + strings.Join(ids, ",")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+TWITTER_BEARER_TOKEN)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	type listResult struct {
		Data []User `json:"data"`
	}
	var result listResult
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("status code: %d, body: %s", res.StatusCode, string(b))
	}
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

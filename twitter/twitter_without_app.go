package twitter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	AUTHORIZATION = "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
	USERAGENT     = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"
)

type GuestTokenResult struct {
	GuestToken string `json:"guest_token"`
}

func GenerateGuestToken(proxyUrl string) (bool, string) {
	activateUrl := "https://api.twitter.com/1.1/guest/activate.json"
	req, _ := http.NewRequest("POST", activateUrl, nil)
	req.Header.Add("Authorization", AUTHORIZATION)
	req.Header.Add("User-Agent", USERAGENT)

	proxy, _ := url.Parse(proxyUrl)
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求出错", err)
		return false, ""
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("响应异常", resp.StatusCode)
		return false, ""
	}
	var result GuestTokenResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, ""
	}
	return true, result.GuestToken
}

type UserV1 struct {
	ID              string `json:"id_str"`
	Name            string `json:"name"`
	UserName        string `json:"screen_name"`
	ProfileImageURL string `json:"profile_image_url_https"`
	FollowersCount  int    `json:"followers_count"`
	FollowingCount  int    `json:"friends_count"`
	TweetCount      int    `json:"statuses_count"`
	ListedCount     int    `json:"listed_count"`
	Verified        bool   `json:"verified"`
}

func ListFriendsByV1(userID string, typ string, nextCursor string, guestToken string) ([]UserV1, string, error) {
	reqUrl := fmt.Sprintf("https://api.twitter.com/1.1/%s/list.json", typ)

	req, _ := http.NewRequest("GET", reqUrl, nil)
	params := req.URL.Query()
	params.Add("user_id", userID)
	params.Add("skip_status", "true")
	params.Add("count", "10")
	if nextCursor != "" {
		params.Add("cursor", nextCursor)
	}
	req.URL.RawQuery = params.Encode()

	req.Header.Add("Authorization", AUTHORIZATION)
	req.Header.Add("x-guest-token", guestToken)
	req.Header.Add("User-Agent", USERAGENT)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("请求出错", err)
		return nil, "", err
	}
	defer resp.Body.Close()

	type listResult struct {
		Users      []UserV1 `json:"users"`
		NextCursor string   `json:"next_cursor_str"`
	}
	var result listResult
	if resp.StatusCode != http.StatusOK {
		fmt.Println("响应出错", resp.StatusCode)
		return nil, "", err
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("解析出错", resp.StatusCode)
		return nil, "", err
	}

	return result.Users, result.NextCursor, nil
}

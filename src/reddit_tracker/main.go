package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Fetcher interface {
	AccessToken() (accessToken string, err error)
	GetNewPosts(url string, authToken string) (body map[string]string, err error)
}

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {

	fetch := &RedditFetcher{}

	accessToken, err := fetch.GetAccessToken()
	//accessToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6IlNIQTI1NjpzS3dsMnlsV0VtMjVmcXhwTU40cWY4MXE2OWFFdWFyMnpLMUdhVGxjdWNZIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyIiwiZXhwIjoxNzcyMjM0OTUxLjIwOTU5NiwiaWF0IjoxNzcyMTQ4NTUxLjIwOTU5NiwianRpIjoiZm8yYzJ0TjhWU1FybGF0cEVOZmpLU3RubGtvSmhBIiwiY2lkIjoiN3JZaTZKVGxwMUNub18zSXdJUGhHZyIsImxpZCI6InQyXzg1MHIxcWlmIiwiYWlkIjoidDJfODUwcjFxaWYiLCJhdCI6MSwibGNhIjoxNjAwNDM0MjE5MDU4LCJzY3AiOiJlSnlLVnRKU2lnVUVBQURfX3dOekFTYyIsImZsbyI6OX0.OrYzmUMehsgbrRCYi5xAk-k9lO_NwpA5hRdZdo5tPjBiTv3zd-w8XfZGfXVGYiZ8zmou3s2D4FvijgFVykZj_cUDA39euJQ3-aBY-q-elp_VAu_GxO6aRaWbvgqDEmdI0ivHE6IY7nXuae6QCH9jbUszmFXYMIG5hHkNJo7lYxLt2G9kjL9NOtM5rOLcA4U2kP5e4oOA-qVaGZGdatoj2ek1jOwB3GrcVsL9nRlqxoa2eY9qaJ-bYE8h53LUur0aZMIF6P_mteQdSk_mrhE6H6hUp5_Pf1SS_5AfO-kpWWKsAJkwqwKsizsoDYxpAAijVFDsjzLXE0gsClShEGy1jA"
	//if err != nil {
	//	log.Fatalln(err)
	//}

	data, err := fetch.GetNewPosts("https://oauth.reddit.com", accessToken)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%v", data)
}

type RedditFetcher struct {
}

func (rf *RedditFetcher) GetAccessToken() (accessToken string, err error) {

	proxyUrl, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	// Use myClient to make requests

	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", "Certain_Albatross")
	form.Add("password", "iby5SC%X^Azq73%b")
	postBody := []byte(form.Encode())

	responseBody := bytes.NewBuffer(postBody)
	// Get the access token
	resp, err := http.NewRequest(http.MethodPost, "https://www.reddit.com/api/v1/access_token", responseBody)
	resp.Header.Add("User-Agent", "RedditTracker by Certain_Albatross")
	resp.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp.SetBasicAuth("7rYi6JTlp1Cno_3IwIPhGg", "npQCEFf4xnHLxlZ2SfiXf6RsmaLlIA")

	response, err := http.DefaultClient.Do(resp)

	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer response.Body.Close()
	//Read the response body
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	body := make(map[string]string)
	json.Unmarshal(res, &body)

	return body["access_token"], nil
}

func (rf *RedditFetcher) GetNewPosts(url string, authToken string) (body map[string]string, err error) {
	resp, err := http.NewRequest(http.MethodGet, url, nil)
	resp.Header.Add("User-Agent", "RedditTracker by Certain_Albatross")
	resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	response, err := http.DefaultClient.Do(resp)

	//Handle Error
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	//Read the response body
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(res, &body)
	return body, nil
}

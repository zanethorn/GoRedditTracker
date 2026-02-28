package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"sync"
	"time"
)

var stopThreads = false

const defaultSubreddit = "startrek"
const defaultAuthorCount = 5
const defaultPostCount = 5

type Fetcher interface {
	GetAccessToken(username string, password string, appId string, appSecret string) (err error)
	GetNewPosts(subreddit string) (posts []RedditPostWrapper, err error)
}

type RedditPost struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	UpVotes   int    `json:"ups"`
	DownVotes int    `json:"downs"`
}

type RedditPostWrapper struct {
	Data RedditPost `json:"data"`
}

type RedditResponse struct {
	Data struct {
		Children []RedditPostWrapper `json:"children"`
	} `json:"data"`
}

type PostStatistics struct {
	Posts map[string]RedditPost
	mu    sync.Mutex
}

type AuthorStatistics struct {
	Author  string
	UpVotes int
}

func (p *PostStatistics) UpdatePosts(newPosts []RedditPostWrapper) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, post := range newPosts {
		p.Posts[post.Data.ID] = post.Data
	}
}

func (p *PostStatistics) UpdateStatistics() {
	// lock the mutex to ensure thread safety while updating statistics
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Posts == nil || len(p.Posts) == 0 {
		return
	}

	// Grab author statistics
	authors := make(map[string]int)
	for _, post := range p.Posts {
		authors[post.Author] = authors[post.Author] + post.UpVotes
	}

	// Aggrigate author statistics
	authorCounts := make([]AuthorStatistics, len(authors))
	ix := 0
	for author, upVotes := range authors {
		authorCounts[ix] = AuthorStatistics{
			author,
			upVotes,
		}
		ix += 1
	}
	slices.SortFunc(authorCounts, func(a, b AuthorStatistics) int {
		return b.UpVotes - a.UpVotes // Sort in descending order
	})

	// Compute the number of authors to show based on environment variable, defaulting to 5
	// If the environment variable is not a valid integer or is less than or equal to 0, use the default value
	// If the environment variable is greater than the number of authors, use the number of authors
	authorCount, err := strconv.Atoi(os.Getenv("TRACKER_AUTHOR_COUNT"))
	if err != nil || authorCount <= 0 {
		authorCount = defaultAuthorCount
	}
	if authorCount > len(authorCounts) {
		authorCount = len(authorCounts)
	}
	authorSlices := authorCounts[:authorCount]

	// Display the top X authors
	fmt.Println("\nAuthor Statistics:")
	for i := 0; i < authorCount; i++ {
		fmt.Println(authorSlices[i].Author, "-", authorSlices[i].UpVotes)
	}

	// grab post statistics
	posts := make([]RedditPost, 0, len(p.Posts))
	for _, v := range p.Posts {
		posts = append(posts, v)
	}
	slices.SortFunc(posts, func(a, b RedditPost) int {
		return b.UpVotes - a.UpVotes // Sort in descending order based on upvotes
	})

	// Compute number of posts to show based on environment variable, defaulting to 5
	// If the environment variable is not a valid integer or is less than or equal to 0, use the default value
	// If the environment variable is greater than the number of posts, use the number of posts
	postCount, err := strconv.Atoi(os.Getenv("TRACKER_POST_COUNT"))
	if err != nil || postCount <= 0 {
		postCount = defaultPostCount
	}
	if postCount > len(posts) {
		postCount = len(posts)
	}
	postSlices := posts[:postCount]

	// Display the top X posts
	fmt.Println("\nPost Statistics:")
	for i := 0; i < postCount; i++ {
		fmt.Println(postSlices[i].Title, "-", postSlices[i].UpVotes)
	}

}

func fetchNewPosts(subreddit string, fetch Fetcher, p *PostStatistics) {
	for !stopThreads {
		data, err := fetch.GetNewPosts(subreddit)
		if err != nil {
			log.Fatalln(err)
		}

		p.mu.Lock()
		for _, post := range data {
			p.Posts[post.Data.ID] = post.Data
		}
		p.mu.Unlock()
	}

}

func updateStatistics(p *PostStatistics) {
	for !stopThreads {
		p.UpdateStatistics()
		time.Sleep(10 * time.Second)
	}
}

type RedditFetcher struct {
	accessToken string
}

func (rf *RedditFetcher) GetAccessToken(username string, password string, appId string, appSecret string) (err error) {

	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("username", username)
	form.Add("password", password)
	postBody := []byte(form.Encode())

	responseBody := bytes.NewBuffer(postBody)
	// Get the access token
	resp, err := http.NewRequest(http.MethodPost, "https://www.reddit.com/api/v1/access_token", responseBody)
	resp.Header.Add("User-Agent", "GoRedditTracker by Certain_Albatross")
	resp.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp.SetBasicAuth(appId, appSecret)

	response, err := http.DefaultClient.Do(resp)

	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer response.Body.Close()
	//Read the response body
	res, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	body := make(map[string]string)
	json.Unmarshal(res, &body)
	rf.accessToken = body["access_token"]
	return nil
}

func (rf *RedditFetcher) GetNewPosts(subreddit string) (posts []RedditPostWrapper, err error) {
	resp, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://oauth.reddit.com/r/%s/new", subreddit), nil)
	if err != nil {
		log.Fatalln(err)
	}
	resp.Header.Add("User-Agent", "RedditTracker by Certain_Albatross")
	resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", rf.accessToken))

	//fmt.Println("Fetching Fresh Data...")
	response, err := http.DefaultClient.Do(resp)
	if err != nil {
		log.Fatalln(err)
	}
	if response.StatusCode != http.StatusOK {
		log.Fatalln("Received non 200 response code")
	}

	defer response.Body.Close()
	//Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var data RedditResponse
	json.Unmarshal(body, &data)

	return data.Data.Children, nil
}

func main() {

	fetch := &RedditFetcher{}
	subreddit := os.Getenv("TRACKER_SUBREDDIT")
	if subreddit == "" {
		subreddit = defaultSubreddit
	}
	username := os.Getenv("TRACKER_USERNAME")
	password := os.Getenv("TRACKER_PASSWORD")
	appId := os.Getenv("TRACKER_APP_ID")
	appSecret := os.Getenv("TRACKER_APP_SECRET")

	err := fetch.GetAccessToken(username, password, appId, appSecret)
	if err != nil {
		log.Fatalln(err)
	}

	statistics := &PostStatistics{
		Posts: make(map[string]RedditPost),
	}
	go fetchNewPosts(subreddit, fetch, statistics)
	go updateStatistics(statistics)

	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()

	stopThreads = true
}

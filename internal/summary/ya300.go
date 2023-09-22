package summary

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/foolin/pagser"
	"github.com/go-resty/resty/v2"
)

type Ya300Summarizer struct {
	endpoint string
	token    string
	client   *resty.Client
	enabled  bool
	mu       sync.Mutex
}

func NewYa300Summarizer(endpoint string, token string) *Ya300Summarizer {
	ya := &Ya300Summarizer{
		endpoint: endpoint,
		client:   resty.New(),
		token:    token,
	}

	log.Printf("ya300 summarizer enabled: %v", token != "")

	if token != "" {
		ya.enabled = true
	}

	return ya
}

func (ya *Ya300Summarizer) Summarize(link string) (string, error) {
	ya.mu.Lock()
	defer ya.mu.Unlock()

	if !ya.enabled {
		return "", nil
	}

	body, err := json.Marshal(Ya300Request{
		ArticleURL: link,
	})

	if err != nil {
		return "", err
	}

	resp, err := ya.client.R().
		SetHeader("Authorization", "OAuth "+ya.token).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(ya.endpoint)

	if err != nil {
		return "", err
	}

	var ya300Response Ya300Response
	if err := json.Unmarshal(resp.Body(), &ya300Response); err != nil {
		return "", err
	}

	if ya300Response.Status != "success" || ya300Response.SharingURL == "" {
		return "", nil
	}

	resp, err = ya.client.R().
		Get(ya300Response.SharingURL)

	if err != nil {
		return "", err
	}

	p := pagser.New()
	var ya300SharingPageData Ya300SummaryPageData

	err = p.Parse(&ya300SharingPageData, string(resp.Body()))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(ya300SharingPageData.Summary), nil
}

type Ya300Request struct {
	ArticleURL string `json:"article_url"`
}

type Ya300Response struct {
	Status     string `json:"status"`
	SharingURL string `json:"sharing_url"`
}

type Ya300SummaryPageData struct {
	Summary string `pagser:"meta[property='og:description']->attr(content)"`
}

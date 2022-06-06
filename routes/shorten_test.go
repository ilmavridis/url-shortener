package routes

import (
	"ilmavridis/url-shortener/config"
	"ilmavridis/url-shortener/redisStorage"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func IsValidURLShortUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	if len(uuid) == 6 {
		return r.MatchString(uuid)
	}
	return false
}

func deleteRedisKey(key string) {
	redisStorage.CreateClient()
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	redisClient.Del(redisStorage.Ctx, key)
}

func TestShortenUrl(t *testing.T) {
	var requests = []request{
		{
			Url:         "www.testsite.com",
			CustomShort: "short0"},
		{
			Url:         "www.testsite2.com",
			CustomShort: "short1"},
		{
			Url: "www.noshorttestsite1.com", // Request without custom short url
		},
		{
			Url: "www.noshorttestsite2.com", // Request without custom short url
		},
	}

	config.Read()

	for _, post := range requests {

		jsonBody, err := json.Marshal(post)
		if err != nil {
			t.Errorf("Error at creating json from struct: %v", err)
		}

		req, err := http.NewRequest(http.MethodPost, "/short", strings.NewReader(string(jsonBody)))
		req.Header.Add("Content-Type", "application/json")
		if err != nil {
			t.Errorf("Error at creating HTTP request: %v", err)
		}

		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(ShortenUrl)
		handler.ServeHTTP(recorder, req)
		if status := recorder.Code; status != http.StatusOK {
			t.Errorf("Error: Handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		bodyBytes, _ := io.ReadAll(recorder.Body)
		var m map[string]interface{}
		json.Unmarshal(bodyBytes, &m)

		if m["url"] != post.Url {
			t.Errorf("Error: Returned wrong URL: got %v want %v", m["url"], post.Url)
		}

		if post.CustomShort != "" && m["short"] != post.CustomShort {
			t.Errorf("Error: Returned wrong custom short URL: got %v want %v", m["short"], post.CustomShort)
		}

		ShortUrlString := fmt.Sprintf("%v", m["short"])
		if post.CustomShort == "" && IsValidURLShortUUID(ShortUrlString) {
			t.Errorf("Error: Produced a wrong short URL which is not a 6 symbol uuid: got %v ", post.CustomShort)
		}

		// Delete the test input from the redis server
		deleteRedisKey(fmt.Sprintf("%v", m["short"]))
	}
}

func TestShortenUrlInvalidUrl(t *testing.T) {

	wrongRequest := request{}
	wrongRequest.Url = "http:foo-url.com/"

	jsonBody, err := json.Marshal(wrongRequest)
	if err != nil {
		t.Errorf("Error at creating json from struct: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/short", strings.NewReader(string(jsonBody)))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		t.Errorf("Error at creating HTTP request: %v", err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(ShortenUrl)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("Error: Handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestShortenUrlShortServerURL(t *testing.T) {
	config.Read()
	conf := config.Get()

	wrongRequest := request{}
	wrongRequest.Url = conf.Server.Address

	jsonBody, err := json.Marshal(wrongRequest)
	if err != nil {
		t.Errorf("Error at creating json from struct: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/short", strings.NewReader(string(jsonBody)))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		t.Errorf("Error at creating http request: %v", err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(ShortenUrl)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusBadRequest {
		t.Errorf("Error: Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	bodyBytes, _ := io.ReadAll(recorder.Body)
	bodyString := strings.Split(string(bodyBytes), "\n")[1]

	if bodyString != "{\"error\":\"you can't short the shortener!\"}" {
		t.Errorf("Error: json response should be {\"error\":\"you can't short the shortener!\"}. Got %v", bodyString)
	}

}

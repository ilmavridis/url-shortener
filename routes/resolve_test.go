package routes

import (
	"ilmavridis/url-shortener/config"
	"ilmavridis/url-shortener/redisStorage"

	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func addRedisKeyValue(shortURL string, URL string) {
	conf := config.Get()

	redisStorage.CreateClient()
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	redisClient.Set(redisStorage.Ctx, shortURL, URL, conf.Redis.Expiry)
}

func TestResolveURL(t *testing.T) {

	var requests = []request{
		{
			Url:         "http://www.testsite1.com",
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

	// In cases where the user does not provide a custom short URL, a 6-symbol uuid will be generated
	for i, request := range requests {
		if request.CustomShort == "" {
			requests[i].CustomShort = uuid.New().String()[:6]
		}
		addRedisKeyValue(requests[i].CustomShort, request.Url)
	}

	err := redisStorage.CreateClient()
	if err != nil {
		t.Errorf("Error at creating redis client: %v", err)
		return
	}
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	for _, request := range requests {
		path := fmt.Sprintf("/%s", request.CustomShort)
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Errorf("Error at creating http request: %v", err)
		}

		rr := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/{shortUrl}", ResolveUrl)
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusPermanentRedirect {
			t.Errorf("Error: Handler returned wrong status code: got %v want %v",
				rr.Code, http.StatusPermanentRedirect)
		}

		redirectUrl := rr.Header().Get("Location")
		if redirectUrl[0:1] == "/" {
			redirectUrl = redirectUrl[1:]
		}

		if redirectUrl != request.Url {
			t.Errorf("Error: Wrong redirect URL: got %v want %v", redirectUrl, request.Url)
		}

	}

	for _, request := range requests {
		deleteRedisKey(request.CustomShort)
	}
}

func TestResolveURLNotFound(t *testing.T) {

	nonExistedURLRequest := request{}
	nonExistedURLRequest.CustomShort = "foo123"

	err := redisStorage.CreateClient()
	if err != nil {
		t.Errorf("Error at creating redis client: %v", err)
		return
	}
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	path := fmt.Sprintf("/%s", nonExistedURLRequest.CustomShort)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Errorf("Error at creating http request: %v", err)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/{shortUrl}", ResolveUrl)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Error: Handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusBadRequest)
	}

	bodyBytes, _ := io.ReadAll(rr.Body)
	bodyString := strings.Split(string(bodyBytes), "\n")[1]

	if bodyString != "{\"error\":\"short url not found\"}" {
		t.Errorf("Error: json response should be {\"error\":\"short url not found\"}. Got %v", bodyString)
	}

}

func TestInfo(t *testing.T) {
	var requests = []request{
		{
			Url:         "http://www.testsite1.com",
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

	conf := config.Get()

	for i, request := range requests {
		if request.CustomShort == "" {
			requests[i].CustomShort = uuid.New().String()[:6]
		}
		addRedisKeyValue(requests[i].CustomShort, request.Url)
	}

	err := redisStorage.CreateClient()
	if err != nil {
		t.Errorf("Error at creating redis client: %v", err)
		return
	}
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	for _, request := range requests {
		path := fmt.Sprintf("/info/%s", request.CustomShort)
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Errorf("Error at creating http request: %v", err)
		}

		rr := httptest.NewRecorder()

		// Creates a router so that the values from the request are added to the context
		router := mux.NewRouter()
		router.HandleFunc("/info/{shortUrl}", Info)
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Error: Handler returned wrong status code: got %v want %v",
				rr.Code, http.StatusOK)
		}

		bodyBytes, _ := io.ReadAll(rr.Body)
		var m map[string]interface{}
		json.Unmarshal(bodyBytes, &m)

		if m["url"] != request.Url {
			t.Errorf("Returned wrong Url: got %v want %v", m["url"], request.Url)
		}

		if m["expires_in_seconds"] != conf.Redis.Expiry.Seconds() {
			t.Errorf("Returned wrong expiration time: got %v want %v", m["expires_in_seconds"], conf.Redis.Expiry.Seconds())
		}

	}

	// Deletes the test values from the redis server
	for _, request := range requests {
		deleteRedisKey(request.CustomShort)
	}
}

func TestInfoURLNotFound(t *testing.T) {

	nonExistedURLRequest := request{}
	nonExistedURLRequest.CustomShort = "foo123"

	err := redisStorage.CreateClient()
	if err != nil {
		t.Errorf("Error at creating redis client: %v", err)
		return
	}
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	path := fmt.Sprintf("/info/%s", nonExistedURLRequest.CustomShort)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Errorf("Error at creating http request: %v", err)
	}

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/info/{shortUrl}", ResolveUrl)
	router.ServeHTTP(rr, req)

	if rr.Code == http.StatusOK {
		t.Errorf("Error: Handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}

	bodyBytes, _ := io.ReadAll(rr.Body)
	bodyString := strings.Split(string(bodyBytes), "\n")[1]

	if bodyString != "{\"error\":\"short url not found\"}" {
		t.Errorf("Error: json response should be {\"error\":\"short url not found\"}. Got %v", bodyString)
	}

}

func TestResolveUrlhome(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("Error at creating http request: %v", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(home)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Error: Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

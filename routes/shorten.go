package routes

import (
	"ilmavridis/url-shortener/config"
	"ilmavridis/url-shortener/helpers"
	"ilmavridis/url-shortener/redisStorage"

	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/golang/gddo/httputil/header"
	"github.com/google/uuid"
)

type request struct {
	Url         string `json:"url"`
	CustomShort string `json:"short"`
}

type response struct {
	Url         string        `json:"url"`
	CustomShort string        `json:"short"`
	ExpiresIn   time.Duration `json:"expires_in_seconds"`
}

func ShortenUrl(w http.ResponseWriter, r *http.Request) {
	conf := config.Get()

	err := redisStorage.CreateClient()
	if err != nil {
		// Encodes Http response in json
		jsonResp, _ := json.Marshal(map[string]string{"error": "create redis client"}) //response to json
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	body := new(request)

	// Checks if there is the Content-Type header and has the value application/json.
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	// Json to struct
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if !govalidator.IsURL(body.Url) {
		jsonResp, _ := json.Marshal(map[string]string{"error": "invalid url"})
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		w.Write(jsonResp)
		return
	}

	// Avoids entering in an infinite loop by checking if the url provided by the user is the service url
	if !helpers.CheckDomain(body.Url, conf.Server.Address) {
		jsonResp, _ := json.Marshal(map[string]string{"error": "you can't short the shortener!"})
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		w.Write(jsonResp)
		return
	}

	// The short url can be user-defined or it will be calulcated automatically
	var shortUrl string
	if body.CustomShort == "" {
		shortUrl = uuid.New().String()[:6]
	} else {
		shortUrl = body.CustomShort
	}

	// Checks if the short url key is already taken
	val, err := redisClient.Get(redisStorage.Ctx, shortUrl).Result()
	if err != redis.Nil {
		takenMessage := fmt.Sprintf("short url %s is already taken. Short %s with another one :)", shortUrl, val)
		jsonResp, _ := json.Marshal(map[string]string{"error": takenMessage})
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		w.Write(jsonResp)
		if err != nil {
			jsonResp, _ := json.Marshal(map[string]string{"error": "conntecting to redis"})
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			w.Write(jsonResp)
			return
		}
		return
	}

	// Sends the new entry to redis server
	err = redisClient.Set(redisStorage.Ctx, shortUrl, body.Url, conf.Redis.Expiry).Err()
	if err != nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "conntecting to redis"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}

	// Returns response in json
	if err := json.NewEncoder(w).Encode(response{body.Url, shortUrl, time.Duration(conf.Redis.Expiry.Seconds())}); err != nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "encoding response to json"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}

	return
}

package routes

import (
	"ilmavridis/url-shortener/config"
	"ilmavridis/url-shortener/redisStorage"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// Returns a simple html page with further info and instructions for our service
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	staticHandler := http.FileServer(http.Dir("./"))
	staticHandler.ServeHTTP(w, r)

	return
}

// Returns images required by index.html
func ReturnImage(w http.ResponseWriter, r *http.Request) {

	imageName := mux.Vars(r)
	imagePath := "images/" + imageName["imageName"]

	fileBytes, err := ioutil.ReadFile(imagePath)
	if err != nil {
		// Encode Http response in json
		jsonResp, _ := json.Marshal(map[string]string{"error": "getting images"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)

	return
}

// Resolves short url
func ResolveUrl(w http.ResponseWriter, r *http.Request) {
	conf := config.Get()
	err := redisStorage.CreateClient()
	if err != nil {
		// Encodes Http response in json
		jsonResp, _ := json.Marshal(map[string]string{"error": "create redis client"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	shortUrl := mux.Vars(r)
	longUrl, err := redisClient.Get(redisStorage.Ctx, shortUrl["shortUrl"]).Result()

	// Sets header to return response in json
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err == redis.Nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "short url not found"})
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		w.Write(jsonResp)
		return
	} else if err != nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "conntecting to redis"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}

	http.Redirect(w, r, longUrl, http.StatusPermanentRedirect)

	// Resets redis ttl for this key/shortUrl
	err = redisClient.Expire(redisStorage.Ctx, shortUrl["shortUrl"], conf.Redis.Expiry).Err()
	if err != nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "failed to reset ttl"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}

	return
}

// Returns information for this key/shortUrl
func Info(w http.ResponseWriter, r *http.Request) {

	err := redisStorage.CreateClient()
	if err != nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "create redis client"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}
	redisClient := redisStorage.Get()
	defer redisClient.Close()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	shortUrl := mux.Vars(r)
	longUrl, err := redisClient.Get(redisStorage.Ctx, shortUrl["shortUrl"]).Result()

	if err == redis.Nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "short url not found"})
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		w.Write(jsonResp)
		return
	} else if err != nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "conntecting to redis"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}

	// Gets ttl for this this key/shortUrl
	ttl, err := redisClient.TTL(redisStorage.Ctx, shortUrl["shortUrl"]).Result()

	if err != nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "conntecting to redis"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}

	if err := json.NewEncoder(w).Encode(response{longUrl, shortUrl["shortUrl"], time.Duration(ttl.Seconds())}); err != nil {
		jsonResp, _ := json.Marshal(map[string]string{"error": "encoding response in json"})
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		w.Write(jsonResp)
		return
	}

	return
}

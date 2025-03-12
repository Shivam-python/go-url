package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type URL struct {
	ID           string    `json:"id"`
	OriginalUrl  string    `json:"original_url"`
	ShortenedUrl string    `json:"shortened_url"`
	Created_at   time.Time `json:"created_at"`
}

var host_url = "http://localhost:8000"

var UrlDb = make(map[string]URL)

func generateShortenedUrl(url string) string {
	hasher := md5.New()
	hasher.Write([]byte(url))

	data := hasher.Sum(nil)

	hash := hex.EncodeToString(data)

	return hash[:8]
}

func createUrl(url string) string {
	shortUrl := generateShortenedUrl(url)
	id := shortUrl

	UrlDb[id] = URL{
		ID:           id,
		OriginalUrl:  url,
		ShortenedUrl: shortUrl,
		Created_at:   time.Now(),
	}
	return shortUrl
}

func fetchUrl(shortUrl string) (URL, error) {
	url_data, ok := UrlDb[shortUrl]
	if ok {
		return url_data, nil
	}
	return URL{}, errors.New("URL Not found")
}

func Home(w http.ResponseWriter, r *http.Request) {

	temp, err := template.ParseFiles("home.html")
	if err == nil {
		temp.Execute(w, nil)
	}
}

func ShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid Payload", http.StatusBadRequest)
	}

	shortUrl := createUrl(data.URL)
	updated_url := strings.Join([]string{host_url, "redirect", shortUrl}, "/")

	response := struct {
		ShortUrl string `json:"shortened_url"`
	}{ShortUrl: updated_url}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func RedirecUrltHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/redirect/"):]
	url_obj, err := fetchUrl(id)
	if err != nil {
		http.Error(w, "URL Not found", http.StatusNotFound)
	}
	http.Redirect(w, r, url_obj.OriginalUrl, http.StatusFound)
}

func main() {
	fmt.Println("Starting url shortner server ....")

	http.HandleFunc("/", Home)
	http.HandleFunc("/shorten", ShortUrlHandler)
	http.HandleFunc("/redirect/", RedirecUrltHandler)

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		fmt.Println("Error starting server on port 8000 :", err)
		return
	}

}

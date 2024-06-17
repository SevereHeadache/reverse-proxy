package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var client http.Client
var authServiceUrl string

type AuthenticationForm struct {
	Html string `json:"html"`
}

type AccessToken struct {
	Value string `json:"access_token"`
}

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("No .env file found")
		os.Exit(1)
	}

	if url, exists := os.LookupEnv("AUTH_SERVICE_URL"); !exists {
		log.Fatal("Auth service url not configured")
		os.Exit(1)
	} else {
		authServiceUrl = url
	}
}

func main() {
	server := http.Server{
		Addr:    ":80",
		Handler: http.HandlerFunc(handleRequest),
	}

	log.Print("Listening on port 80")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error starting server on port 80")
	}
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	targetUrl := request.Header.Get("X-Forwarded-Host")
	log.Print("Proxy request " + targetUrl)

	if authenticated := authenticate(writer, request); !authenticated {
		return
	}

	if targetUrl == "" {
		http.Error(
			writer,
			"No X-Forwarded-Host header provided",
			http.StatusInternalServerError,
		)

		return
	}
	proxyRequest, err := http.NewRequest(
		request.Method,
		targetUrl+request.URL.Path,
		request.Body,
	)
	if err != nil {
		http.Error(
			writer,
			"Error creating proxy request",
			http.StatusInternalServerError,
		)

		return
	}

	for name, values := range request.Header {
		for _, value := range values {
			proxyRequest.Header.Add(name, value)
		}
	}

	response, err := client.Do(proxyRequest)
	if err != nil {
		http.Error(
			writer,
			"Error sending proxy request",
			http.StatusInternalServerError,
		)

		return
	}
	defer response.Body.Close()

	for name, values := range response.Header {
		for _, value := range values {
			writer.Header().Add(name, value)
		}
	}

	writer.WriteHeader(response.StatusCode)

	io.Copy(writer, response.Body)
}

func authenticate(writer http.ResponseWriter, request *http.Request) bool {
	formData := request.Header.Get("Auth-Service-Form")
	if formData != "" {
		authRequest, err := http.NewRequest(
			"POST",
			authServiceUrl+"/authenticate",
			request.Body,
		)
		if err != nil {
			http.Error(
				writer,
				"Error creating auth request",
				http.StatusInternalServerError,
			)

			return false
		}

		for name, values := range request.Header {
			for _, value := range values {
				authRequest.Header.Add(name, value)
			}
		}

		response, err := client.Do(authRequest)
		if err != nil {
			http.Error(
				writer,
				"Error sending auth request",
				http.StatusInternalServerError,
			)

			return false
		}
		defer response.Body.Close()

		writer.WriteHeader(response.StatusCode)
		io.Copy(writer, response.Body)

		return false
	}

	authRequest, err := http.NewRequest(
		"GET",
		authServiceUrl,
		nil,
	)
	if err != nil {
		http.Error(
			writer,
			"Error creating auth request",
			http.StatusInternalServerError,
		)

		return false
	}

	if cookie, err := request.Cookie("Authorization"); err == nil {
		authRequest.Header.Add("Authorization", cookie.Value)
	}
	authRequest.Header.Add("X-Client", request.Header.Get("X-Forwarded-For"))

	response, err := client.Do(authRequest)
	if err != nil {
		http.Error(
			writer,
			"Error sending auth request",
			http.StatusInternalServerError,
		)

		return false
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case 200:
		return true
	default:
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return false
		}

		var form AuthenticationForm
		if err := json.Unmarshal(body, &form); err != nil {
			http.Error(
				writer,
				"Incorrect response from auth service",
				http.StatusInternalServerError,
			)
		} else {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			writer.WriteHeader(response.StatusCode)
			fmt.Fprint(writer, form.Html)
		}

		return false
	}
}

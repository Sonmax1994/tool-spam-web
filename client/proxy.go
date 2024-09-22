package client

import (
	"dd/models"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func BuildURL(proxy models.Proxy) string {
	urlProxy := proxy.Method
	if proxy.Username != "" && proxy.Password != "" {
		urlProxy = fmt.Sprintf("%s://%s:%s@", urlProxy, proxy.Username, proxy.Password)
	}
	urlProxy = fmt.Sprintf("%s%s", urlProxy, proxy.Host)
	return urlProxy
}

func CreateProxyClient(proxyURL string) (*http.Client, error) {
	proxyURLParsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURLParsed),
		// You can add more customizations here, like TLS configuration
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30, // Set a timeout for requests
	}

	return client, nil
}

//func main() {
//	// Replace with your actual proxy URL
//	proxyURL := "http://proxy.example.com:8080"
//
//	client, err := createProxyClient(proxyURL)
//	if err != nil {
//		log.Fatalf("Failed to create proxy client: %v", err)
//	}
//
//	// Make a request to a website
//	resp, err := client.Get("https://api.ipify.org?format=json")
//	if err != nil {
//		log.Fatalf("Failed to make request: %v", err)
//	}
//	defer resp.Body.Close()
//
//	// Read and print the response
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		log.Fatalf("Failed to read response: %v", err)
//	}
//
//	fmt.Printf("Response status: %s\n", resp.Status)
//	fmt.Printf("Response body: %s\n", string(body))
//}

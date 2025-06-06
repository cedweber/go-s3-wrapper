package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	s3 "github.com/cedweber/spin-s3-api"
	"github.com/spinframework/spin-go-sdk/v2/variables"
	"github.com/spinframework/spin-go-sdk/v2/wit"
	"github.com/ydnar/wasi-http-go/wasihttp" // enable wasi-http
)

// This is purely for wit interfaces
var _ = wit.Wit

func main() {}

func init() {
	wasihttp.Serve(&WasiHTTP{})
}

type WasiHTTP struct{}

func (ww *WasiHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	getFile(w, r)
}

type FileConfig struct {
	FilePath string
}

func getFile(w http.ResponseWriter, r *http.Request) {

	var config FileConfig

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&config); err != nil {
		log.Fatal("Error: App Configuration Creation Error", err)
		http.Error(w, "App Config Error", http.StatusInternalServerError)
		return
	}

	url, err := variables.Get("base_url")
	if err != nil {
		http.Error(w, "Failed to read access_token var", 500)

	}

	accessToken, err := variables.Get("access_token")
	if err != nil {
		http.Error(w, "Failed to read access_token var", 500)
	}

	secretToken, err := variables.Get("secret_token")
	if err != nil {
		http.Error(w, "Failed to read secrets_token var", 500)
	}

	region, err := variables.Get("region")
	if err != nil {
		http.Error(w, "Failed to read region var", 500)
	}

	bucketName, err := variables.Get("bucket")
	if err != nil {
		http.Error(w, "Failed to read bucket var", 500)
	}

	cfg := s3.Config{
		Endpoint:  url,
		AccessKey: accessToken,
		SecretKey: secretToken,
		Region:    region,
	}

	httpclient := &http.Client{
		Transport: &wasihttp.Transport{},
	}

	s3Client, err := s3.New(cfg, httpclient)
	if err != nil {
		fmt.Printf("failed to create target client %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	fmt.Printf("Retrieving file information...\n")

	// Get File Size by creating a HEAD request
	resp, err := s3Client.HeadObject(ctx, bucketName, config.FilePath)
	if err != nil {
		fmt.Printf("failed to get file info %v\n", err)
	}
	contentLength, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		fmt.Printf("failed to retrieve file size info: %v\n", err)
		return
	}

	if contentLength == 0 {
		fmt.Println("File does not exist")
		return
	}

	fmt.Printf("Starting file Streaming of file %s with size %d in bucket %s  ...\n", config.FilePath, contentLength, bucketName)

	file, err := s3Client.GetObject(ctx, bucketName, config.FilePath)
	if err != nil && err != io.EOF {
		log.Fatal("Error retrieving file", err)
		http.Error(w, "Error retrieving file", http.StatusInternalServerError)
	}
	defer file.Close()

	// Blocking read all
	data, err := io.ReadAll(file)
	if err != nil && err != io.EOF {
		log.Fatal("Error reading file", err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
	}

	file.Close()

	w.Write(data)
}

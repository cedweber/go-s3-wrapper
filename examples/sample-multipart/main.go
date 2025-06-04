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
	"github.com/ydnar/wasi-http-go/wasihttp"
)

type WasiHTTP struct{}

func init() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(418)
		w.Write([]byte("Do you like tea?\n"))
		w.Write([]byte("Because I'm a teapot"))

	})
}

func main() {}

func handleCopyPart(w http.ResponseWriter, r *http.Request) {

	var config CopyConfig

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&config); err != nil {
		log.Fatal("Error: App Configuration Creation Error", err)
		http.Error(w, "App Config Error", http.StatusInternalServerError)
		return
	}

	// Create a Config with appropriate credentials.
	cfgSource := s3.Config{
		Endpoint:  config.SourceBaseDomain,
		AccessKey: config.SourceAccessKey,
		SecretKey: config.SourceSecretKey,
		Region:    config.SourceRegion,
	}

	cfgTarget := s3.Config{
		Endpoint:  config.TargetBaseDomain,
		AccessKey: config.TargetAccessKey,
		SecretKey: config.TargetSecretKey,
		Region:    config.TargetRegion,
	}

	// Create a New S3 client.
	s3ClientS, err := s3.New(cfgSource, &http.Client{
		Transport: &wasihttp.Transport{},
	})
	if err != nil {
		fmt.Printf("failed to create source client %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s3ClientT, err := s3.New(cfgTarget, &http.Client{
		Transport: &wasihttp.Transport{},
	})
	if err != nil {
		fmt.Printf("failed to create target client %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	fmt.Printf("Retrieving file information...\n")

	// Get File Size by creating a HEAD request
	resp, err := s3ClientS.HeadObject(ctx, config.SourceBucketName, config.FilePath)
	if err != nil {
		fmt.Printf("failed to get file info %v\n", err)
	}
	contentLength, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		fmt.Printf("failed to retrieve file size info: %v\n", err)
		return
	}

	if contentLength == 0 {
		fmt.Println("File not found or file does not exist")
		return
	}

	fmt.Printf("Starting file Streaming of file %s with size %d in bucket %s  ...\n", config.FilePath, contentLength, config.SourceBucketName)

	file, err := s3ClientS.GetObject(ctx, config.SourceBucketName, config.FilePath)
	if err != nil && err != io.EOF {
		log.Fatal("Error retrieving file", err)
		http.Error(w, "Error retrieving file", http.StatusInternalServerError)
	}

	fmt.Printf("File Downstream completed")

	defer file.Close()

	resp, err = s3ClientT.PutObjectStream(ctx, config.TargetBucketName, config.FilePath, file, &s3.PutObjectMetadata{
		ContentLength: contentLength,
	})
	if err != nil && err != io.EOF {
		fmt.Printf("Error reading file: %v\n", err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
	}
	resp.Body.Close()

	fmt.Printf("File Upload finished")
	w.WriteHeader(200)
	w.Header().Set("Connection", "Close")
}

// Handle File Copy
// If file size > file part => multipart upload otherwise load file into memory and upload
func handleCopy(w http.ResponseWriter, r *http.Request) {

	var config CopyConfig

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&config); err != nil {
		log.Fatal("Error: App Configuration Creation Error", err)
		http.Error(w, "App Config Error", http.StatusInternalServerError)
		return
	}

	// Create a Config with appropriate credentials.
	cfgSource := s3.Config{
		Endpoint:  config.SourceBaseDomain,
		AccessKey: config.SourceAccessKey,
		SecretKey: config.SourceSecretKey,
		Region:    config.SourceRegion,
	}

	cfgTarget := s3.Config{
		Endpoint:  config.TargetBaseDomain,
		AccessKey: config.TargetAccessKey,
		SecretKey: config.TargetSecretKey,
		Region:    config.TargetRegion,
	}

	// Create a New S3 client.
	s3ClientS, err := s3.New(cfgSource, &http.Client{
		Transport: &wasihttp.Transport{},
	})
	if err != nil {
		fmt.Printf("failed to create source client %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s3ClientT, err := s3.New(cfgTarget, &http.Client{
		Transport: &wasihttp.Transport{},
	})
	if err != nil {
		fmt.Printf("failed to create target client %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	fmt.Printf("Retrieving file information...\n")

	resp, err := s3ClientS.HeadObject(ctx, config.SourceBucketName, config.FilePath)
	if err != nil {
		fmt.Printf("failed to get file info %v\n", err)
	}
	i, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		fmt.Printf("failed to retrieve file size info %v\n", err)
	}

	if i == 0 {
		w.WriteHeader(204)
		return
	}

	fmt.Printf("Starting file Streaming of file %s with size %d in bucket %s  ...\n", config.FilePath, i, config.SourceBucketName)

	partSize := 10100000
	current := 1
	partCount := i/partSize + 1

	partsInfo := []s3.CompletedPart{}

	buf := make([]byte, partSize)

	if i > partSize {
		var tag string
		var part io.ReadCloser

		uploadData, err := s3ClientT.CreateMultipartUpload(ctx, config.TargetBucketName, config.FilePath)
		if err != nil {
			fmt.Printf("failed to create multi part file part %v\n", err)

		}

		fmt.Printf("Retrieving file %s with size %v in bucket %s\n", config.FilePath, i, config.TargetBucketName)

		fmt.Printf("Part Size: %d \n", partSize)

		for current <= partCount {

			fmt.Printf("Loading Part %d of %d\n", current, partCount)

			if current == partCount {
				part, err = s3ClientS.GetObjectPart(ctx, config.SourceBucketName, config.FilePath, int(current-1)*int(partSize), i-1)
				if err != nil {
					fmt.Printf("failed to get file part %v\n", err)
					//	http.Error(w, err), http.StatusInternalServerError)

					return
				}
				defer part.Close()

				tag, err = s3ClientT.UploadPart(ctx, config.TargetBucketName, config.FilePath, part, int(i%partSize), int(current), uploadData.UploadId)
				if err != nil && err != io.EOF && err != io.ErrClosedPipe && err != io.ErrUnexpectedEOF {
					fmt.Printf("failed to put file info %v\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				part.Close()

				partData := s3.CompletedPart{PartNumber: int(current), ETag: tag}
				partsInfo = append(partsInfo, partData)

			} else {
				part, err = s3ClientS.GetObjectPart(ctx, config.SourceBucketName, config.FilePath, int(current-1)*int(partSize), int(current*partSize-1))
				if err != nil {
					fmt.Printf("failed to get file part %v\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer part.Close()

				tag, err = s3ClientT.UploadPart(ctx, config.TargetBucketName, config.FilePath, part, partSize, current, uploadData.UploadId)
				if err != nil && err != io.EOF {
					fmt.Printf("failed to put file info %v\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				part.Close()

				partData := s3.CompletedPart{PartNumber: current, ETag: tag}
				partsInfo = append(partsInfo, partData)

			}
			current++

		}

		fmt.Printf("Completing Multipart...\n")
		err = s3ClientT.CompleteMultipartUpload(ctx, config.TargetBucketName, config.FilePath, uploadData.UploadId, partsInfo)
		if err != nil {
			fmt.Printf("Error Completing Multipart %v", err)
		}
		fmt.Printf("Copy Completed\n")
	} else {

		file, err := s3ClientS.GetObject(ctx, config.SourceBucketName, config.FilePath)
		if err != nil && err != io.EOF {
			fmt.Printf("Error retrieving object %v", err.Error())
			return
		}

		for {
			n, err := file.Read(buf)
			if n > 0 {
				break
			}
			if err != nil && err != io.EOF {
				fmt.Printf("%v", err.Error())
				break
			}

			if err != nil && err == io.EOF {
				break
			}

		}

		err = s3ClientT.PutObject(ctx, config.TargetBucketName, config.FilePath, buf)
		if err != nil {
			fmt.Printf("error putting object %v\n", err)

		}

	}

	w.WriteHeader(200)
	w.Header().Set("Connection", "Close")
}

func deleteAllExistingMultipartUploads(w http.ResponseWriter, r *http.Request) {
	var config FileBucketConfig

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	fmt.Printf("Request: List Multipart files")

	if err := dec.Decode(&config); err != nil {
		log.Fatal("Error: App Configuration Creation Error", err)
		http.Error(w, "App Config Error", http.StatusInternalServerError)
		return
	}

	// Create a Config with appropriate credentials.
	cfgSource := s3.Config{
		Endpoint:  config.BaseDomain,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Region:    config.Region,
	}

	// Create a New S3 client.
	s3ClientS, err := s3.New(cfgSource, &http.Client{
		Transport: &wasihttp.Transport{},
	})
	if err != nil {
		fmt.Printf("failed to create source client %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	parts, err := s3ClientS.ListMultipartUploads(ctx, config.BucketName, make(map[string]string))
	if err != nil {
		log.Fatal("Error: Multipart listing", err)
		http.Error(w, "Error listing files", http.StatusInternalServerError)
		return
	}

	for index, part := range parts.Uploads {
		fmt.Printf("Part %d of %d", index, len(parts.Uploads)-1)
		err = s3ClientS.AbortMultipartUpload(ctx, config.BucketName, part.Key, part.UploadId)
		if err != nil {
			log.Fatal("Error: Aborting Multipart", err)
			http.Error(w, "Error listing files", http.StatusInternalServerError)
			return
		}

	}

}

func listMultipartParts(w http.ResponseWriter, r *http.Request) {
	var config FileBucketConfig

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&config); err != nil {
		log.Fatal("Error: App Configuration Creation Error", err)
		http.Error(w, "App Config Error", http.StatusInternalServerError)
		return
	}

	// Create a Config with appropriate credentials.
	cfgSource := s3.Config{
		Endpoint:  config.BaseDomain,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Region:    config.Region,
	}

	// Create a New S3 client.
	s3ClientS, err := s3.New(cfgSource, &http.Client{
		Transport: &wasihttp.Transport{},
	})
	if err != nil {
		fmt.Printf("failed to create source client %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.Background()

	uploadData, err := s3ClientS.CreateMultipartUpload(ctx, config.BucketName, config.FilePath)
	if err != nil {
		fmt.Printf("failed to create upload %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	parts, err := s3ClientS.ListParts(ctx, config.BucketName, config.FilePath, uploadData.UploadId, make(map[string]string))
	if err != nil {
		fmt.Printf("failed to retrieve parts %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, part := range parts.Parts {
		fmt.Printf("Part-Nr: %v \n Size: %d and tag %v", part.PartNumber, part.Size, part.ETag)
	}
}

package s3

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

// S3 Client
// Source From Fermyon Spin Go SDK : https://github.com/spinframework/spin-go-sdk

// New creates a new Client.
func New(config Config) (*Client, error) {
	u, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}
	client := &Client{
		config:      config,
		endpointURL: u.String(),
	}

	return client, nil
}

// buildEndpoint returns an endpoint
func (c *Client) buildEndpoint(bucketName, path string) (string, error) {
	u, err := url.Parse(c.endpointURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse endpoint: %w", err)
	}
	if bucketName != "" {
		u.Host = bucketName + "." + u.Host
	}
	return u.JoinPath(path).String(), nil

}

// Dev State
func (c *Client) buildEndpointWithQuery(bucketName, path string, query map[string]string) (string, error) {
	u, err := url.Parse(c.endpointURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse endpoint: %w", err)
	}
	if bucketName != "" {
		u.Host = bucketName + "." + u.Host
	}

	if query != nil {
		endpoint := u.JoinPath(path)

		if len(query) > 0 {

			//	return "https://lds-test-bucket-b.object.storage.eu01.onstackit.cloud/?uploads=", nil

			queryPart := []string{}
			for k, v := range query {

				if k == "" {
					continue
				}
				queryPart = append(queryPart, fmt.Sprintf("%v=%v", k, v))
			}
			queryData := ""
			for i, v := range queryPart {
				if i == 0 {
					queryData = fmt.Sprintf("%v", v)
				} else {
					queryData = fmt.Sprintf("%v&%v", queryData, v)
				}

			}

			if path == "" {
				path = "/"
			}

			endpoint := fmt.Sprintf("%s?%s", u.JoinPath(path).String(), queryData)

			fmt.Printf("\n endpoint: %s \n", endpoint)

			return endpoint, nil

		}

		return endpoint.String(), nil
	}

	if path == "" {
		path = "/"
	}

	fmt.Printf("\n endpoint: %s \n", u.JoinPath(path).String())

	return u.JoinPath(path).String(), nil
}

func (c *Client) newRequestWithQuery(ctx context.Context, method string, bucketName string, path string, query map[string]string, body []byte) (*http.Request, error) {
	now := time.Now().UTC()
	endpointURL, err := c.buildEndpointWithQuery(bucketName, path, query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the AWS authentication headers
	// Some are required for STACKIT Object Storage
	payloadHash := getPayloadHash(&body)
	req.Header.Set("Authorization", getAuthorizationHeader(req, payloadHash, c.config.Region, c.config.AccessKey, c.config.SecretKey, now))
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", now.Format(timeFormat))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Length", fmt.Sprint(len(body)))

	return req, nil
}

func (c *Client) newRequest(ctx context.Context, method, bucketName, path string, body []byte) (*http.Request, error) {

	endpointURL, err := c.buildEndpoint(bucketName, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	now := time.Now().UTC()

	// Set the AWS authentication headers
	// Some are required for STACKIT Object Storage
	payloadHash := getPayloadHash(&body)
	req.Header.Set("Authorization", getAuthorizationHeader(req, payloadHash, c.config.Region, c.config.AccessKey, c.config.SecretKey, now))
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", now.Format(timeFormat))
	req.Header.Set("User-Agent", userAgent)

	req.Header.Set("Content-Length", fmt.Sprint(len(body)))
	return req, nil
}

func (c *Client) newRequestStream(ctx context.Context, method string, bucketName string, path string, body io.Reader) (*http.Request, error) {
	endpointURL, err := c.buildEndpoint(bucketName, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	now := time.Now().UTC()

	// Set the AWS authentication headers
	// Some are required for STACKIT Object Storage
	req.Header.Set("Authorization", getAuthorizationHeader(req, "UNSIGNED-PAYLOAD", c.config.Region, c.config.AccessKey, c.config.SecretKey, now))
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
	req.Header.Set("x-amz-date", now.Format(timeFormat))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/octet-stream")

	return req, nil
}

func (c *Client) newRequestStreamParts(ctx context.Context, method string, bucketName string, path string, partNumber int, uploadId string, body io.Reader) (*http.Request, error) {
	endpointURL, err := c.buildEndpoint(bucketName, path)
	if err != nil {
		return nil, err
	}
	endpointURLVersion := fmt.Sprintf("%s?partNumber=%d&uploadId=%s", endpointURL, partNumber, uploadId)

	req, err := http.NewRequestWithContext(ctx, method, endpointURLVersion, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	now := time.Now().UTC()

	// Set the AWS authentication headers
	// Some are required for STACKIT Object Storage
	req.Header.Set("Authorization", getAuthorizationHeader(req, "UNSIGNED-PAYLOAD", c.config.Region, c.config.AccessKey, c.config.SecretKey, now))
	req.Header.Set("x-amz-content-sha256", "UNSIGNED-PAYLOAD")
	req.Header.Set("x-amz-date", now.Format(timeFormat))
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/octet-stream")

	return req, nil
}

// do sends the request and handles any error response.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := spinhttp.Send(req)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode >= 300 {
		var errorResponse ErrorResponse
		if err := xml.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {

			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return nil, errorResponse
	}

	return resp, nil
}

// do sends the request and handles any error response.
func (c *Client) doNoXml(req *http.Request) (*http.Response, error) {
	resp, err := spinhttp.Send(req)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

func (c *Client) CreateBucket(ctx context.Context, name string) error {
	req, err := c.newRequest(ctx, http.MethodPut, "", name, nil)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	resp.Body.Close()
	return err
}

// ListBuckets returns a list of buckets.
func (c *Client) ListBuckets(ctx context.Context) (*ListBucketsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "", "", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results ListBucketsResponse
	if err := xml.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &results, nil
}

// ListObjects returns a list of objects within a specified bucket.
func (c *Client) ListObjects(ctx context.Context, bucketName string) (*ListObjectsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, bucketName, "", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results ListObjectsResponse
	if err := xml.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &results, nil
}

// HeadObject get object metadata, in this case the file size
func (c *Client) HeadObject(ctx context.Context, bucketName string, objectName string) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodHead, bucketName, objectName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetObject fetches an object.
// TODO: Create a struct to contain meta? etag,last modified, etc
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	req, err := c.newRequest(ctx, http.MethodGet, bucketName, objectName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Body, nil
}

// GetObject fetches an object.
func (c *Client) GetObjectPart(ctx context.Context, bucketName, objectName string, start int, end int) (io.ReadCloser, error) {
	req, err := c.newRequest(ctx, http.MethodGet, bucketName, objectName, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := c.do(req)
	if err != nil && err != io.EOF {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Body, nil
}

// PutObject uploads an object to the specified bucket.
func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, data []byte) error {
	req, err := c.newRequest(ctx, http.MethodPut, bucketName, objectName, data)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// PutObject uploads an object to the specified bucket.
func (c *Client) PutObjectStream(ctx context.Context, bucketName, objectName string, data io.Reader) (*http.Response, error) {
	req, err := c.newRequestStream(ctx, http.MethodPut, bucketName, objectName, data)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}

// Initiate Multipart Upload and receive the uploadId
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateMultipartUpload.html
func (c *Client) CreateMultipartUpload(ctx context.Context, bucketName string, filePath string) (*MultiPartUploadInitData, error) {

	var uploadData MultiPartUploadInitData

	query := make(map[string]string, 1)
	query["uploads"] = ""

	req, err := c.newRequestWithQuery(ctx, http.MethodPost, bucketName, filePath, query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&uploadData)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return &uploadData, nil
}

// Upload a part
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_UploadPart.html
func (c *Client) UploadPart(ctx context.Context, bucketName string, objectName string, data io.Reader, size int, partNumber int, uploadId string) (string, error) {

	req, err := c.newRequestStreamParts(ctx, http.MethodPut, bucketName, objectName, partNumber, uploadId, data)
	if err != nil && err != io.EOF {
		return "", err
	}

	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))

	resp, err := c.do(req)
	if err != nil && err != io.EOF {
		return "", err
	}

	resp.Body.Close()

	return resp.Header.Get("ETag"), nil
}

// Complete the upload
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CompleteMultipartUpload.html
func (c *Client) CompleteMultipartUpload(ctx context.Context, bucketName string, objectName string, uploadId string, parts []UploadedPart) error {

	query := make(map[string]string)
	query["uploadId"] = string(uploadId)

	completeUpload := CompleteMultipartUpload{
		Parts: parts,
	}
	xmlData, err := xml.Marshal(completeUpload)
	if err != nil {
		fmt.Printf("Error parsing response: %v", xmlData)
	}

	// Complete Writing
	endReq, err := c.newRequestWithQuery(ctx, http.MethodPost, bucketName, objectName, query, xmlData)
	if err != nil {
		return err
	}
	endReq.Header.Set("Content-Type", "application/xml")

	_, err = c.do(endReq)
	if err != nil {
		return err
	}

	return nil
}

// lists in-progress multipart uploads within a bucket
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListMultipartUploads.html
func (c *Client) ListMultipartUploads(ctx context.Context, bucketName string, query map[string]string) (*ListMultipartUploadsResult, error) {
	var listPartsResult ListMultipartUploadsResult

	query["uploads"] = ""

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}

	err = xml.Unmarshal(data, &listPartsResult)
	if err != nil {
		return nil, err
	}

	return &listPartsResult, nil
}

// Abort a previously started multipart upload
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_AbortMultipartUpload.html
func (c *Client) AbortMultipartUpload(ctx context.Context, bucketName string, filePath string, uploadId string) error {

	query := make(map[string]string)
	query["uploadId"] = uploadId

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodDelete, bucketName, filePath, query, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// List uploaded parts of specific multipart
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListParts.html
func (c *Client) ListParts(ctx context.Context, bucketName string, filePath string, uploadId string, query map[string]string) (*ListPartsResult, error) {

	queryData := make(map[string]string)

	for k, v := range queryData {
		queryData[k] = v
	}

	queryData["uploadId"] = uploadId

	var listPartsResult ListPartsResult
	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, filePath, queryData, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&listPartsResult)
	if err != nil {
		return nil, err
	}

	return &listPartsResult, nil
}

// Tagging

// Put/Update object tagging
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObjectTagging.html
func (c *Client) PutObjectTagging(ctx context.Context, bucketName string, filePath string, tagging Tagging, versionId string) (string, error) {
	var query map[string]string
	query["tagging"] = ""

	if versionId != "" {
		query["versionId"] = versionId
	}

	data, err := xml.Marshal(tagging)
	if err != nil {
		return "", err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, filePath, query, data)
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}

	return resp.Header.Get("x-amz-version-id"), nil
}

// Removes the entire tag from the specified oject
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObjectTagging.html
func (c *Client) DeleteObjectTagging(ctx context.Context, bucketName string, filePath string, query map[string]string, versionId string) error {

	query["tagging"] = ""

	if versionId != "" {
		query["versionId"] = versionId
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodDelete, bucketName, filePath, query, []byte{})
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// Retrieve object metadata
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObjectAttributes.html
func (c *Client) GetObjectAttributes(ctx context.Context, bucketName string, filePath string, query map[string]string) (*GetObjectAttributesResponse, error) {
	var attributes GetObjectAttributesResponse

	query["attributes"] = ""

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, filePath, query, []byte{})
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&attributes)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return nil, err
	}

	return &attributes, nil
}

// List all buckets
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListDirectoryBuckets.html
func (c *Client) ListDirectoryBuckets(ctx context.Context, query map[string]string) (*ListAllMyDirectoryBucketsResult, error) {
	var list ListAllMyDirectoryBucketsResult

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodGet, "", "", query, []byte{})
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return nil, err
	}

	return &list, err
}

// Website

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketWebsite.html
func (c *Client) GetBucketWebsite(ctx context.Context, bucketName string) (*WebsiteConfiguration, error) {
	var config WebsiteConfiguration
	var query map[string]string
	query["website"] = ""

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketWebsite.html
func (c *Client) PutBucketWebsite(ctx context.Context, bucketName string, config WebsiteConfiguration) error {
	var query map[string]string
	query["website"] = ""

	data, err := xml.Marshal(config)
	if err != nil {
		return err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, "", nil, data)
	if err != nil {
		return err
	}

	hash := md5.New().Sum(data)
	req.Header.Set("Content-MD5", string(hash))

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketWebsite.html
func (c *Client) DeleteBucketWebsite(ctx context.Context, bucketName string) error {
	var query map[string]string
	query["website"] = ""

	req, err := c.newRequestWithQuery(ctx, http.MethodDelete, bucketName, "", query, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil

}

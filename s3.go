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

// Create a bucket
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateBucket.html
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
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListBuckets.html
func (c *Client) ListBuckets(ctx context.Context) (*ListBucketsResponse, error) {
	var results ListBucketsResponse
	req, err := c.newRequest(ctx, http.MethodGet, "", "", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := xml.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	resp.Body.Close()

	return &results, nil
}

// ListObjects returns a list of objects within a specified bucket.
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html
func (c *Client) ListObjects(ctx context.Context, bucketName string) (*ListObjectsResponse, error) {
	var results ListObjectsResponse
	req, err := c.newRequest(ctx, http.MethodGet, bucketName, "", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := xml.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	resp.Body.Close()

	return &results, nil
}

// ListObjectsV2 returns a list of objects within a specified bucket.
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjectsV2.html
func (c *Client) ListObjectsV2(ctx context.Context, bucketName string, query map[string]string) (*ListObjectsResponse, error) {

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
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

	resp.Body.Close()

	return &results, nil
}

// ListObjectVersions returns a list of objects with metadata in a bucket
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjectVersions.html
func (c *Client) ListObjectVersions(ctx context.Context, bucketName string, query map[string]string) (*ListVersionsResult, error) {
	var results ListVersionsResult

	if query == nil {
		query = make(map[string]string)
	}

	query["versions"] = ""

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := xml.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	resp.Body.Close()

	return &results, nil
}

// HeadObject get object metadata, in this case the file size
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_HeadObject.html
func (c *Client) HeadObject(ctx context.Context, bucketName string, objectName string) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodHead, bucketName, objectName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	return resp, nil
}

// GetObject fetches an object.
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html
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
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html
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
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html
func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, data []byte) error {
	req, err := c.newRequest(ctx, http.MethodPut, bucketName, objectName, data)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

//	Delete a single specified object.
//
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObject.html
func (c *Client) DeleteObject(ctx context.Context, bucketName, objectName string, versionId string) error {

	query := make(map[string]string)
	if versionId != "" {
		query["versionId"] = versionId
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodDelete, bucketName, objectName, query, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// Delete multiple objects in a single request
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObjects.html
func (c *Client) DeleteObjects(ctx context.Context, bucketName string, objects Delete) (*DeleteResult, error) {
	var deletionResponse DeleteResult

	query := make(map[string]string)
	query["delete"] = ""

	data, err := xml.Marshal(objects)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodDelete, bucketName, "", query, data)
	if err != nil {
		return nil, err
	}

	hash := md5.New().Sum(data)
	req.Header.Set("Content-MD5", string(hash))

	_, err = c.do(req)
	if err != nil {
		return nil, err
	}

	return &deletionResponse, nil
}

// PutObject uploads an object to the specified bucket.
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html
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

// Multipart

// Initiate Multipart Upload and receive the uploadId
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateMultipartUpload.html
func (c *Client) CreateMultipartUpload(ctx context.Context, bucketName string, filePath string) (*InitiateMultipartUploadResult, error) {

	var uploadData InitiateMultipartUploadResult

	query := make(map[string]string)
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
func (c *Client) CompleteMultipartUpload(ctx context.Context, bucketName string, objectName string, uploadId string, parts []CompletedPart) error {

	query := make(map[string]string)
	query["uploadId"] = string(uploadId)

	completeUpload := CompletedMultipartUpload{
		Parts: parts,
	}
	xmlData, err := xml.Marshal(completeUpload)
	if err != nil {
		fmt.Printf("Error parsing response: %v", xmlData)
	}

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

	if query == nil {
		query = make(map[string]string)
	}

	query["uploads"] = ""

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

	resp.Body.Close()

	return &listPartsResult, nil
}

// Abort a previously started multipart upload
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_AbortMultipartUpload.html
func (c *Client) AbortMultipartUpload(ctx context.Context, bucketName string, filePath string, uploadId string) error {

	query := make(map[string]string)
	query["uploadId"] = uploadId

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

	if query == nil {
		query = make(map[string]string)
	}

	query["uploadId"] = uploadId

	var listPartsResult ListPartsResult
	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, filePath, query, nil)
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

	resp.Body.Close()

	return &listPartsResult, nil
}

// Tagging

// Put/Update object tagging
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObjectTagging.html
func (c *Client) PutObjectTagging(ctx context.Context, bucketName string, filePath string, tagging Tagging, versionId string) (string, error) {
	query := make(map[string]string)
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

	resp.Body.Close()

	return resp.Header.Get("x-amz-version-id"), nil
}

// Removes the entire tag from the specified oject
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteObjectTagging.html
func (c *Client) DeleteObjectTagging(ctx context.Context, bucketName string, filePath string, query map[string]string, versionId string) error {

	if query == nil {
		query = make(map[string]string)
	}

	query["tagging"] = ""

	if versionId != "" {
		query["versionId"] = versionId
	}

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

// Retrieve object metadata
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObjectAttributes.html
func (c *Client) GetObjectAttributes(ctx context.Context, bucketName string, filePath string, query map[string]string) (*GetObjectAttributesResponse, error) {
	var attributes GetObjectAttributesResponse

	if query == nil {
		query = make(map[string]string)
	}

	query["attributes"] = ""

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, filePath, query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&attributes)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return nil, err
	}

	resp.Body.Close()

	return &attributes, nil
}

// List all buckets
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListDirectoryBuckets.html
func (c *Client) ListDirectoryBuckets(ctx context.Context, query map[string]string) (*ListAllMyDirectoryBucketsResult, error) {
	var list ListAllMyDirectoryBucketsResult

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, "", "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return nil, err
	}

	resp.Body.Close()

	return &list, err
}

// Website

// Retrieve bucket website configuration
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketWebsite.html
func (c *Client) GetBucketWebsite(ctx context.Context, bucketName string) (*WebsiteConfiguration, error) {
	var config WebsiteConfiguration
	query := make(map[string]string)
	query["website"] = ""

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&config)
	if err != nil {
		return nil, err
	}

	resp.Body.Close()

	return &config, nil
}

// Put bucket website configuration
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketWebsite.html
func (c *Client) PutBucketWebsite(ctx context.Context, bucketName string, config WebsiteConfiguration) error {
	query := make(map[string]string)
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

// Delete bucket website
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketWebsite.html
func (c *Client) DeleteBucketWebsite(ctx context.Context, bucketName string) error {
	query := make(map[string]string)
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

// Bucket Versioning

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketVersioning.html
func (c *Client) GetBucketVersioning(ctx context.Context, bucketName string) (*VersioningConfiguration, error) {
	var config VersioningConfiguration
	query := make(map[string]string)
	query["versioning"] = ""

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

// Put Bucket Versioning
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketVersioning.html
func (c *Client) PutBucketVersioning(ctx context.Context, bucketName string, version VersioningConfiguration) error {
	query := make(map[string]string)
	query["versioning"] = ""

	data, err := xml.Marshal(version)
	if err != nil {
		return err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, "", query, data)
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

// Bucket Tagging

// get bucket tagigng
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketTagging.html
func (c *Client) GetBucketTagging(ctx context.Context, bucketName string) (*Tagging, error) {
	var tagging Tagging
	query := make(map[string]string)
	query["tagging"] = ""

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&tagging)
	if err != nil {
		return nil, err
	}

	return &tagging, nil
}

// Put bucket tagging
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketTagging.html
func (c *Client) PutBucketTagging(ctx context.Context, bucketName string, tagging Tagging) (string, error) {
	query := make(map[string]string)
	query["tagging"] = ""

	data, err := xml.Marshal(tagging)
	if err != nil {
		return "", err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, "", query, data)
	if err != nil {
		return "", err
	}

	hash := md5.New().Sum(data)
	req.Header.Set("Content-MD5", string(hash))

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	return resp.Header.Get("x-amz-version-id"), nil
}

// Delete bucket tagging
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketTagging.html
func (c *Client) DeleteBucketTagging(ctx context.Context, bucketName string) error {
	query := make(map[string]string)
	query["tagging"] = ""

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

// Object Lock

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObjectLockConfiguration.html
func (c *Client) PutObjectLockConfiguration(ctx context.Context, bucketName string, filePath string, config ObjectLockConfiguration) error {
	query := make(map[string]string)
	query["object-lock"] = ""

	data, err := xml.Marshal(config)
	if err != nil {
		return err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, filePath, query, data)
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

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObjectLockConfiguration.html
func (c *Client) GetObjectLockConfiguration(ctx context.Context, bucketName string) (*ObjectLockConfiguration, error) {
	var config ObjectLockConfiguration
	query := make(map[string]string)
	query["object-lock"] = ""

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

// Object Retention

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObjectRetention.html
func (c *Client) GetObjectRetention(ctx context.Context, bucketName string, filePath string) (*Retention, error) {
	var retention Retention
	query := make(map[string]string)
	query["retention"] = ""

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&retention)
	if err != nil {
		return nil, err
	}

	return &retention, nil
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObjectRetention.html
func (c *Client) PutObjectRetention(ctx context.Context, bucketName string, filePath string, retention Retention) error {
	query := make(map[string]string)
	query["retention"] = ""

	data, err := xml.Marshal(retention)
	if err != nil {
		return err
	}

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, "", query, data)
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

// Get object access control list
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObjectAcl.html
func (c *Client) GetObjectAcl(ctx context.Context, bucketName string, filePath string, versionId string) (*AccessControlPolicy, error) {
	var policy AccessControlPolicy
	query := make(map[string]string)
	query["acl"] = ""

	if versionId != "" {
		query["versionId"] = versionId
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, filePath, query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&policy)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// Put object access control list
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObjectAcl.html
func (c *Client) PutObjectAcl(ctx context.Context, bucketName string, filePath string, policy AccessControlPolicy) (string, error) {
	query := make(map[string]string)
	query["acl"] = ""

	data, err := xml.Marshal(policy)
	if err != nil {
		return "", err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, filePath, query, data)
	if err != nil {
		return "", err
	}

	hash := md5.New().Sum(data)
	req.Header.Set("Content-MD5", string(hash))

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}

	return resp.Header.Get("x-amz-request-charged"), nil
}

// ACL

// Get bucket access control list
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketAcl.html
func (c *Client) GetBucketAcl(ctx context.Context, bucketName string, filePath string) (*AccessControlPolicy, error) {
	var policy AccessControlPolicy
	query := make(map[string]string)
	query["acl"] = ""

	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&policy)
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// Put Bucket Access Control List
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketAcl.html
func (c *Client) PutBucketAcl(ctx context.Context, bucketName string, policy AccessControlPolicy) error {
	query := make(map[string]string)
	query["acl"] = ""

	data, err := xml.Marshal(policy)
	if err != nil {
		return err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, "", query, data)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// Logging

// Get bucket logging information
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketLogging.html
func (c *Client) GetBucketLogging(ctx context.Context, bucketName string) (*BucketLoggingStatus, error) {
	var config BucketLoggingStatus
	query := make(map[string]string)
	query["logging"] = ""

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

// Put Bucket Logging
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLogging.html
func (c *Client) PutBucketLogging(ctx context.Context, bucketName string, config BucketLoggingStatus) error {
	query := make(map[string]string)
	query["logging"] = ""

	data, err := xml.Marshal(config)
	if err != nil {
		return err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, "", query, data)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// Access Block

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetPublicAccessBlock.html
func (c *Client) GetPublicAccessBlock(ctx context.Context, bucketName string) (*PublicAccessBlockConfiguration, error) {
	var config PublicAccessBlockConfiguration
	query := make(map[string]string)
	query["publicAccessBlock"] = ""

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

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutPublicAccessBlock.html
func (c *Client) PutPublicAccessBlock(ctx context.Context, bucketName string, config PublicAccessBlockConfiguration) error {
	query := make(map[string]string)
	query["publicAccessBlock"] = ""

	data, err := xml.Marshal(config)
	if err != nil {
		return err
	}

	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, "", query, data)
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

// Remove PublicAccessBlock config for a bucket
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeletePublicAccessBlock.html
func (c *Client) DeletePublicAccessBlock(ctx context.Context, bucketName string) error {
	query := make(map[string]string)
	query["publicAccessBlock"] = ""

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

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketLifecycleConfiguration.html
func (c *Client) GetBucketLifecycleConfiguration(ctx context.Context, bucketName string) (*LifecycleConfiguration, error) {
	var config LifecycleConfiguration
	var query map[string]string
	query["lifecycle"] = ""

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

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLifecycleConfiguration.html
func (c *Client) PutBucketLifecycleConfiguration(ctx context.Context, bucketName string, lifecycle LifecycleConfiguration) (string, error) {
	var query map[string]string
	query["lifecycle"] = ""

	data, err := xml.Marshal(lifecycle)
	if err != nil {
		return "", err
	}

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodPut, bucketName, "", query, data)
	if err != nil {
		return "", err
	}

	hash := md5.New().Sum(data)
	req.Header.Set("Content-MD5", string(hash))

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}

	return resp.Header.Get("x-amz-transition-default-minimum-object-size"), nil
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_control_DeleteBucketLifecycleConfiguration.html
func (c *Client) DeleteBucketLifecycle(ctx context.Context, bucketName string) error {
	var query map[string]string
	query["lifecycle"] = ""

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodDelete, bucketName, fmt.Sprintf("/v20180820/bucket/%s/lifecycleconfiguration", bucketName), query, nil)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

// Bucket Metadata

// Get bucket metadata table config
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketMetadataTableConfiguration.html
func (c *Client) GetBucketMetadataTableConfiguratio(ctx context.Context, bucketName string) (*GetBucketMetadataTableConfigurationResult, error) {

	var metadata GetBucketMetadataTableConfigurationResult
	query := make(map[string]string)
	query["metadataTable"] = ""

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodGet, bucketName, "", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}

	err = xml.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

// Create bucket metadata table configuration
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateBucketMetadataTableConfiguration.html
func (c *Client) CreateBucketMetadataTableConfiguration(ctx context.Context, bucketName string, metadata MetadataTableConfiguration) error {

	query := make(map[string]string)
	query["metadataTable"] = ""

	data, err := xml.Marshal(metadata)
	if err != nil {
		return err
	}

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodPost, bucketName, "", query, data)
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return err
}

// Delete bucket metadata configuration
// https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketMetadataTableConfiguration.html
func (c *Client) DeleteBucketMetadataTableConfiguration(ctx context.Context, bucketName string) error {

	query := make(map[string]string)
	query["metadataTable"] = ""

	// Complete Writing
	req, err := c.newRequestWithQuery(ctx, http.MethodDelete, bucketName, "", query, []byte{})
	if err != nil {
		return err
	}

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return err
}

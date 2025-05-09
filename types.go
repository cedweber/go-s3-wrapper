package s3

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Config contains the available options for configuring a Client.
type Config struct {
	// S3 Access key ID
	AccessKey string
	// S3 Secret Access key
	SecretKey string
	// S3 Session Token
	SessionToken string
	// S3 region
	Region string
	// Endpoint is URL to the s3 service.
	Endpoint string
}

// Client provides an interface for interacting with the S3 API.
type Client struct {
	config      Config
	endpointURL string
}

type MultiPartUploadInitData struct {
	UploadId string `xml:"UploadId"`
}

type UploadedPart struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

type CompleteMultipartUpload struct {
	XmlName xml.Name       `xml:"CompleteMultipartUpload"`
	Parts   []UploadedPart `xml:"Part"`
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListBuckets.html#API_ListBuckets_ResponseSyntax
type ListBucketsResponse struct {
	Buckets []BucketInfo `xml:"Buckets>Bucket"`
	Owner   Owner
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_Bucket.html
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html#API_ListObjects_ResponseSyntax
type ListObjectsResponse struct {
	CommonPrefixes []CommonPrefix
	Contents       []ObjectInfo
	Delimiter      string
	EncodingType   string
	IsTruncated    bool
	Marker         string
	MaxKeys        int
	Name           string
	NextMarker     string
	Prefix         string
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_CommonPrefix.html
type CommonPrefix struct {
	Prefix string
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_Object.html
type ObjectInfo struct {
	Key          string
	ETag         string
	Size         int
	LastModified time.Time
	StorageClass string
	Owner        Owner
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_Owner.html
type Owner struct {
	DisplayName string
	ID          string
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html#RESTErrorResponses
type ErrorResponse struct {
	Code      string `xml:"Code"`
	Message   string `xml:"Message"`
	Resource  string `xml:"Resource"`
	RequestID string `xml:"RequestId"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type ListMultipartUploadsResult struct {
	XMLName            xml.Name     `xml:"ListMultipartUploadsResult"`
	Bucket             string       `xml:"Bucket"`
	KeyMarker          string       `xml:"KeyMarker"`
	UploadIdMarker     string       `xml:"UploadIdMarker"`
	NextKeyMarker      string       `xml:"NextKeyMarker"`
	Prefix             string       `xml:"Prefix"`
	Delimiter          string       `xml:"Delimiter"`
	NextUploadIdMarker string       `xml:"NextUploadIdMarker"`
	MaxUploads         int          `xml:"MaxUploads"`
	IsTruncated        bool         `xml:"IsTruncated"`
	Uploads            []Upload     `xml:"Upload"`
	CommonPrefixes     []PrefixItem `xml:"CommonPrefixes>Prefix"`
	EncodingType       string       `xml:"EncodingType,omitempty"`
}

type PrefixItem struct {
	Prefix string `xml:"Prefix"`
}

type Upload struct {
	ChecksumAlgorithm string `xml:"ChecksumAlgorithm,omitempty"`
	ChecksumType      string `xml:"ChecksumType,omitempty"`
	Initiated         string `xml:"Initiated"`
	Initiator         *User  `xml:"Initiator"`
	Key               string `xml:"Key"`
	Owner             *User  `xml:"Owner"`
	StorageClass      string `xml:"StorageClass"`
	UploadId          string `xml:"UploadId"`
}

type User struct {
	DisplayName string `xml:"DisplayName"`
	ID          string `xml:"ID"`
}

type ListPartsResult struct {
	XMLName              xml.Name `xml:"ListPartsResult"`
	Bucket               string   `xml:"Bucket"`
	Key                  string   `xml:"Key"`
	UploadId             string   `xml:"UploadId"`
	PartNumberMarker     int      `xml:"PartNumberMarker"`
	NextPartNumberMarker int      `xml:"NextPartNumberMarker"`
	MaxParts             int      `xml:"MaxParts"`
	IsTruncated          bool     `xml:"IsTruncated"`
	Parts                []Part   `xml:"Part"`
	Initiator            *User    `xml:"Initiator,omitempty"`
	Owner                *User    `xml:"Owner,omitempty"`
	StorageClass         string   `xml:"StorageClass,omitempty"`
	ChecksumAlgorithm    string   `xml:"ChecksumAlgorithm,omitempty"`
	ChecksumType         string   `xml:"ChecksumType,omitempty"`
}

type Part struct {
	ChecksumCRC32     string `xml:"ChecksumCRC32,omitempty"`
	ChecksumCRC32C    string `xml:"ChecksumCRC32C,omitempty"`
	ChecksumCRC64NVME string `xml:"ChecksumCRC64NVME,omitempty"`
	ChecksumSHA1      string `xml:"ChecksumSHA1,omitempty"`
	ChecksumSHA256    string `xml:"ChecksumSHA256,omitempty"`
	ETag              string `xml:"ETag"`
	LastModified      string `xml:"LastModified"`
	PartNumber        int    `xml:"PartNumber"`
	Size              int64  `xml:"Size"`
}

type GetObjectAttributesResponse struct {
	XMLName      xml.Name    `xml:"GetObjectAttributesResponse"`
	ETag         string      `xml:"ETag"`
	Checksum     Checksum    `xml:"Checksum"`
	ObjectParts  ObjectParts `xml:"ObjectParts"`
	StorageClass string      `xml:"StorageClass"`
	ObjectSize   int64       `xml:"ObjectSize"`
}

type Checksum struct {
	ChecksumCRC32     string `xml:"ChecksumCRC32"`
	ChecksumCRC32C    string `xml:"ChecksumCRC32C"`
	ChecksumCRC64NVME string `xml:"ChecksumCRC64NVME"`
	ChecksumSHA1      string `xml:"ChecksumSHA1"`
	ChecksumSHA256    string `xml:"ChecksumSHA256"`
	ChecksumType      string `xml:"ChecksumType"`
}

type ObjectParts struct {
	IsTruncated          bool   `xml:"IsTruncated"`
	MaxParts             int    `xml:"MaxParts"`
	NextPartNumberMarker int    `xml:"NextPartNumberMarker"`
	PartNumberMarker     int    `xml:"PartNumberMarker"`
	Parts                []Part `xml:"Part"`
	PartsCount           int    `xml:"PartsCount"`
}

type Tagging struct {
	XMLName xml.Name `xml:"Tagging"`
	TagSet  TagSet   `xml:"TagSet"`
}

type TagSet struct {
	Tags []Tag `xml:"Tag"`
}

type Tag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

type ListAllMyDirectoryBucketsResult struct {
	XMLName           xml.Name              `xml:"ListAllMyDirectoryBucketsResult"`
	Buckets           []DirectoryBucketInfo `xml:"Buckets>Bucket"`
	ContinuationToken string                `xml:"ContinuationToken"`
}

type DirectoryBucketInfo struct {
	BucketRegion string    `xml:"BucketRegion"`
	CreationDate time.Time `xml:"CreationDate"`
	Name         string    `xml:"Name"`
}

type WebsiteConfiguration struct {
	XMLName               xml.Name               `xml:"WebsiteConfiguration"`
	RedirectAllRequestsTo *RedirectAllRequestsTo `xml:"RedirectAllRequestsTo,omitempty"`
	IndexDocument         *IndexDocument         `xml:"IndexDocument,omitempty"`
	ErrorDocument         *ErrorDocument         `xml:"ErrorDocument,omitempty"`
	RoutingRules          []RoutingRule          `xml:"RoutingRules>RoutingRule,omitempty"`
}

type RedirectAllRequestsTo struct {
	HostName string `xml:"HostName"`
	Protocol string `xml:"Protocol,omitempty"` // optional
}

type IndexDocument struct {
	Suffix string `xml:"Suffix"`
}

type ErrorDocument struct {
	Key string `xml:"Key"`
}
type RoutingRule struct {
	Condition *Condition `xml:"Condition,omitempty"`
	Redirect  Redirect   `xml:"Redirect"`
}

type Condition struct {
	HttpErrorCodeReturnedEquals string `xml:"HttpErrorCodeReturnedEquals,omitempty"`
	KeyPrefixEquals             string `xml:"KeyPrefixEquals,omitempty"`
}

type Redirect struct {
	HostName             string `xml:"HostName,omitempty"`
	HttpRedirectCode     string `xml:"HttpRedirectCode,omitempty"`
	Protocol             string `xml:"Protocol,omitempty"`
	ReplaceKeyPrefixWith string `xml:"ReplaceKeyPrefixWith,omitempty"`
	ReplaceKeyWith       string `xml:"ReplaceKeyWith,omitempty"`
}

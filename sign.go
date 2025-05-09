package s3

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	userAgent  = "spin-s3"
	timeFormat = "20060102T150405Z"
	dateFormat = "20060102"
)
const signedHeaders = "host;x-amz-content-sha256;x-amz-date"

func getAuthorizationHeader(req *http.Request, payloadHash, region, accessKey, secretKey string, now time.Time) string {
	canonicalRequest := getCanonicalRequest(req, payloadHash, now)
	stringToSign := getStringToSign(canonicalRequest, region, now)
	signature := getSignature(stringToSign, region, secretKey, now)
	credential := strings.Join([]string{
		accessKey, now.Format(dateFormat), region, "s3", "aws4_request",
	}, "/")
	return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s, SignedHeaders=%s, Signature=%s",
		credential, signedHeaders, signature)
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html#request-string
func getStringToSign(canonicalRequest, region string, now time.Time) string {
	// Create the hash of the canonical request
	canonicalRequestHash := sha256.New()
	canonicalRequestHash.Write([]byte(canonicalRequest))
	canonicalRequestHashString := hex.EncodeToString(canonicalRequestHash.Sum(nil))

	// Create the string to sign
	return fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s/%s/s3/aws4_request\n%s",
		now.Format(timeFormat), now.Format(dateFormat), region, canonicalRequestHashString)
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html#signing-key
func getSignature(stringToSign, region, secretKey string, now time.Time) string {
	dateKey := hmacSHA256([]byte("AWS4"+secretKey), []byte(now.Format(dateFormat)))
	regionKey := hmacSHA256(dateKey, []byte(region))
	serviceKey := hmacSHA256(regionKey, []byte("s3"))
	signingKey := hmacSHA256(serviceKey, []byte("aws4_request"))

	return hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-header-based-auth.html#canonical-request
func getCanonicalRequest(req *http.Request, payloadHash string, now time.Time) string {
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		req.Host, payloadHash, now.Format(timeFormat))

	return strings.Join([]string{
		req.Method,
		req.URL.EscapedPath(),
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
}

func hmacSHA256(key, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

func getPayloadHash(payload *[]byte) string {
	hash := sha256.New()
	hash.Write(*payload)
	hashData := hex.EncodeToString(hash.Sum(nil))
	return hashData
}

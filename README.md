

````
Note:
This is a alpha WIP repository. Current structures may change in the feature 
````


## Get Started

To use the ``Spin S3 API``

1. Import the module into your spin go components ``go.mod```

````go
require github.com/cedweber/spin-s3-api v0.0.0
````


2. Use the module within your logic

````go 
s3 "github.com/cedweber/spin-s3-api"
````


3. Use the S3 API


Create a new config and a client

````go

	cfg := s3.Config{
		Endpoint:  baseDomain,
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
	}

	// Create a New S3 client.
	s3Client, err := s3.New(cfg)
	if err != nil {
		fmt.Printf("failed to create source client %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

````

Use the client to interact via REST with S3, e.g.

````go
	// Get http response from HEAD request
	resp, err := s3Client.HeadObject(ctx, bucketName, filePath)
	if err != nil {
		fmt.Printf("failed to get file info %\n", err)
	}
````

## Supported Operations

The following operations are supported:

- CreateBucket
- ListBuckets

##### Object Operations

- ListObjects
- ListObjectsV2
- ListObjectVersions
- HeadObject
- GetObject
- GetObjectPart
- PutObject
- PutObjectStream
- DeleteObject
- DeleteObjects

##### Multipart

- CreateMultipartUpload
- UploadPart
- CompleteMultipartUpload
- ListMultipartUploads
- AbortMultipartUpload
- ListParts

##### Object Tagging

- GetObjectTagging
- PutObjectTagging
- DeleteObjectTagging

##### Others

- GetObjectAttributes
- ListDirectoryBuckets

##### Bucket Website

- GetBucketWebsite
- PutBucketWebsite
- DeleteBucketWebsite

##### Bucket Versioning

- GetBucketVersioning
- PutBucketVersioning

##### Bucket Tagging

- GetBucketTagging
- PutBucketTagging
- DeleteBucketTagging

##### Object Lock Config

- PutObjectLockConfiguration
- GetObjectLockConfiguration

##### Object Retention

- GetObjectRetention
- PutObjectRetention

##### Object Access Control

- GetObjectAcl
- PutObjectAcl

##### Bucket Access Control

- GetBucketAcl
- PutBucketAcl

##### Bucket Logging

- GetBucketLogging
- PutBucketLogging

##### Public Access Block

- GetPublicAccessBlock
- PutPublicAccessBlock
- DeletePublicAccessBlock

##### Bucket Notification Configuration

- GetBucketNotificationConfiguration
- PutBucketNotificationConfiguration

##### Bucket Metrics

- GetBucketMetricsConfiguration
- ListBucketMetricsConfigurations
- PutBucketMetricsConfiguration
- DeleteBucketMetricsConfiguration

##### Object Legal Hold

- GetObjectLegalHold
- PutObjectLegalHold

##### Bucket Policy

- GetBucketPolicyStatus
- GetBucketPolicy
- PutBucketPolicy
- DeleteBucketPolicy

##### Bucket Lifecycle Configuration

- GetBucketLifecycleConfiguration
- PutBucketLifecycleConfiguration
- DeleteBucketLifecycle

##### Bucket Metadata Configuration

- GetBucketMetadataTableConfiguration
- CreateBucketMetadataTableConfiguration
- DeleteBucketMetadataTableConfiguration


## Examples

Examples can be found in the ``examples`` subfolder. 



## Contribution

If you think something is missing or wrong feel free to contribute.


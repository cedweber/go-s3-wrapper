

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
- ListObjects
- HeadObject
- GetObject
- GetObjectPart
- PutObject
- PutObjectStream
- ListMultipartUploads
- CreateMultipartUpload
- UploadPart
- CompleteMultipartUpload
- AbortMultipartUpload
- ListParts
- PutObjectTagging
- DeleteObjectTagging
- GetObjectAttributes
- ListDirectoryBuckets
- GetBucketWebsite
- PutBucketWebsite
- DeleteBucketWebsite




## Examples

We provide various examples in the ``examples`` subfolder. 



## Contribution

If you think something is missing feel free to contribute.


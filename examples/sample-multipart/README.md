

## S3 Sample Multipart File


### Pre-Requisites 

- Spin installed
- Object Storage Access



### Configuration:

This sample expects the configuration to be sent with every request on ``/trigger/copy`` as 
``text/plain`` Content-Type via POST HTTP request method. 

````json
{
"sourceBucketName" : "test-bucket-a",
"targetBucketName" : "test-bucket-b",
"sourceBaseDomain" : "https://object.storage.eu01.onstackit.cloud",
"targetBaseDomain" : "https://object.storage.eu01.onstackit.cloud",
"sourceRegion" : "eu1",
"targetRegion" : "eu1",
"sourceAccessKey" : "",
"targetAccessKey" : "",
"sourceSecretKey" : "",
"targetSecretKey" : "",
"filePath" : "/sample/file"
}
````

A valid sample call via ``curl`` could look like:

````shell
curl localhost:3000/trigger/copy -H "Content-Type: text/plain" -X POST -d '{
"sourceBucketName" : "test-bucket-a",
"targetBucketName" : "test-bucket-b",
"sourceBaseDomain" : "https://object.storage.eu01.onstackit.cloud",
"targetBaseDomain" : "https://object.storage.eu01.onstackit.cloud",
"sourceRegion" : "eu1",
"targetRegion" : "eu1",
"sourceAccessKey" : "",
"targetAccessKey" : "",
"sourceSecretKey" : "",
"targetSecretKey" : "",
"filePath" : "/sample/file"
}' -v
````


The examples assume the application to be running (locally) on port ``3000`` 



### Pre-Requisites 

- Spin installed 
- Object Storage Credentials present (Access Key + Secret Key for source and target IS)


### Configuration:

This branch expects the configuration to be sent with every request on ``/trigger/copy`` as 
``text/plain`` Content-Type via POST HTTP request method. 

````json
{
"sourceBucketName" : "lds-test-bucket-a",
"targetBucketName" : "lds-test-bucket-b",
"sourceBaseDomain" : "https://object.storage.eu01.onstackit.cloud",
"targetBaseDomain" : "https://object.storage.eu01.onstackit.cloud",
"sourceRegion" : "eu1",
"targetRegion" : "eu1",
"sourceAccessKey" : "",
"targetAccessKey" : "",
"sourceSecretKey" : "",
"targetSecretKey" : "",
"filePath" : "/sample/newDataX"
}
````

A valid sample call via ``curl`` could look like:

````shell
curl localhost:3000/trigger/copy -H "Content-Type: text/plain" -X POST -d '{
"sourceBucketName" : "lds-test-bucket-a",
"targetBucketName" : "lds-test-bucket-b",
"sourceBaseDomain" : "https://object.storage.eu01.onstackit.cloud",
"targetBaseDomain" : "https://object.storage.eu01.onstackit.cloud",
"sourceRegion" : "eu1",
"targetRegion" : "eu1",
"sourceAccessKey" : "",
"targetAccessKey" : "",
"sourceSecretKey" : "",
"targetSecretKey" : "",
"filePath" : "/sample/newDataX"
}' -v
````


The examples assume the application to be running on port ``3000`` 

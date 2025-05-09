package main

type CopyConfig struct {
	SourceBucketName string
	SourceBaseDomain string
	SourceRegion     string
	SourceAccessKey  string
	SourceSecretKey  string
	TargetBucketName string
	TargetBaseDomain string
	TargetRegion     string
	TargetAccessKey  string
	TargetSecretKey  string
	FilePath         string
}

type FileBucketConfig struct {
	BucketName string
	BaseDomain string
	Region     string
	AccessKey  string
	SecretKey  string
	FilePath   string
}

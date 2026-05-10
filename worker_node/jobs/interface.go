package jobs

import (
	"bytes"
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var s3Client *minio.Client

func InitS3() {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("Warning: Failed to initialize S3 client: %v", err)
		return
	}

	// Make a new bucket called 'chunks' if it doesn't exist.
	ctx := context.Background()
	bucketName := "chunks"
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("S3 Bucket '%s' already exists", bucketName)
		} else {
			log.Printf("Warning: Failed to create bucket: %v", err)
		}
	} else {
		log.Printf("Successfully created S3 bucket '%s'", bucketName)
	}

	s3Client = minioClient
}

type Job interface {
	Execute(payload []byte) error
	GetStatus() string
}

type StorageJob struct {
	JobID    string
	ChunkID  string
	State    string
}

func (s *StorageJob) Execute(payload []byte) error {
	s.State = "saving"
	if s3Client != nil {
		ctx := context.Background()
		bucketName := "chunks"
		objectName := s.JobID + "/" + s.ChunkID
		reader := bytes.NewReader(payload)
		
		_, err := s3Client.PutObject(ctx, bucketName, objectName, reader, int64(len(payload)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
		if err != nil {
			s.State = "failed"
			log.Printf("S3 Upload Failed for %s: %v", objectName, err)
			return err
		}
		log.Printf("Successfully uploaded %s to S3 bucket %s", objectName, bucketName)
	} else {
		log.Printf("S3 client not initialized. Dropping chunk %s", s.ChunkID)
	}

	s.State = "completed"
	return nil
}

func (s *StorageJob) GetStatus() string { return s.State }

type ComputeJob struct {
	JobID string
	Image string
	State string
	Args  []string
}

func (c *ComputeJob) Execute(payload []byte) error {
	c.State = "pulling_image"
	c.State = "completed"
	return nil
}

func (c *ComputeJob) GetStatus() string { return c.State }

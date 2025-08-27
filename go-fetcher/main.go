package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
		log.Println("Continuing with system environment variables...")
	} else {
		log.Println("Successfully loaded .env file")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Create S3 client and downloader
	s3Client := s3.NewFromConfig(cfg)
	downloader := manager.NewDownloader(s3Client)

	bucketName := os.Getenv("AWS_S3_BUCKET")
	if bucketName == "" {
		log.Fatal("AWS_S3_BUCKET environment variable is required")
	}

	log.Println("Starting S3 fetcher...")
	log.Printf("Using bucket: %s", bucketName)
	log.Printf("Using region: %s", os.Getenv("AWS_REGION"))
	
	// Perform initial fetch
	log.Println("Performing initial fetch...")
	fetchFiles(s3Client, downloader, bucketName)
	log.Println("Initial fetch complete")

	// Set up ticker for 10-minute intervals
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	log.Println("Waiting for next cycle (10 minutes)...")

	for {
		select {
		case <-ticker.C:
			log.Println("Performing scheduled fetch...")
			fetchFiles(s3Client, downloader, bucketName)
			log.Println("Fetch cycle complete")
			log.Println("Waiting for next cycle (10 minutes)...")
		}
	}
}

func fetchFiles(s3Client *s3.Client, downloader *manager.Downloader, bucketName string) {
	// Create local directory
	localDir := bucketName
	if err := os.MkdirAll(localDir, 0755); err != nil {
		log.Printf("Error creating directory: %v", err)
		return
	}

	// List all objects in the bucket
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	objectCount := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Printf("Error listing objects: %v", err)
			return
		}

		for _, obj := range page.Contents {
			objectCount++
			key := aws.ToString(obj.Key)
			
			// Skip if it's a folder (ends with /)
			if strings.HasSuffix(key, "/") {
				continue
			}

			localPath := filepath.Join(localDir, key)
			
			// Create directory structure
			dir := filepath.Dir(localPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Printf("Error creating directory %s: %v", dir, err)
				continue
			}

			// Check if file already exists and has same size
			if fileInfo, err := os.Stat(localPath); err == nil {
				if fileInfo.Size() == aws.ToInt64(obj.Size) {
					log.Printf("File already exists with same size, skipping: %s", key)
					continue
				}
			}

			log.Printf("Downloading: %s", key)
			
			// Download the file
			file, err := os.Create(localPath)
			if err != nil {
				log.Printf("Error creating file %s: %v", localPath, err)
				continue
			}

			_, err = downloader.Download(context.TODO(), file, &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(key),
			})
			
			file.Close()

			if err != nil {
				log.Printf("Error downloading %s: %v", key, err)
				os.Remove(localPath) // Clean up partial download
				continue
			}

			log.Printf("Downloaded: %s", key)
		}
	}

	if objectCount == 0 {
		log.Printf("Bucket is empty, no files to download")
	} else {
		log.Printf("Listed objects in bucket: %s", bucketName)
	}
}

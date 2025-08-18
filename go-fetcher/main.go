package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	// Get config from environment variables
	bucketName := os.Getenv("AWS_S3_BUCKET")
	if bucketName == "" {
		log.Fatal("AWS_S3_BUCKET environment variable must be set")
	}

	// Load AWS config from environment (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	downloader := manager.NewDownloader(s3Client)

	log.Printf("🚀 Starting S3 Fetcher for bucket: %s", bucketName)
	log.Println("Will fetch files every 10 minutes.")

	// Run immediately on start, then every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		log.Println("--- Starting fetch cycle ---")
		fetchAndDownloadFiles(downloader, bucketName)
		log.Println("--- Finished fetch cycle. Waiting for next run. ---")
		// The first tick happens immediately, so we run the job right away.
		// Subsequent ticks will wait for the 10-minute duration.
		if ticker.C == nil { // This is to ensure we run it once at the start.
			fetchAndDownloadFiles(downloader, bucketName)
		}
	}
}

// fetchAndDownloadFiles lists all objects and downloads them.
func fetchAndDownloadFiles(downloader *manager.Downloader, bucketName string) {
	// List objects in the bucket
	paginator := s3.NewListObjectsV2Paginator(downloader.S3, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	filesDownloaded := 0
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			log.Printf("❌ Could not list objects in bucket %s: %v", bucketName, err)
			return // Exit this cycle on error
		}

		if len(page.Contents) == 0 {
			log.Println(" bucket is empty. Nothing to download.")
			return
		}

		for _, item := range page.Contents {
			// S3 objects ending with '/' are directories, skip them
			if *item.Key == "" || (*item.Key)[len(*item.Key)-1] == '/' {
				continue
			}

			// Create local file path
			localPath := filepath.Join(bucketName, *item.Key)
			if err := os.MkdirAll(filepath.Dir(localPath), os.ModePerm); err != nil {
				log.Printf("❌ Could not create directory %s: %v", filepath.Dir(localPath), err)
				continue
			}

			// Create the file
			file, err := os.Create(localPath)
			if err != nil {
				log.Printf("❌ Could not create local file %s: %v", localPath, err)
				continue
			}
			defer file.Close()

			// Download the object from S3
			_, err = downloader.Download(context.TODO(), file, &s3.GetObjectInput{
				Bucket: aws.String(bucketName),
				Key:    item.Key,
			})

			if err != nil {
				log.Printf("❌ Failed to download %s: %v", *item.Key, err)
			} else {
				log.Printf("✅ Downloaded s3://%s/%s to %s", bucketName, *item.Key, localPath)
				filesDownloaded++
			}
		}
	}
	log.Printf("Downloaded %d files in this cycle.", filesDownloaded)
}

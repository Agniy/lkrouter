package gcp

import (
	"bufio"
	"cloud.google.com/go/storage"
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"lkrouter/config"
	"log"
	"os"
	"strings"
	"time"
)

const storageLink = "https://storage.googleapis.com/"

type Service struct {
	googleAppCredPath string
}

func (s *Service) getClient(ctx *context.Context) *storage.Client {
	client, err := storage.NewClient(*ctx, option.WithCredentialsFile(s.googleAppCredPath))
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func (s *Service) GetSignedURL(link string) (string, error) {
	ctx := context.Background()
	client := s.getClient(&ctx)
	urlComponentsString := strings.Replace(link, storageLink, "", 1)
	urlComponentsList := strings.Split(urlComponentsString, "/")

	bucket := ""
	if len(urlComponentsList) > 0 {
		bucket = urlComponentsList[0]
	}

	if bucket == "" {
		return "", errors.New("bucket not found")
	}

	path := strings.Replace(link, storageLink+bucket+"/", "", 1)

	// Generate a signed URL for an object.
	googleBacket := client.Bucket(bucket)
	url, err := googleBacket.SignedURL(path, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(60 * time.Minute),
	})

	if err != nil {
		return "", err
	}

	return url, nil
}

func (s *Service) StoreGCS(bucketName, filePath string, fileName string) (string, error) {
	// Create GCS connection
	ctx := context.Background()
	client := s.getClient(&ctx)

	defer client.Close()

	// Connect to bucket
	bucket := client.Bucket(bucketName)
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file by path: %s,with error: %v", filePath, err)
	}

	// Wright to gc bucket
	obj := bucket.Object(fileName)
	w := obj.NewWriter(ctx)
	if _, err := io.Copy(w, bufio.NewReader(file)); err != nil {
		return "", fmt.Errorf("failed to copy to bucket: %v", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close: %v", err)
	}

	fileLink := storageLink + bucketName + "/" + fileName

	return fileLink, nil
}

func (s *Service) ExplicitReads(projectID string) {
	ctx := context.Background()
	client := s.getClient(&ctx)

	defer client.Close()

	fmt.Println("Buckets:")
	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(battrs.Name)
	}
}

// NewService returns app config.
func NewService() *Service {
	cfg := config.GetConfig()
	return &Service{cfg.App.GoogleAppCredPath}
}

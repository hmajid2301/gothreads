package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	bucketName      = "gothreads"
	s3Endpoint      = "http://localhost:8333"
	s3AccessKey     = "admin"
	s3SecretKey     = "admin123"
	publicURLPrefix = "http://localhost:8333/gothreads/"
)

var s3Client *s3.S3

func initS3() error {
	sess, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3AccessKey, s3SecretKey, ""),
		Endpoint:         aws.String(s3Endpoint),
		Region:           aws.String("us-east-1"),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create S3 session: %w", err)
	}

	s3Client = s3.New(sess)

	// Create bucket if it doesn't exist
	_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Ignore error if bucket already exists
		log.Printf("Note: Bucket may already exist (error: %v)", err)
	}

	log.Printf("📦 S3 connected: %s (bucket: %s)", s3Endpoint, bucketName)
	return nil
}

type UploadImageRequest struct {
	Image string `json:"image"` // base64 encoded image (with or without data URI prefix)
}

type UploadImageResponse struct {
	URL string `json:"url"` // Public URL to the uploaded image
}

func handleS3Upload(w http.ResponseWriter, r *http.Request) {
	var req UploadImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Image == "" {
		writeError(w, http.StatusBadRequest, "Image is required")
		return
	}

	// Strip data URI prefix if present
	imageData := req.Image
	contentType := "image/png"
	if strings.HasPrefix(imageData, "data:") {
		// Extract content type
		if idx := strings.Index(imageData, ";"); idx != -1 {
			contentType = imageData[5:idx]
		}
		// Strip prefix
		if idx := strings.Index(imageData, ","); idx != -1 {
			imageData = imageData[idx+1:]
		}
	}

	// Decode base64
	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid base64 image data: "+err.Error())
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("items/%d.png", time.Now().UnixNano())

	log.Printf("Uploading image to S3: %s (%d bytes)", filename, len(imgBytes))

	// Upload to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(filename),
		Body:          bytes.NewReader(imgBytes),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(imgBytes))),
		ACL:           aws.String("public-read"),
	})

	if err != nil {
		log.Printf("ERROR: Failed to upload to S3: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to upload image: "+err.Error())
		return
	}

	// Return public URL
	publicURL := publicURLPrefix + filename

	log.Printf("Image uploaded successfully: %s", publicURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadImageResponse{
		URL: publicURL,
	})
}

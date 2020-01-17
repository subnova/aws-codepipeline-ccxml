package main

import (
	"context"
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/renameio"
)

// PersistenceProvider allows the current project state to be persisted
type PersistenceProvider interface {
	PersistProjects(projects []Project) error
}

// AWSS3PersistenceProvider persists the current project state to S3
type AWSS3PersistenceProvider struct {
	config aws.Config
	bucket string
	key    string
}

// PersistProjects to an S3 bucket
func (p *AWSS3PersistenceProvider) PersistProjects(projects []Project) error {
	var b bytes.Buffer
	Encode(projects, bufio.NewWriter(&b))

	svc := s3.New(p.config)

	cacheControl := "no-cache"
	contentType := "text/xml"
	req := svc.PutObjectRequest(&s3.PutObjectInput{
		ACL:          s3.ObjectCannedACLPublicRead,
		CacheControl: &cacheControl,
		ContentType:  &contentType,
		Body:         bytes.NewReader(b.Bytes()),
		Bucket:       &p.bucket,
		Key:          &p.key,
	})

	_, err := req.Send(context.Background())
	if err != nil {
		return fmt.Errorf("unable to persist to S3 s3://%s/%s: %v", p.bucket, p.key, err)
	}

	return nil
}

// FilePersistenceProvider persists the current project state to a local file
type FilePersistenceProvider struct {
	filename string
}

// PersistProjects to a local file
func (p *FilePersistenceProvider) PersistProjects(projects []Project) error {
	var b bytes.Buffer
	Encode(projects, bufio.NewWriter(&b))
	err := renameio.WriteFile(p.filename, b.Bytes(), os.FileMode(0666))
	if err != nil {
		return fmt.Errorf("unable to write file %s: %v", p.filename, err)
	}
	return nil
}

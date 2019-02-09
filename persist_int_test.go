// +build integration

package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestFilePersistenceProvider(t *testing.T) {
	projects := []Project{
		Project{
			Name:            "test-project",
			Activity:        ActivityBuilding,
			LastBuildStatus: LastBuildStatusSuccess,
			LastBuildTime:   "2019-01-01T00:00:00Z",
			WebURL:          "https://acme.com/build",
		},
	}

	file, err := ioutil.TempFile(os.TempDir(), "test")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(file.Name())
	fpp := FilePersistenceProvider{file.Name()}

	err = fpp.PersistProjects(projects)
	if err != nil {
		t.Fatalf("failed to persist project: %v", err)
	}

	expected := `<Projects><Project name="test-project" activity="Building" lastBuildStatus="Success" lastBuildTime="2019-01-01T00:00:00Z" webUrl="https://acme.com/build"></Project></Projects>`

	data, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	actual := fmt.Sprintf("%s", data)
	if actual != expected {
		t.Errorf(`strings did not match: got "%s" expected "%s"`, actual, expected)
	}
}

func TestAWSS3PersistenceProvider(t *testing.T) {
	bucket, ok := os.LookupEnv("TEST_BUCKET")
	if !ok {
		t.Fatalf("no test bucket specified")
	}

	projects := []Project{
		Project{
			Name:            "test-project",
			Activity:        ActivityBuilding,
			LastBuildStatus: LastBuildStatusSuccess,
			LastBuildTime:   "2019-01-01T00:00:00Z",
			WebURL:          "https://acme.com/build",
		},
	}

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		t.Fatalf("unable to load AWS config: %v", err)
	}

	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	key := "test" + hex.EncodeToString(randBytes) + ".xml"

	s3pp := AWSS3PersistenceProvider{cfg, bucket, key}

	err = s3pp.PersistProjects(projects)
	if err != nil {
		t.Fatalf("failed to persist project: %v", err)
	}

	expected := `<Projects><Project name="test-project" activity="Building" lastBuildStatus="Success" lastBuildTime="2019-01-01T00:00:00Z" webUrl="https://acme.com/build"></Project></Projects>`

	svc := s3.New(cfg)

	req := svc.GetObjectRequest(&s3.GetObjectInput{Bucket: &bucket, Key: &key})

	resp, err := req.Send()
	if err != nil {
		t.Fatalf("unable to read object from bucket: %v", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	actual := fmt.Sprintf("%s", data)
	if actual != expected {
		t.Errorf(`strings did not match: got "%s" expected "%s"`, actual, expected)
	}
}

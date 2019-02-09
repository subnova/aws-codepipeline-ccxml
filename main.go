package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	bucket   = kingpin.Flag("bucket", "The S3 bucket to write data to").Envar("BUCKET").String()
	key      = kingpin.Flag("key", "The S3 bucket key to write data to").Envar("KEY").Default("cc.xml").String()
	file     = kingpin.Flag("file", "The file to write to").String()
	isLambda = kingpin.Flag("lambda", "To run as a lambda").Default("true").Bool()
)

func updateProjectsStatus(stateProvider PipelineStateProvider, persistenceProvider PersistenceProvider) error {
	pipelineStates, err := stateProvider.GetPipelineState()
	if err != nil {
		return fmt.Errorf("unable to get state pipeline state: %v", err)
	}

	err = persistenceProvider.PersistProjects(Convert(pipelineStates))
	if err != nil {
		return fmt.Errorf("unable to persist projects data: %v", err)
	}

	return nil
}

// HandleRequest is triggered when the Lambda receives an event
func HandleRequest(ctx context.Context, event events.CodePipelineEvent) (string, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return "", err
	}

	psp := AWSPipelineStateProvider{cfg}
	s3pp := AWSS3PersistenceProvider{cfg, *bucket, *key}

	err = updateProjectsStatus(&psp, &s3pp)
	if err != nil {
		return "", err
	}

	return "Done", nil
}

func runLocally() error {
	var persistenceProvider PersistenceProvider

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return err
	}

	if *file != "" {
		persistenceProvider = &FilePersistenceProvider{*file}
	} else {
		if *bucket == "" || *key == "" {
			log.Fatal("must either specify the bucket name and key or file")
		}

		persistenceProvider = &AWSS3PersistenceProvider{cfg, *bucket, *key}
	}

	psp := AWSPipelineStateProvider{cfg}

	err = updateProjectsStatus(&psp, persistenceProvider)

	return err
}

func main() {
	kingpin.Version("0.1.0")
	kingpin.Parse()

	if *isLambda {
		if *bucket == "" || *key == "" {
			log.Fatal("must specify the bucket name and key")
		}
		lambda.Start(HandleRequest)
	} else {
		err := runLocally()
		if err != nil {
			log.Fatalf("failed to update project status: %v", err)
		}
	}
}

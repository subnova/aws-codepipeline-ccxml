package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
)

// Convert the pipeline states to Projects
func Convert(pipelineStates []PipelineState) []Project {
	projects := make([]Project, 0)

	for _, pipeline := range pipelineStates {
		for _, stage := range pipeline.StageStates {
			projects = append(projects,
				Project{
					Name:            buildName(pipeline.Name, stage),
					LastBuildStatus: buildLastBuildStatus(stage),
					Activity:        buildActivity(stage),
					LastBuildTime:   buildLastBuildTime(pipeline.Created, stage),
				})
		}
	}

	return projects
}

func buildName(name string, stage codepipeline.StageState) string {
	return fmt.Sprintf("%s :: %s", name, *stage.StageName)
}

func buildLastBuildStatus(stage codepipeline.StageState) LastBuildStatus {
	if stage.LatestExecution == nil {
		return LastBuildStatusUnknown
	}

	switch stage.LatestExecution.Status {
	case codepipeline.StageExecutionStatusFailed:
		return LastBuildStatusFailure
	case codepipeline.StageExecutionStatusSucceeded:
		return LastBuildStatusSuccess
	}

	// assume Success as no easy way to work out previous state
	return LastBuildStatusSuccess
}

func buildActivity(stage codepipeline.StageState) Activity {
	if stage.LatestExecution != nil && stage.LatestExecution.Status == codepipeline.StageExecutionStatusInProgress {
		return ActivityBuilding
	}

	return ActivitySleeping
}

func buildLastBuildTime(created time.Time, stage codepipeline.StageState) string {
	if stage.ActionStates == nil || len(stage.ActionStates) == 0 ||
		stage.ActionStates[0].LatestExecution == nil || stage.ActionStates[0].LatestExecution.LastStatusChange == nil {
		// assume last action was time pipeline was created
		return created.Format(time.RFC3339)
	}

	return stage.ActionStates[0].LatestExecution.LastStatusChange.Format(time.RFC3339)
}

package main

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
)

func createTime(rfc3339 string) time.Time {
	time, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		panic("unable to parse test date")
	}
	return time
}

func TestConvert(t *testing.T) {
	stageNames := []string{"stage-1", "stage-2", "stage-3"}
	latestExecutions := []codepipeline.StageExecution{
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusSucceeded},
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusFailed},
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusInProgress},
	}
	latestExecutionTimes := []time.Time{createTime("2019-02-06T20:33:15Z"), createTime("2019-02-06T21:14:13Z"), createTime("2019-02-07T01:12:50Z")}

	pipelineState1 := PipelineState{
		Name: "test-pipeline",
		StageStates: []codepipeline.StageState{
			codepipeline.StageState{
				StageName:       &stageNames[0],
				LatestExecution: &latestExecutions[0],
				ActionStates: []codepipeline.ActionState{
					codepipeline.ActionState{
						LatestExecution: &codepipeline.ActionExecution{
							LastStatusChange: &latestExecutionTimes[0],
						},
					},
				},
			},
			codepipeline.StageState{
				StageName:       &stageNames[1],
				LatestExecution: &latestExecutions[1],
				ActionStates: []codepipeline.ActionState{
					codepipeline.ActionState{
						LatestExecution: &codepipeline.ActionExecution{
							LastStatusChange: &latestExecutionTimes[1],
						},
					},
				},
			},
			codepipeline.StageState{
				StageName:       &stageNames[2],
				LatestExecution: &latestExecutions[2],
				ActionStates: []codepipeline.ActionState{
					codepipeline.ActionState{
						LatestExecution: &codepipeline.ActionExecution{
							LastStatusChange: &latestExecutionTimes[2],
						},
					},
				},
			},
		},
	}

	pipelineStates := []PipelineState{pipelineState1}

	projects := Convert(pipelineStates)

	if len(projects) != len(pipelineState1.StageStates) {
		t.Errorf(`Convert(%v) does not return %d projects`, pipelineState1, len(pipelineState1.StageStates))
	}

	expectedNames := []string{"test-pipeline :: stage-1", "test-pipeline :: stage-2", "test-pipeline :: stage-3"}
	expectedLastBuildStatus := []LastBuildStatus{LastBuildStatusSuccess, LastBuildStatusFailure, LastBuildStatusSuccess}
	expectedActivity := []Activity{ActivitySleeping, ActivitySleeping, ActivityBuilding}
	expectedExecutionTimes := []string{"2019-02-06T20:33:15Z", "2019-02-06T21:14:13Z", "2019-02-07T01:12:50Z"}

	for index, project := range projects {
		if project.Name != expectedNames[index] {
			t.Errorf("Convert(%v) project name %d is %s not %s", pipelineState1, index, project.Name, expectedNames[index])
		}
		if project.LastBuildStatus != expectedLastBuildStatus[index] {
			t.Errorf("Convert(%v) last build status %d is %s not %s", pipelineState1, index, project.LastBuildStatus, expectedLastBuildStatus[index])
		}
		if project.Activity != expectedActivity[index] {
			t.Errorf("Convert(%v) activity %d is %s not %s", pipelineState1, index, project.Activity, expectedActivity[index])
		}
		if project.LastBuildTime != expectedExecutionTimes[index] {
			t.Errorf("Convert(%v) last build time %d is %s not %s", pipelineState1, index, project.LastBuildTime, expectedActivity[index])
		}
	}
}

func TestBuildName(t *testing.T) {
	stageName := "stage-1"
	expectName := "test-pipeline :: stage-1"
	actualName := buildName("test-pipeline", codepipeline.StageState{StageName: &stageName})
	if actualName != expectName {
		t.Errorf(`buildName(...) is %s not %s`, actualName, expectName)
	}
}

func TestBuildLastBuildStatus(t *testing.T) {
	inputs := []codepipeline.StageExecution{
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusInProgress},
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusFailed},
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusSucceeded},
	}

	expectedOutputs := []LastBuildStatus{LastBuildStatusSuccess, LastBuildStatusFailure, LastBuildStatusSuccess}

	for index, input := range inputs {
		actual := buildLastBuildStatus(codepipeline.StageState{LatestExecution: &input})
		if actual != expectedOutputs[index] {
			t.Errorf(`buildLastBuildStatus("%s") is %s not %s`, input.Status, actual, expectedOutputs[index])
		}
	}

	actual := buildLastBuildStatus(codepipeline.StageState{})
	if actual != LastBuildStatusUnknown {
		t.Errorf("buildLastBuildStatus(nil) is %s not %s", actual, LastBuildStatusUnknown)
	}
}

func TestBuildActivity(t *testing.T) {
	inputs := []codepipeline.StageExecution{
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusInProgress},
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusFailed},
		codepipeline.StageExecution{Status: codepipeline.StageExecutionStatusSucceeded},
	}

	expectedOutputs := []Activity{ActivityBuilding, ActivitySleeping, ActivitySleeping}

	for index, input := range inputs {
		actual := buildActivity(codepipeline.StageState{LatestExecution: &input})
		if actual != expectedOutputs[index] {
			t.Errorf(`buildActivity("%s") is %s not %s`, input.Status, actual, expectedOutputs[index])
		}
	}

	actual := buildActivity(codepipeline.StageState{})
	if actual != ActivitySleeping {
		t.Errorf("buildActivity(nil) is %s not %s", actual, ActivitySleeping)
	}
}

func TestBuildLastBuildTime(t *testing.T) {
	created := "2019-02-01T12:00:00Z"
	expected := "2019-02-06T20:33:15Z"
	lastStatusChange := createTime(expected)
	input := codepipeline.StageState{
		ActionStates: []codepipeline.ActionState{
			codepipeline.ActionState{
				LatestExecution: &codepipeline.ActionExecution{
					LastStatusChange: &lastStatusChange,
				},
			},
		},
	}

	actual := buildLastBuildTime(createTime(created), input)

	if actual != expected {
		t.Errorf(`buildLastBuildTime(%v) is %s not %s`, input, actual, expected)
	}

	input = codepipeline.StageState{
		ActionStates: []codepipeline.ActionState{
			codepipeline.ActionState{},
		},
	}

	actual = buildLastBuildTime(createTime(created), input)
	if actual != created {
		t.Errorf(`buildLastBuildTime(%v) is %s not %s`, input, actual, created)
	}
}

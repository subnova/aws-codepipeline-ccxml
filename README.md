# AWS CodePipeline CCTray XML feed support

Many CI/CD servers provide an XML feed that provides status information about the pipelines its provides.  It is known as the CCTray XML feed format and can be used to drive build monitors, such as [Nevergreen](https://github.com/build-canaries/nevergreen).

This project adds support for generating this file for [AWS CodePipeline](https://docs.aws.amazon.com/codepipeline/latest/userguide/welcome.html).

## How it works

AWS CodePipeline emits CloudWatch events whenever the status of a stage within a pipeline changes.  This project provides an AWS Lambda that collects CodePipeline status information and writes it, in CCTray XML feed format, to an S3 bucket whenever one of these events is emitted.

If the S3 bucket has been configured as an S3 website, then build monitors can access the feed over HTTP.

## Setup

1. Create the S3 bucket and configure it as an S3 website 
2. Create the Lambda function, ensure the associated IAM role has the required permissions
3. Create a CloudWatch Events rule that matches the AWS CodePipeline status change events
4. Configure the rule to trigger the Lambda function

### CloudWatch Event Rule

```json
{
  "source": [
    "aws.codepipeline"
  ],
  "detail-type": [
    "CodePipeline Stage Execution State Change"
  ]
}
```

### IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
          "codepipeline:ListPipelines",
          "codepipeline:GetPipelineState"
      ],
      "Resource": [
          "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
          "s3:PutObject"
      ],
      "Resource": [
        "arn:aws:s3:::<bucket>/<key>"
      ]
    }
  ]
}
```
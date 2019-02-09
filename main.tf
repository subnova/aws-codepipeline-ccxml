variable "bucket" {
  description = "The bucket that will contain the feed"
}

variable "key" {
  description = "The key within the bucket that will contain the feed"
  default     = "cc.xml"
}

resource "aws_s3_bucket" "ccxml" {
  bucket = "${var.bucket}"
  acl    = "public-read"

  website {
    index_document = "cc.xml"
  }
}

output "website" {
  value = "https://${aws_s3_bucket.ccxml.website_endpoint}/${var.key}"
}

module "ccxml" {
  source = "howdio/lambda/aws//modules/package"

  name = "ccxml"
  path = "${path.module}/aws-codepipeline-ccxml"
}

resource "aws_lambda_function" "ccxml" {
  filename         = "${module.ccxml.path}"
  function_name    = "ccxml"
  handler          = "aws-codepipeline-ccxml"
  description      = "Handler that responds to CodePipeline events by updating a CCTray XML feed"
  memory_size      = 128
  timeout          = 20
  runtime          = "go1.x"
  source_code_hash = "${module.ccxml.base64sha256}"
  role             = "${aws_iam_role.ccxml.arn}"

  environment {
    variables {
      BUCKET = "${var.bucket}"
      KEY    = "${var.key}"
    }
  }
}

data "aws_iam_policy_document" "ccxml_assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type = "Service"

      identifiers = [
        "lambda.amazonaws.com",
      ]
    }
  }
}

resource "aws_iam_role" "ccxml" {
  name = "ccxml"

  assume_role_policy = "${data.aws_iam_policy_document.ccxml_assume_role_policy.json}"
}

data "aws_iam_policy_document" "ccxml_role_policy" {
  statement {
    effect = "Allow"

    actions = [
      "codepipeline:ListPipelines",
      "codepipeline:GetPipelineState",
    ]

    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"

    actions = [
      "s3:PutObject",
      "s3:PutObjectAcl"
    ]

    resources = [
      "arn:aws:s3:::${var.bucket}/${var.key}",
    ]
  }

  statement {
      effect = "Allow"

      actions = [
          "logs:*"
      ]

      resources = [
          "*"
      ]
  }
}

resource "aws_iam_role_policy" "ccxml" {
  role = "${aws_iam_role.ccxml.id}"

  policy = "${data.aws_iam_policy_document.ccxml_role_policy.json}"
}

locals {
  event_pattern = {
    source      = ["aws.codepipeline"]
    detail-type = ["CodePipeline Stage Execution State Change"]
  }

  event_pattern_json = "${jsonencode(local.event_pattern)}"
}

resource "aws_cloudwatch_event_rule" "ccxml" {
  name        = "ccxml"
  description = "Rule that matches CodePipeline State Execution State Changes"

  event_pattern = "${local.event_pattern_json}"
}

resource "aws_cloudwatch_event_target" "ccxml" {
  target_id = "ccxml"
  rule      = "${aws_cloudwatch_event_rule.ccxml.name}"
  arn       = "${aws_lambda_function.ccxml.arn}"
}

resource "aws_lambda_permission" "ccxml" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.ccxml.function_name}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.ccxml.arn}"
}

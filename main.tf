terraform {
  backend "remote" {
    hostname     = "app.terraform.io"
    organization = "Quames"

    workspaces {
      name = "astoria-pollen-twitter"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_lambda_function" "astoria_pollen_twitter" {
  s3_bucket     = "james-lambda-builds"
  s3_key        = "astoria-pollen-twitter/astoria-pollen-twitter.zip"
  function_name = "astoria-pollen-twitter"
  role          = aws_iam_role.astoria_pollen_twitter.arn
  handler       = "astoria-pollen-twitter"
  timeout       = 15

  runtime = "go1.x"

  lifecycle {
    ignore_changes = [
      environment,
    ]
  }
}

resource "aws_cloudwatch_event_rule" "astoria_pollen_twitter" {
  name                = "astoria-pollen-twitter-invocation"
  description         = "Runs the astoria-pollen-twitter bot daily"
  schedule_expression = "cron(0 11 * * ? *)"
}

resource "aws_cloudwatch_event_target" "astoria_pollen_twitter" {
  rule      = aws_cloudwatch_event_rule.astoria_pollen_twitter.name
  target_id = "astoria_pollen_twitter"
  arn       = aws_lambda_function.astoria_pollen_twitter.arn
}

resource "aws_lambda_permission" "astoria_pollen_twitter" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.astoria_pollen_twitter.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.astoria_pollen_twitter.arn
}

data "aws_iam_policy_document" "astoria_pollen_twitter_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }

}

data "aws_iam_policy_document" "astoria_pollen_twitter_permissions" {
  statement {
    actions = [
      "ssm:GetParametersByPath",
      "ssm:GetParameters",
      "ssm:GetParameter"
    ]
    resources = [
      "arn:aws:ssm:*:579709515411:parameter/astoria-pollen/*",
      "arn:aws:ssm:*:579709515411:parameter/astoria-pollen"
    ]
  }
}

resource "aws_iam_role" "astoria_pollen_twitter" {
  name               = "astoria-pollen-twitter"
  assume_role_policy = data.aws_iam_policy_document.astoria_pollen_twitter_assume_role.json
}

resource "aws_iam_policy" "astoria_pollen_twitter_policy" {
  policy = data.aws_iam_policy_document.astoria_pollen_twitter_permissions.json
}

resource "aws_iam_role_policy_attachment" "astoria_pollen_twitter_policy" {
  role       = aws_iam_role.astoria_pollen_twitter.name
  policy_arn = aws_iam_policy.astoria_pollen_twitter_policy.arn
}

resource "aws_iam_policy" "astoria_pollen_twitter_lambda_logging" {
  name        = "astoria_pollen_twitter_lambda_logging"
  path        = "/"
  description = "IAM policy for logging from a lambda"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "astoria_pollen_twitter_lambda_logs" {
  role       = aws_iam_role.astoria_pollen_twitter.name
  policy_arn = aws_iam_policy.astoria_pollen_twitter_lambda_logging.arn
}

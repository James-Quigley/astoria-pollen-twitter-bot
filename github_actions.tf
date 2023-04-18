resource "aws_iam_user" "astoria_pollen_twitter" {
  name = "astoria_pollen_twitter"
}

data "aws_iam_policy_document" "astoria_pollen_twitter_policy" {
  statement {
    actions = [
      "s3:GetAccessPoint",
      "s3:PutAccountPublicAccessBlock",
      "s3:GetAccountPublicAccessBlock",
      "s3:ListAllMyBuckets",
      "s3:ListAccessPoints",
      "s3:ListJobs",
      "s3:CreateJob",
      "s3:HeadBucket"
    ]

    resources = ["*"]
  }

  statement {
    actions = [
      "s3:*"
    ]
    resources = [
      "arn:aws:s3:::james-lambda-builds",
      "arn:aws:s3:::james-lambda-builds/*"
    ]
  }
}

resource "aws_iam_policy" "astoria_pollen_twitter" {
  name   = "astoria_pollen_twitter"
  policy = data.aws_iam_policy_document.astoria_pollen_twitter_policy.json
}

resource "aws_iam_user_policy_attachment" "astoria_pollen_twitter" {
  policy_arn = aws_iam_policy.astoria_pollen_twitter.arn
  user       = aws_iam_user.astoria_pollen_twitter.name
}

data "aws_iam_policy_document" "lambda_update_policy" {
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
      "iam:ListRoles",
      "lambda:UpdateFunctionCode",
      "lambda:CreateFunction",
      "lambda:GetFunction",
      "lambda:UpdateFunctionConfiguration",
      "lambda:GetFunctionConfiguration"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "astoria_pollen_twitter_lambda" {
  name   = "astoria_pollen_twitter_lambda"
  policy = data.aws_iam_policy_document.lambda_update_policy.json
}

resource "aws_iam_user_policy_attachment" "astoria_pollen_twitter_lambda" {
  policy_arn = aws_iam_policy.astoria_pollen_twitter_lambda.arn
  user       = aws_iam_user.astoria_pollen_twitter.name
}

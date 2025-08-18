# Configure the AWS provider
provider "aws" {
  region = "ap-south-1" # You can change this to your preferred region
}

# Create a random string to ensure the S3 bucket name is unique
resource "random_pet" "bucket_name" {
  prefix = "my-unique-bucket"
  length = 2
}

# Create an S3 bucket
resource "aws_s3_bucket" "app_bucket" {
  bucket = random_pet.bucket_name.id
}

# Create an IAM user for the Go application
resource "aws_iam_user" "app_user" {
  name = "go-s3-fetcher-user"
}

# Create an access key for the IAM user
resource "aws_iam_access_key" "app_user_key" {
  user = aws_iam_user.app_user.name
}

# Define an IAM policy that grants read-only access to the S3 bucket
resource "aws_iam_policy" "read_only_bucket_policy" {
  name        = "S3-ReadOnly-Policy"
  description = "Allows read-only access to a specific S3 bucket"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ],
        Resource = [
          aws_s3_bucket.app_bucket.arn,
          "${aws_s3_bucket.app_bucket.arn}/*" # Important: Allows access to objects inside the bucket
        ]
      }
    ]
  })
}

# Attach the policy to the user
resource "aws_iam_user_policy_attachment" "app_user_attach" {
  user       = aws_iam_user.app_user.name
  policy_arn = aws_iam_policy.read_only_bucket_policy.arn
}

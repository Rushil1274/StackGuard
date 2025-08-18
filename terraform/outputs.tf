output "s3_bucket_name" {
  value = aws_s3_bucket.app_bucket.id
}

output "app_user_access_key_id" {
  value = aws_iam_access_key.app_user_key.id
}

output "app_user_secret_access_key" {
  value     = aws_iam_access_key.app_user_key.secret
  sensitive = true # Marks the output as sensitive
}

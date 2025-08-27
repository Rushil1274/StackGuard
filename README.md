# AWS S3 Provisioning and Go Fetcher.

This project contains two main parts:
1.  **Terraform & GitHub Actions**: Automatically provisions an AWS S3 bucket and an IAM user with read-only access.
2.  **Go S3 Fetcher**: A Go application that periodically downloads all files from the S3 bucket to a local directory.

---

## 🚀 Part 1: Terraform & CI/CD Setup

### Step 1: AWS Credentials for Terraform

The GitHub Actions pipeline needs AWS credentials with permissions to create S3 buckets and IAM resources.

1.  Create an IAM user in your AWS account with `AdministratorAccess` (for simplicity in this project).
2.  Generate an access key and secret key for this user.

### Step 2: Configure GitHub Repository Secrets

1.  In your GitHub repository, go to **Settings** -> **Secrets and variables** -> **Actions**.
2.  Click **New repository secret**.
3.  Create two secrets:
    * `AWS_ACCESS_KEY_ID`: Your AWS access key ID.
    * `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key.

### Step 3: Run the Pipeline

Commit and push all the files to your `main` branch. The GitHub Actions workflow will automatically trigger, run `terraform apply`, and provision your AWS resources.

Check the **Actions** tab in your repository to see the output. It will display the `s3_bucket_name`, `app_user_access_key_id`, and `app_user_secret_access_key`. You will need these for the Go application.

---

## 🏃 Part 2: Running the Go S3 Fetcher

### Step 1: Set Environment Variables

The Go application reads its configuration from environment variables. Set them in your terminal.

```bash
# Get these values from the Terraform output in your GitHub Actions log
export AWS_ACCESS_KEY_ID="<app_user_access_key_id_from_output>"
export AWS_SECRET_ACCESS_KEY="<app_user_secret_access_key_from_output>"
export AWS_S3_BUCKET="<s3_bucket_name_from_output>"

# Set the AWS region to match the one in your terraform/main.tf file
export AWS_REGION="us-east-1"
```

### Step 2: Run the Application

Navigate to the `go-fetcher` directory and run the program.

```bash
# First, initialize the Go module (only needs to be done once)
cd go-fetcher
go mod init my-fetcher
go mod tidy

# Now, run the application
go run main.go
```

The application will start, perform an initial fetch, and then run again every 10 minutes. It will create a local folder named after your S3 bucket and download the files into it, preserving the folder structure from S3.

**To test it**, you can manually upload some files (e.g., `folder1/test.txt` and `folder2/image.jpg`) to your S3 bucket using the AWS Console and watch the program download them.

---

## ✅ How Edge Cases Are Handled

* **Network Interruptions**: The official **AWS SDK for Go v2 has a built-in default retry mechanism** with exponential backoff. It automatically retries failed API calls due to transient network issues.
* **Empty Containers/Bucket**: If the S3 bucket is empty, the `ListObjectsV2` API call returns an empty list. The program logs a message and waits for the next cycle. No errors occur.
* **Large Files**: The Go app uses `manager.NewDownloader`, which is part of the AWS SDK. It is **optimized for performance and handles large files by streaming them in concurrent parts**, so it doesn't load the entire file into memory.
* **Token Expiry**: This solution uses long-lived IAM user access keys, which do not expire. If you were using temporary credentials (e.g., from an IAM role), the AWS SDK's default credential provider would handle **automatic token refreshing** behind the scenes.

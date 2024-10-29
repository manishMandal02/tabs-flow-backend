#!/bin/bash

# Usage message function
usage() {
  echo "Usage: $0 -p <AWS_PROFILE> -r <AWS_REGION>"
  exit 1
}

# Parse command-line flags
while getopts ":p:r:" opt; do
  case $opt in
    p) AWS_PROFILE="$OPTARG" ;;
    r) AWS_REGION="$OPTARG" ;;
    *) usage ;;
  esac
done

# Check if both profile and region are provided
if [ -z "$AWS_PROFILE" ] || [ -z "$AWS_REGION" ]; then
  usage
fi

# Define bucket pattern parts
BUCKET_PREFIX="cdk"
BUCKET_KEYWORD="assets"

# Find the bucket matching the specified pattern
BUCKET_NAME=$(aws s3api list-buckets --query "Buckets[?starts_with(Name, '$BUCKET_PREFIX') && contains(Name, '$BUCKET_KEYWORD') && contains(Name, '$AWS_REGION')].Name" --output text --profile "$AWS_PROFILE")

# Check if a bucket was found
if [ -z "$BUCKET_NAME" ]; then
  echo "No bucket found matching pattern: ${BUCKET_PREFIX}*${BUCKET_KEYWORD}*${AWS_REGION}"
  exit 1
fi

echo "Found bucket: $BUCKET_NAME"


# Empty the bucket
echo "Emptying bucket: $BUCKET_NAME"
aws s3 rm "s3://$BUCKET_NAME" --recursive --profile "$AWS_PROFILE"

# Confirm emptying was successful
if [ $? -eq 0 ]; then
  echo "Bucket $BUCKET_NAME emptied successfully."
else
  echo "Error emptying bucket $BUCKET_NAME."
  exit 1
fi



# Usage
# chmod +x scripts/empty-cdk-bucket.sh
# scripts/empty-cdk-bucket.sh -p AWS_PROFILE -r AWS_REGION

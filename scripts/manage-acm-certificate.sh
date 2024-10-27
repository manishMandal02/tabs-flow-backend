#!/bin/bash



# Check if ACM certificate ARN exists in SSM
EXISTING_CERT_ARN=$(aws ssm get-parameter --name "$ACM_CERTIFICATE_ARN_SSM_PARAM_NAME" --query "Parameter.Value" --output text 2>/dev/null)

if [ -z "$EXISTING_CERT_ARN" ]; then
  # If SSM parameter is not found or is empty
  echo "SSM parameter not found or empty, checking ACM for existing certificate..."

  # Check if an ACM certificate exists for the domain
  EXISTING_CERT_ARN=$(aws acm list-certificates --query "CertificateSummaryList[?DomainName=='$API_DOMAIN_NAME'].CertificateArn" --output text)

  if [ -z "$EXISTING_CERT_ARN" ]; then
    # If no certificate exists in ACM
    echo "No ACM certificate found for $API_DOMAIN_NAME. Run the ACM stack to create the certificate."
    exit 1  # Exit with error since the certificate is missing
  else
    # If ACM certificate exists
    echo "ACM certificate found: $EXISTING_CERT_ARN"
    
    # Store the ACM certificate ARN in SSM Parameter Store
    echo "Storing the ACM certificate ARN in SSM Parameter Store..."
    aws ssm put-parameter --name "$ACM_CERTIFICATE_ARN_SSM_PARAM_NAME" --value "$EXISTING_CERT_ARN" --type String --overwrite
  fi
else
  # If the certificate was found in SSM
  echo "ACM certificate ARN found in SSM: $EXISTING_CERT_ARN"
fi

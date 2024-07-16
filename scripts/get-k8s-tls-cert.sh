#!/usr/bin/env bash

# Step 1: Define the Certificate resource name and namespace
CERT_NAME=$1
NAMESPACE=$2
# DEST_FOLDER defaults to ./certs
DEST_FOLDER="${3:-./certs}"

mkdir -p $DEST_FOLDER

# check that CERT_NAME and NAMESPACE are provided
if [ -z "$CERT_NAME" ] || [ -z "$NAMESPACE" ]; then
  echo "Usage: get-k8s-tls-cert.sh <CERT_NAME> <NAMESPACE> <DEST_FOLDER>"
  echo "DEST_FOLDER defaults to ./certs"
  exit 1
fi 

# Step 2: Get the Secret name from the Certificate
SECRET_NAME=$(kubectl get certificate $CERT_NAME -n $NAMESPACE -o jsonpath='{.spec.secretName}')

# Step 3: Retrieve the Secret and extract the certificate data
CERT_DATA=$(kubectl get secret $SECRET_NAME -n $NAMESPACE -o jsonpath='{.data.tls\.crt}')

# Step 4: Decode the certificate data from base64
echo $CERT_DATA | base64 --decode > $DEST_FOLDER/$CERT_NAME.crt

echo "TLS certificate saved to $DEST_FOLDER/$CERT_NAME.crt"
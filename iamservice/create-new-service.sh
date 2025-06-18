#!/bin/bash

# This script was tested on macOS only

# Function to display usage message and exit
show_usage() {
  echo "Usage: create-new-service.sh <service-name> <package> <output-dir>"
  echo ""
  echo "  e.g: ./create-new-service.sh myservice github.com/somehandler/myservice ../../myservice"
  exit 1
}

# let's read first argument (service-name)
SERVICE_NAME=$1
if [ -z "$SERVICE_NAME" ]; then
  show_usage
fi

# let's read the second argument (package)
PACKAGE=$2
if [ -z "$PACKAGE" ]; then
  show_usage
fi

# let's read the third argument (output-dir)
OUTPUT_DIR=$3
if [ -z "$OUTPUT_DIR" ]; then
  show_usage
fi

# Create the output directory if it does not exist
if [ ! -d "$OUTPUT_DIR" ]; then
  mkdir -p "$OUTPUT_DIR"
else
  echo "Directory $OUTPUT_DIR already exists"
  exit 1
fi
# Copy current directory to output directory excluding .git and .idea directories
rsync -a --exclude={.git,.idea} . "$OUTPUT_DIR"

find "$OUTPUT_DIR" -type f -exec sed -i '' "s|github.com/mobiletoly/gokatana-samples/iamservice|$PACKAGE|g" {} \;

# In file internal/core/app/configload.go replace substring "IAMSERVICE" in quotes with the service name in uppercase and with removed dashes and underscores
sed -i '' "s/\"IAMSERVICE\"/\"$(echo "$SERVICE_NAME" | tr '[:lower:]' '[:upper:]' | tr -d '-' | tr -d '_')\"/g" "$OUTPUT_DIR"/internal/core/app/configload.go

# remove "replace github.com/mobiletoly/gokatana => ../../gokatana" string
sed -i '' '/replace github\.com\/mobiletoly\/gokatana => \.\.\/\.\.\/gokatana/d' "$OUTPUT_DIR"/go.mod

rm -rf "$OUTPUT_DIR"/.git
rm "$OUTPUT_DIR"/README.md
rm "$OUTPUT_DIR"/create-new-service.sh

CURRENT_DIR=$(pwd)
cd "$OUTPUT_DIR" || exit 1
go mod tidy
cd "$CURRENT_DIR" || exit 1

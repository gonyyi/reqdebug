#!/bin/sh

source ~/.bashrc-get-build-no.sh

BUILD_NO=$(GetBuildNo "zBuild.txt")
BUILD_DATE=$(date '+%Y-%m%d-%H%M')

# Compile with build date and build version.
go build -ldflags "-X main.buildDate=${BUILD_DATE} -X main.buildNo=${BUILD_NO}" -o ./_reqtest
if [ "$?" != "0" ]; then
  exit 1
fi

# Run after compile only when "run" was given as a param
if [ "$1" = "run" ]; then
  ./_exec/sbi
fi

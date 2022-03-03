#!/bin/bash

if [[ -z ${TARGETPLATFORM} ]]; then
  echo "TARGETPLATFORM must be defined"
  exit 1
fi

for file in "$@"; do
  RESULT=$(upx -q "$file")
  echo "$RESULT"

  # verify the file arch is correct
  if ! echo "$RESULT" | grep "${TARGETPLATFORM}" > /dev/null; then
    echo "Error ${file} must be of arch ${TARGETPLATFORM}"
    exit 2
  fi

done


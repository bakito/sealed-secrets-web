#!/bin/bash
set -e

./runTestKubeseal.sh

./runTestCertificate.sh

./runTestDencode.sh

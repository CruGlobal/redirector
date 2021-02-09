#!/usr/bin/env bash

docker build -t 056154071827.dkr.ecr.us-east-1.amazonaws.com/redirector:$ENVIRONMENT-$BUILD_NUMBER .

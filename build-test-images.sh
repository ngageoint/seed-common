#!/usr/bin/env bash
cd "$(dirname "$0")"
pwd
docker build -t localhost:5000/my-job-0.1.0-seed:0.1.0 testdata/complete
docker tag localhost:5000/my-job-0.1.0-seed:0.1.0 localhost:5000/testorg/my-job-0.1.0-seed:0.1.0
docker login localhost:5000 -u testuser -p testpassword
docker push localhost:5000/my-job-0.1.0-seed:0.1.0
docker push localhost:5000/testorg/my-job-0.1.0-seed:0.1.0
[![Build Status](https://travis-ci.org/emilkey/miniosqs.svg?branch=master)](https://travis-ci.org/emilkey/miniosqs)
# MinioSQS

MinioSQS is a small utility that uses [MinIO](https://github.com/minio/minio) and [ElasticMQ](https://github.com/softwaremill/elasticmq) to enable local testing of systems that use [AWS S3 event notifications](https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html) or [AWS s3 Server Access Logging + CloudTrail + SQS](https://docs.aws.amazon.com/AmazonS3/latest/userguide/cloudtrail-logging-s3-info.html#cloudtrail-logging-s3-requests).

## Basic Instructions

1. Download and run ElasticMQ
2. Download and run MinioSQS
    a. For audit logs, use: `--msgtype AUDIT` (or without flag, AUDIT is default message type)
    b. For Event Notifications, use: `--msgtype EVENT`
3. Download and run MinIO
4. 
    a. To subscribe bucket event notifications; Create a bucket and configure it's settings to send event notifications via webhook to the MinioSQS endpoint

    b. To subscribe audit logs (server access logging); Open minio dashboard, go to general settings and set minisqs endpont as audit log webhook.


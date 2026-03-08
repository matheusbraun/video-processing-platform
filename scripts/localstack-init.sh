#!/bin/bash
# Cria os buckets S3 no LocalStack automaticamente ao iniciar
set -e

echo "Criando buckets S3 no LocalStack..."

aws --endpoint-url=http://localhost:4566 s3 mb s3://video-platform-uploads  --region us-east-1
aws --endpoint-url=http://localhost:4566 s3 mb s3://video-platform-processed --region us-east-1

echo "Buckets criados com sucesso:"
aws --endpoint-url=http://localhost:4566 s3 ls

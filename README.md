# catbox

Vulnerability scan and management serverless system for AWS ECR images with [Trivy](https://github.com/aquasecurity/trivy).

![](https://user-images.githubusercontent.com/605953/108437063-ea328000-728f-11eb-81eb-b444ec43d4a9.png)

*meow*

## Deploy

### Prerequisite

- aws-cdk >= 1.90.0

## Development

### Test

Docker image `amazon/dynamodb-local` is required to run for testing database I/O.

```bash
$ docker run -d -p 8000:8000 amazon/dynamodb-local
$ go test ./...
```

# narratorfiles

Read-only S3 file server with the special care for audio files.
Those can be listened to directly in the browser.
The rest of the files can be downloaded through the signed URL directly from the underlying S3 store.

### Docker

```
docker pull ghcr.io/projmayhem/narratorfiles:latest
```

### Configuration

The following environment variables are supported:

```
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_S3_ENDPOINT
AWS_REGION
AWS_S3_BUCKET
AWS_S3_PREFIX
```

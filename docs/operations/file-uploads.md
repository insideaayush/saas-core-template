# File Uploads

This template supports tenant-scoped file uploads via an adapter boundary.

## Local development (disk)

Defaults:

- `FILE_STORAGE_PROVIDER=disk`
- `FILE_STORAGE_DISK_PATH=./.data/uploads` (relative to `backend/` when running `make dev-api`)

Uploads use a direct API endpoint:

- `POST /api/v1/files/upload-url` → returns an upload URL
- `POST /api/v1/files/{id}/upload` → multipart upload (disk provider)
- `POST /api/v1/files/{id}/complete` → marks presigned uploads complete (S3/R2)

## Production (S3 / R2)

Set:

- `FILE_STORAGE_PROVIDER=s3`
- `S3_BUCKET=<bucket>`
- `S3_REGION=<region>` (R2 uses `auto`)
- `S3_ENDPOINT=<optional endpoint>` (required for R2)
- `S3_ACCESS_KEY_ID`
- `S3_SECRET_ACCESS_KEY`
- `S3_FORCE_PATH_STYLE=true` (usually required for R2)

Uploads use presigned PUT URLs, so the frontend uploads directly to object storage.

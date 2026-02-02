# File Storage Architecture

This document details the file storage configuration for the Loomio application.

## 1. Storage Backend Options

Loomio supports **6 storage backends** via ActiveStorage, configured in `config/storage.yml`.

### 1.1 Test (Disk)

**Purpose:** Automated testing only

```yaml
test:
  service: Disk
  root: <%= Rails.root.join("tmp/storage") %>
```

- **File:** `config/storage.yml:1-3`
- **Usage:** `config/environments/test.rb:40` - `config.active_storage.service = :test`

### 1.2 Local (Disk)

**Purpose:** Development and simple deployments

```yaml
local:
  service: Disk
  root: <%= Rails.root.join("storage") %>
```

- **File:** `config/storage.yml:5-7`
- **Usage:** Default for development (`config/environments/development.rb:35`)
- **Selection Logic:** `config/application.rb:53` - Used when `AWS_BUCKET` is not set and `ACTIVE_STORAGE_SERVICE` defaults to `:local`

### 1.3 Amazon S3

**Purpose:** Production cloud storage on AWS

```yaml
amazon:
  service: S3
  access_key_id: <%= ENV['AWS_ACCESS_KEY_ID'] %>
  secret_access_key: <%= ENV['AWS_SECRET_ACCESS_KEY'] %>
  bucket: <%= ENV['AWS_BUCKET'] %>
  region: <%= ENV['AWS_REGION'] %>
```

- **File:** `config/storage.yml:9-14`
- **Selection Logic:** `config/application.rb:50-51` - Automatically selected when `AWS_BUCKET` is present

### 1.4 DigitalOcean Spaces

**Purpose:** S3-compatible object storage on DigitalOcean

```yaml
digitalocean:
  service: S3
  endpoint: <%= ENV['DO_ENDPOINT'] %>
  access_key_id: <%= ENV['DO_ACCESS_KEY_ID'] %>
  secret_access_key: <%= ENV['DO_SECRET_ACCESS_KEY'] %>
  bucket: <%= ENV['DO_BUCKET'] %>
  region: ignored
```

- **File:** `config/storage.yml:16-22`
- **Note:** Region is set to `ignored` (DigitalOcean Spaces doesn't use AWS regions)

### 1.5 S3-Compatible

**Purpose:** Generic S3-compatible services (MinIO, Backblaze B2, Wasabi, etc.)

```yaml
s3_compatible:
  service: S3
  endpoint: <%= ENV.fetch('STORAGE_ENDPOINT', '') %>
  access_key_id: <%= ENV.fetch('STORAGE_ACCESS_KEY_ID', '') %>
  secret_access_key: <%= ENV.fetch('STORAGE_SECRET_ACCESS_KEY', '') %>
  region: <%= ENV.fetch('STORAGE_REGION', '') %>
  bucket: <%= ENV.fetch('STORAGE_BUCKET_NAME', '') %>
  force_path_style: <%= ENV.fetch('STORAGE_FORCE_PATH_STYLE', false) %>
  request_checksum_calculation: "when_required"
  response_checksum_validation: "when_required"
```

- **File:** `config/storage.yml:24-33`
- **Note:** Supports `force_path_style` for services that require path-style URLs

### 1.6 Google Cloud Storage

**Purpose:** Production cloud storage on GCP

```yaml
google:
  service: GCS
  credentials: <%= ENV.fetch('GCS_CREDENTIALS', '') %>
  project: <%= ENV.fetch('GCS_PROJECT', '') %>
  bucket: <%= ENV.fetch('GCS_BUCKET', '') %>
```

- **File:** `config/storage.yml:35-39`

**Confidence: HIGH** - All backends are explicitly defined in `config/storage.yml`

---

## 2. Environment Variable Mapping

### 2.1 Storage Backend Selection

| Variable | Default | Description | File Reference |
|----------|---------|-------------|----------------|
| `AWS_BUCKET` | - | If set, forces `amazon` backend | `config/application.rb:50-51` |
| `ACTIVE_STORAGE_SERVICE` | `local` | Backend name when `AWS_BUCKET` not set | `config/application.rb:53` |

**Selection Logic** (from `config/application.rb:50-54`):
```ruby
if ENV['AWS_BUCKET']
  config.active_storage.service = :amazon
else
  config.active_storage.service = ENV.fetch('ACTIVE_STORAGE_SERVICE', :local)
end
```

### 2.2 Amazon S3

| Variable | Required | Description |
|----------|----------|-------------|
| `AWS_ACCESS_KEY_ID` | Yes | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | Yes | AWS secret key |
| `AWS_BUCKET` | Yes | S3 bucket name |
| `AWS_REGION` | Yes | AWS region (e.g., `us-east-1`) |

### 2.3 DigitalOcean Spaces

| Variable | Required | Description |
|----------|----------|-------------|
| `DO_ENDPOINT` | Yes | Spaces endpoint URL (e.g., `https://nyc3.digitaloceanspaces.com`) |
| `DO_ACCESS_KEY_ID` | Yes | Spaces access key |
| `DO_SECRET_ACCESS_KEY` | Yes | Spaces secret key |
| `DO_BUCKET` | Yes | Spaces bucket name |

### 2.4 S3-Compatible

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_ENDPOINT` | `''` | Service endpoint URL |
| `STORAGE_ACCESS_KEY_ID` | `''` | Access key |
| `STORAGE_SECRET_ACCESS_KEY` | `''` | Secret key |
| `STORAGE_REGION` | `''` | Region identifier |
| `STORAGE_BUCKET_NAME` | `''` | Bucket name |
| `STORAGE_FORCE_PATH_STYLE` | `false` | Use path-style URLs instead of virtual-hosted |

### 2.5 Google Cloud Storage

| Variable | Default | Description |
|----------|---------|-------------|
| `GCS_CREDENTIALS` | `''` | JSON credentials string or file path |
| `GCS_PROJECT` | `''` | GCP project ID |
| `GCS_BUCKET` | `''` | GCS bucket name |

**Confidence: HIGH** - All variables extracted directly from `config/storage.yml`

---

## 3. File Size Limits

### 3.1 Hard Limits

**Finding: NO EXPLICIT FILE SIZE LIMITS FOUND**

After comprehensive search of the codebase:
- No `validates :file, size: ...` patterns
- No `ActiveStorage::Blob` size validations
- No frontend JavaScript file size checks
- No `MAX_ATTACHMENT_BYTES` environment variable implementation

### 3.2 Indirect Limits

| Limit Type | Value | File Reference |
|------------|-------|----------------|
| Rate limit on uploads | 20 requests/hour/IP | `config/initializers/rack_attack.rb:39` |
| Message length limit | 100,000 characters | `app/extras/app_config.rb:146` - `max_message_length` |
| User avatar constant | 100 MB (unused?) | `app/models/user.rb:23` - `MAX_AVATAR_IMAGE_SIZE_CONST = 100.megabytes` |

**Note on `MAX_AVATAR_IMAGE_SIZE_CONST`:** This constant is defined but **not used anywhere** in the codebase for validation. It appears to be a remnant or placeholder.

### 3.3 External Limits

File size limits are enforced at the infrastructure level:
- **Nginx/web server:** `client_max_body_size` (not configured in Rails)
- **S3/GCS:** Provider-specific limits (typically 5GB for single PUT)
- **ActiveStorage Direct Upload:** Streams directly to storage, bypassing Rails

**Confidence: HIGH** - Comprehensive grep searches confirm no application-level file size validation

---

## 4. Image Processing Configuration

### 4.1 Processing Library

| Setting | Value | File Reference |
|---------|-------|----------------|
| Processor | `vips` | `config/application.rb:48` |
| Gem | `image_processing ~> 1.14` | `Gemfile:31` |
| Vips binding | `ruby-vips` | `Gemfile:32` |

**Docker dependencies** (from `Dockerfile:42-44`):
- `libvips` - Image processing library
- `ffmpeg` - Audio/video processing
- `imagemagick` - Fallback image processing

### 4.2 Image Variants

#### Default Preview Options

**Defined in `app/models/concerns/has_rich_text.rb:2-8`:**
```ruby
PREVIEW_OPTIONS = {
  resize_to_limit: [1280, 1280],
  saver: {
    quality: 85,
    strip: true
  }
}
```

| Option | Value | Description |
|--------|-------|-------------|
| `resize_to_limit` | `[1280, 1280]` | Max dimensions (preserves aspect ratio) |
| `quality` | `85` | JPEG/WebP compression quality |
| `strip` | `true` | Remove EXIF metadata |

#### User Avatar Variants

**Defined in `app/models/concerns/has_avatar.rb:67-69`:**
```ruby
uploaded_avatar.representation(resize_to_limit: [size, size], saver: {quality: 80, strip: true})
```

| Usage | Size | Quality |
|-------|------|---------|
| Thumbnail | 128px | 80 |
| Default | 512px | 80 |

#### Group Logo Variants

**Defined in `app/models/group.rb:170-172`:**
```ruby
logo.representation(resize_to_limit: [size, size], saver: {quality: 80, strip: true})
```

#### Group Cover Photo Variants

**Defined in `app/models/group.rb:186-188`:**
```ruby
cover_photo.representation(HasRichText::PREVIEW_OPTIONS.merge(resize_to_limit: [size*4, size]))
```

| Default size param | Dimensions | Aspect ratio |
|--------------------|------------|--------------|
| 512 | 2048x512 | 4:1 |
| 256 | 1024x256 | 4:1 |

### 4.3 Variant Tracking

**Enabled in `config/initializers/new_framework_defaults_6_1.rb:13`:**
```ruby
Rails.application.config.active_storage.track_variants = true
```

This stores variant records in the database for efficient caching.

### 4.4 Content Types Allowed Inline

**Defined in `config/application.rb:108`:**
```ruby
config.active_storage.content_types_allowed_inline = %w(
  audio/webm
  video/webm
  image/png
  image/gif
  image/jpeg
  image/tiff
  image/vnd.adobe.photoshop
  image/vnd.microsoft.icon
  application/pdf
)
```

Other content types will be served with `Content-Disposition: attachment`.

### 4.5 Frontend Accepted Types

**Profile avatar** (`vue/src/components/profile/change_picture_form.vue:95`):
```
accept="image/png, image/jpeg, image/webp"
```

**Group cover/logo** (`vue/src/components/group/form.vue:231-232`):
```
accept="image/png, image/jpeg, image/webp"
```

**Confidence: HIGH** - Image processing configuration is well-documented in code

---

## 5. Models with ActiveStorage Attachments

| Model | Attachment | Type | File Reference |
|-------|------------|------|----------------|
| User | `uploaded_avatar` | `has_one_attached` | `app/models/user.rb:46` |
| Group | `cover_photo` | `has_one_attached` | `app/models/group.rb:129` |
| Group | `logo` | `has_one_attached` | `app/models/group.rb:130` |
| Document | `file` | `has_one_attached` | `app/models/document.rb:11` |
| ReceivedEmail | `attachments` | `has_many_attached` | `app/models/received_email.rb:2` |
| HasRichText concern | `files` | `has_many_attached` | `app/models/concerns/has_rich_text.rb:66` |
| HasRichText concern | `image_files` | `has_many_attached` | `app/models/concerns/has_rich_text.rb:67` |

**Models using HasRichText concern:**
- Discussion
- Poll
- Comment
- Outcome
- User (short_bio)
- Group (description)
- And others with rich text fields

**Confidence: HIGH** - All attachments discovered via grep search

---

## 6. Upload Flow

### 6.1 Direct Upload Endpoint

**Controller:** `app/controllers/direct_uploads_controller.rb`

```ruby
class DirectUploadsController < ActiveStorage::DirectUploadsController
  # Returns:
  # - signed_id
  # - direct_upload.url (presigned URL)
  # - direct_upload.headers
  # - download_url
  # - preview_url (if representable)
end
```

**Route:** `POST /direct_uploads` (`config/routes.rb:339`)

### 6.2 Rate Limiting

**From `config/initializers/rack_attack.rb:39`:**
```ruby
'/rails/active_storage/direct_uploads' => 20  # per hour per IP
```

### 6.3 Frontend Upload

**File:** `vue/src/shared/services/file_uploader.js`

Uses ActiveStorage JavaScript's `DirectUpload` class to stream files directly to storage backend.

**Confidence: HIGH** - Upload flow is standard ActiveStorage with minor customizations

---

## 7. Summary

| Aspect | Finding | Confidence |
|--------|---------|------------|
| Storage backends | 6 options (2 disk, 4 cloud) | HIGH |
| Backend selection | `AWS_BUCKET` presence OR `ACTIVE_STORAGE_SERVICE` env | HIGH |
| File size limits | **NONE** at application level | HIGH |
| Image processor | vips with 1280x1280 default resize | HIGH |
| Variant tracking | Enabled | HIGH |
| Rate limiting | 20 uploads/hour/IP | HIGH |

### Key Observations

1. **No File Size Validation:** The application does not enforce file size limits. This is delegated to infrastructure (web server, cloud storage limits).

2. **Image Processing:** All image variants use vips with quality 80-85 and metadata stripping for privacy/size optimization.

3. **Backend Priority:** If `AWS_BUCKET` is set, Amazon S3 is always used regardless of `ACTIVE_STORAGE_SERVICE`.

4. **Direct Upload:** Files bypass Rails and stream directly to storage via presigned URLs.

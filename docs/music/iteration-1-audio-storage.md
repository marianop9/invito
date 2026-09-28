# Iteration 1: Audio Storage & Media Pipeline Compatibility

## 1. Goal & Rationale
Extend Invito's existing media pipeline to support audio file uploads (MP3, M4A/AAC, OGG, WAV, WebM) alongside images, while maintaining 100% backward compatibility with the `MediaStorage` interface (`pkg/storage/media.go`) and ensuring that byte-range requests (`HTTP 206 Partial Content`) work out-of-the-box for instant browser streaming.

---

## 2. Interface Compatibility Analysis

The `MediaStorage` interface declared in [`pkg/storage/media.go`](file:///home/nano/projects/invitation/pkg/storage/media.go) is defined as:

```go
type MediaStorage interface {
    Save(ctx context.Context, input SaveMediaInput) (*MediaInfo, error)
    Exists(ctx context.Context, filename string) (bool, error)
    Delete(ctx context.Context, filename string) error
    PublicURL(filename string) string
}
```

### Compatibility Guarantee
* **No changes to method signatures**: `Save`, `Exists`, `Delete`, and `PublicURL` remain unchanged.
* `SaveMediaInput` already provides generic fields: `OriginalName`, `ContentType`, `Reader`, `Size`, and `Slug`.
* The `MediaInfo` output struct already holds `URL`, `Filename`, `ContentType`, `Size`, and `CreatedAt`.
* Therefore, cloud implementations (future AWS S3 or Cloudflare R2 providers) can implement the exact same interface for both images and audio assets without branching.

---

## 3. Storage Layer Updates (`pkg/storage/local_media.go`)

### 3.1 MIME Sniffing Extension
Currently, `sniffMIMEType(header []byte)` in [`pkg/storage/local_media.go`](file:///home/nano/projects/invitation/pkg/storage/local_media.go) only verifies image magic bytes. It will be extended to detect the following audio signatures:

| Format | Content-Type | Extension | Magic Bytes / Header Inspection |
| :--- | :--- | :--- | :--- |
| **MP3 (with ID3v2)** | `audio/mpeg` | `.mp3` | Header starts with `ID3` (`\x49\x44\x33`) |
| **MP3 (raw frame)** | `audio/mpeg` | `.mp3` | Sync bytes `\xff\xfb`, `\xff\xfa`, `\xff\xf3`, or `\xff\xf2` |
| **M4A / MP4 Audio** | `audio/mp4` | `.m4a` | Bytes 4..8 equal `ftyp` with subtype `M4A `, `mp42`, or `isom` |
| **OGG Audio** | `audio/ogg` | `.ogg` | Header starts with `OggS` (`\x4f\x67\x67\x53`) |
| **WAV Audio** | `audio/wav` | `.wav` | Bytes 0..4 equal `RIFF` and bytes 8..12 equal `WAVE` |
| **WebM Audio** | `audio/webm` | `.weba` | Header starts with Matroska/EBML `\x1a\x45\xdf\xa3` |

### 3.2 Error and Limit Adjustments
* Update `ErrInvalidMediaType` error message:
  ```go
  ErrInvalidMediaType = errors.New("unsupported media type: allowed formats are JPEG, PNG, WebP, GIF images and MP3, M4A, OGG, WAV audio")
  ```
* Leave `DefaultMaxUploadSize` as is (`10 * 1024 * 1024` (10MB)).

---

## 4. Server Route & Handler Updates (`pkg/server/server.go`)

### 4.1 Flexible Form Field Parsing in `handleAPIUpload`
Currently, `handleAPIUpload` in [`pkg/server/server.go`](file:///home/nano/projects/invitation/pkg/server/server.go) explicitly looks for `r.FormFile("image")`.

To support audio files while maintaining backward compatibility with existing image uploaders:
```go
// Check "audio", "file", and fallback to "image"
file, header, err := r.FormFile("audio")
if err != nil {
    file, header, err = r.FormFile("file")
}
if err != nil {
    file, header, err = r.FormFile("image")
}
if err != nil {
    // Return friendly error indicating missing multipart file field
}
```

### 4.2 Byte-Range Audio Streaming via `handleServeUpload`
The existing `handleServeUpload` handler executes:
```go
http.ServeFile(w, r, fullPath)
```
`http.ServeFile` provides full RFC 7233 byte-range request support (`HTTP 206 Partial Content`), handling `Range: bytes=start-end` headers automatically.
* Mobile browsers request audio in chunks as playback progresses rather than downloading the entire file upfront.
* The response sets correct headers: `Accept-Ranges: bytes`, `Content-Length`, `Content-Range`, and `Content-Type`.

---

## 5. Verification & Testing Checklist

- [ ] **Unit Tests (`local_media_test.go`)**:
  - Upload valid MP3 (with ID3v2 tag and raw sync).
  - Upload valid M4A (`ftypM4A`), OGG, and WAV files.
  - Verify generated URLs point to `/uploads/{slug}-{timestamp}-{rand}.{ext}`.
  - Verify rejection of unapproved formats (e.g. `.exe`, `.pdf`, `.mp4` video).
  - Verify size enforcement (> 10MB returns `ErrFileTooLarge`).
- [ ] **Integration Tests (`server_test.go`)**:
  - `POST /api/upload` with multipart field `audio`.
  - `POST /api/upload` with legacy multipart field `image`.
  - `GET /uploads/{audio-file}` verifying `HTTP 200` and `HTTP 206` when sending a `Range: bytes=0-1024` request header.

# Lesson 4: Performance: Brotli Compression, Caching, and Validation

## Learning Objectives

- Enable Brotli compression for eligible responses
- Configure Cache-Control headers with path-based overrides
- Enforce request body limits and content-type requirements

## Key Concepts

- **Brotli Compression**: The middleware buffers the response body, checks if the client accepts Brotli (`Accept-Encoding: br`), and compresses only if the body exceeds `MinLength` and the content type is compressible. Sets `Content-Encoding: br` and `Vary: Accept-Encoding` on compressed responses.
- **Cache Control**: The cache middleware sets a `Cache-Control` header per response. Path overrides allow different caching strategies (e.g., `public, max-age=31536000` for `/static/` and `no-store` for API routes). `StaticAssetsConfig()` provides a convenience preset.
- **Input Validation**: Uses `http.MaxBytesReader` to enforce body size limits. Rejects POST/PUT/PATCH requests missing an allowed Content-Type with 415 Unsupported Media Type. Helper functions `SanitizeString()` and `SanitizeHeader()` are available for manual input cleaning.

## Code Walkthrough

### Source: `pkg/brotli/brotli.go`

The `brotliWriter` buffers all output. On `finish()`, it checks content type against `CompressibleTypes` and body size against `MinLength`, then either compresses with `brotli.NewWriterLevel` or passes through raw bytes.

### Source: `pkg/cache/cache.go`

`buildDirective()` assembles the Cache-Control string from config fields (`Public`, `Private`, `MaxAge`, `NoCache`, `NoStore`, `MustRevalidate`). Path overrides are checked with `strings.HasPrefix` on each request.

### Source: `pkg/validation/validation.go`

Content-type matching normalizes the header value (strips charset parameters) before comparison. `SanitizeString` removes null bytes and trims whitespace. `SanitizeHeader` removes `\r` and `\n` to prevent header injection.

## Practice Exercise

1. Configure Brotli compression with `MinLength: 100`. Send a small response (50 bytes) and a large response (500 bytes). Verify only the large one is compressed.
2. Set up cache middleware with `no-store` default and `public, max-age=3600` for `/assets/` paths. Verify headers on both path types.
3. Configure validation requiring `application/json` and a 1 MB body limit. Send a 2 MB POST and verify the server rejects it.

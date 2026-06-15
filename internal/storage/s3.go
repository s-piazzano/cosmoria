package storage

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type S3Client struct {
	endpoint  string
	accessKey string
	secretKey string
	bucket    string
	region    string
	useSSL    bool
	client    *http.Client
}

func NewS3Client(endpoint, accessKey, secretKey, bucket, region string, useSSL bool) *S3Client {
	return &S3Client{
		endpoint:  endpoint,
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		region:    region,
		useSSL:    useSSL,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *S3Client) PutObject(key string, reader io.Reader, size int64, contentType string) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("s3: read body: %w", err)
	}
	payloadHash := sha256Hex(body)

	now := time.Now().UTC()
	dateStr := now.Format("20060102T150405Z")
	dateShort := now.Format("20060102")

	encKey := encodeKey(key)
	canonicalURI := "/" + s.bucket + "/" + encKey
	signedHeaders := "host;x-amz-content-sha256"
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\n", s.host(), payloadHash)
	canonicalRequest := fmt.Sprintf("PUT\n%s\n\n%s\n%s\n%s", canonicalURI, canonicalHeaders, signedHeaders, payloadHash)

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateShort, s.region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", dateStr, credentialScope, sha256Hex([]byte(canonicalRequest)))

	signingKey := s.signingKey(dateShort)
	signature := hex.EncodeToString(s.hmacSHA256(signingKey, []byte(stringToSign)))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s,SignedHeaders=%s,Signature=%s",
		s.accessKey, credentialScope, signedHeaders, signature)

	u := fmt.Sprintf("%s/%s/%s", s.baseURL(), s.bucket, encKey)
	req, err := http.NewRequest("PUT", u, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("s3: create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", size))
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("Authorization", authHeader)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("s3: put %s: %w", key, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("s3: put %s: %d %s", key, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func (s *S3Client) PresignedGetURL(key string, expiry time.Duration) (string, error) {
	now := time.Now().UTC()
	dateStr := now.Format("20060102T150405Z")
	dateShort := now.Format("20060102")

	expires := int(expiry.Seconds())
	if expires <= 0 {
		expires = 3600
	}

	encKey := encodeKey(key)
	canonicalURI := "/" + s.bucket + "/" + encKey

	q := url.Values{}
	q.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	q.Set("X-Amz-Credential", fmt.Sprintf("%s/%s/%s/s3/aws4_request", s.accessKey, dateShort, s.region))
	q.Set("X-Amz-Date", dateStr)
	q.Set("X-Amz-Expires", fmt.Sprintf("%d", expires))
	q.Set("X-Amz-SignedHeaders", "host")

	canonicalQuery := q.Encode()
	payloadHash := "UNSIGNED-PAYLOAD"
	signedHeaders := "host"
	canonicalHeaders := fmt.Sprintf("host:%s\n", s.host())
	canonicalRequest := fmt.Sprintf("GET\n%s\n%s\n%s\n%s\n%s", canonicalURI, canonicalQuery, canonicalHeaders, signedHeaders, payloadHash)

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateShort, s.region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", dateStr, credentialScope, sha256Hex([]byte(canonicalRequest)))

	signingKey := s.signingKey(dateShort)
	signature := hex.EncodeToString(s.hmacSHA256(signingKey, []byte(stringToSign)))

	u := fmt.Sprintf("%s/%s/%s?%s&X-Amz-Signature=%s", s.baseURL(), s.bucket, encKey, canonicalQuery, signature)
	return u, nil
}

func (s *S3Client) DeleteObject(key string) error {
	payloadHash := sha256Hex(nil)

	now := time.Now().UTC()
	dateStr := now.Format("20060102T150405Z")
	dateShort := now.Format("20060102")

	encKey := encodeKey(key)
	canonicalURI := "/" + s.bucket + "/" + encKey
	signedHeaders := "host;x-amz-content-sha256"
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\n", s.host(), payloadHash)
	canonicalRequest := fmt.Sprintf("DELETE\n%s\n\n%s\n%s\n%s", canonicalURI, canonicalHeaders, signedHeaders, payloadHash)

	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateShort, s.region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s", dateStr, credentialScope, sha256Hex([]byte(canonicalRequest)))

	signingKey := s.signingKey(dateShort)
	signature := hex.EncodeToString(s.hmacSHA256(signingKey, []byte(stringToSign)))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s,SignedHeaders=%s,Signature=%s",
		s.accessKey, credentialScope, signedHeaders, signature)

	u := fmt.Sprintf("%s/%s/%s", s.baseURL(), s.bucket, encKey)
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		return fmt.Errorf("s3: create delete request: %w", err)
	}
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("Authorization", authHeader)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("s3: delete %s: %w", key, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("s3: delete %s: %d %s", key, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

func (s *S3Client) hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func (s *S3Client) signingKey(dateShort string) []byte {
	kSecret := []byte("AWS4" + s.secretKey)
	kDate := s.hmacSHA256(kSecret, []byte(dateShort))
	kRegion := s.hmacSHA256(kDate, []byte(s.region))
	kService := s.hmacSHA256(kRegion, []byte("s3"))
	return s.hmacSHA256(kService, []byte("aws4_request"))
}

func (s *S3Client) baseURL() string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, s.endpoint)
}

func (s *S3Client) host() string {
	return s.endpoint
}

func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func encodeKey(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

func generateFileID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func sanitizeFilename(name string) string {
	var clean []byte
	for _, c := range []byte(name) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '_' {
			clean = append(clean, c)
		} else {
			clean = append(clean, '_')
		}
	}
	return string(clean)
}

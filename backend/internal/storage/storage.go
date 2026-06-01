// Package storage wraps Supabase Storage REST API: upload, download, signed URL.
package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SupabaseStorage is a thin client over Supabase Storage REST endpoints.
// Uses the service-role key (Bearer + apikey headers) so it can bypass RLS on storage.objects.
type SupabaseStorage struct {
	baseURL    string // e.g. https://<ref>.supabase.co
	serviceKey string
	hc         *http.Client
}

func NewSupabaseStorage(supabaseURL, serviceKey string) *SupabaseStorage {
	return &SupabaseStorage{
		baseURL:    supabaseURL,
		serviceKey: serviceKey,
		hc:         &http.Client{Timeout: 60 * time.Second},
	}
}

// UploadObject puts content into bucket at path, returning the stored path.
func (s *SupabaseStorage) UploadObject(bucket, path string, content []byte, contentType string) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, bucket, path)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(content))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")
	resp, err := s.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s — %s", resp.Status, string(body))
	}
	return path, nil
}

// DownloadObject fetches the raw bytes of an object.
func (s *SupabaseStorage) DownloadObject(bucket, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, bucket, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("apikey", s.serviceKey)
	resp, err := s.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed: %s — %s", resp.Status, string(body))
	}
	return io.ReadAll(resp.Body)
}

// SignedURL creates a time-limited URL for direct browser download.
func (s *SupabaseStorage) SignedURL(bucket, path string, ttlSeconds int) (string, error) {
	url := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", s.baseURL, bucket, path)
	body, _ := json.Marshal(map[string]any{"expiresIn": ttlSeconds})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("signed url failed: %s — %s", resp.Status, string(b))
	}
	var out struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return s.baseURL + "/storage/v1" + out.SignedURL, nil
}

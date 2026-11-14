package skill

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	archiveName = "agentmesh-skill.zip"
	releasesURL = "https://github.com/alanjm0532/agentmesh/releases/download/"
	releasesAPI = "https://api.github.com/repos/alanjm0532/agentmesh/releases"
)

var validTag = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

type release struct {
	version   string
	archive   []byte
	checksums []byte
}

type githubRelease struct {
	Tag    string `json:"tag_name"`
	Assets []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"assets"`
}

func fetchRelease(ctx context.Context, client *http.Client, endpoint string) (*githubRelease, error) {
	data, err := download(ctx, client, endpoint, 1<<20, "application/vnd.github+json")
	if err != nil {
		return nil, err
	}
	var metadata githubRelease
	if err = json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	if !validTag.MatchString(metadata.Tag) || metadata.Tag == "latest" {
		return nil, fmt.Errorf("invalid release tag %q", metadata.Tag)
	}
	return &metadata, nil
}

func downloadRelease(ctx context.Context, client *http.Client, version string) (release, error) {
	var metadata *githubRelease
	if version == "latest" {
		var err error
		metadata, err = fetchRelease(ctx, client, releasesAPI+"/latest")
		if err != nil {
			return release{}, err
		}
		version = metadata.Tag
	}
	asset := func(name string, limit int64) ([]byte, error) {
		data, directErr := download(ctx, client, releasesURL+version+"/"+name, limit, "")
		if directErr == nil {
			return data, nil
		}
		if ctx.Err() != nil {
			return nil, directErr
		}
		// Cache metadata so both assets use the same release and one API lookup.
		fallback := func() ([]byte, error) {
			if metadata == nil {
				var err error
				metadata, err = fetchRelease(ctx, client, releasesAPI+"/tags/"+version)
				if err != nil {
					return nil, err
				}
			}
			if metadata.Tag != version {
				return nil, fmt.Errorf("release tag mismatch: requested %s, got %s", version, metadata.Tag)
			}
			for _, candidate := range metadata.Assets {
				if candidate.Name == name && candidate.ID > 0 {
					return download(ctx, client, fmt.Sprintf("%s/assets/%d", releasesAPI, candidate.ID), limit, "application/octet-stream")
				}
			}
			return nil, fmt.Errorf("release %s has no asset %s", version, name)
		}
		data, apiErr := fallback()
		if apiErr != nil {
			return nil, errors.Join(directErr, fmt.Errorf("GitHub API fallback: %w", apiErr))
		}
		return data, nil
	}
	archive, err := asset(archiveName, 10<<20)
	if err != nil {
		return release{}, err
	}
	checksums, err := asset("SHA256SUMS", 1<<20)
	if err != nil {
		return release{}, err
	}
	return release{version: version, archive: archive, checksums: checksums}, nil
}

func download(ctx context.Context, client *http.Client, url string, limit int64, accept string) (data []byte, resultErr error) {
	// Bound each attempt separately so a stalled direct download leaves time for fallback.
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "agentmesh-cli")
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close download response: %w", closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: HTTP %s", url, resp.Status)
	}
	return readLimited(resp.Body, limit)
}

func readLimited(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("download or extracted content exceeds %d bytes", limit)
	}
	return data, nil
}

func verifyChecksum(archive, manifest []byte) error {
	var expected string
	for line := range strings.SplitSeq(string(manifest), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == archiveName {
			if expected != "" {
				return fmt.Errorf("duplicate checksum for %s", archiveName)
			}
			expected = fields[0]
		}
	}
	digest, err := hex.DecodeString(expected)
	if err != nil || len(digest) != sha256.Size {
		return fmt.Errorf("missing or invalid checksum for %s", archiveName)
	}
	actual := sha256.Sum256(archive)
	if !bytes.Equal(digest, actual[:]) {
		return fmt.Errorf("checksum mismatch for %s", archiveName)
	}
	return nil
}

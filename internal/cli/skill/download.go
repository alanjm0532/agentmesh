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
)

const (
	archiveName = "agentmesh-skill.zip"
	releasesURL = "https://github.com/alanjm0532/agentmesh/releases/download/"
	latestURL   = "https://api.github.com/repos/alanjm0532/agentmesh/releases/latest"
)

var validTag = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

type release struct {
	version   string
	archive   []byte
	checksums []byte
}

func downloadRelease(ctx context.Context, client *http.Client, version string) (release, error) {
	if version == "latest" {
		data, err := download(ctx, client, latestURL, 1<<20)
		if err != nil {
			return release{}, err
		}
		var latest struct {
			Tag string `json:"tag_name"`
		}
		if err = json.Unmarshal(data, &latest); err != nil {
			return release{}, fmt.Errorf("decode latest release: %w", err)
		}
		if !validTag.MatchString(latest.Tag) || latest.Tag == "latest" {
			return release{}, fmt.Errorf("invalid latest release tag %q", latest.Tag)
		}
		version = latest.Tag
	}
	baseURL := releasesURL + version + "/"
	archive, err := download(ctx, client, baseURL+archiveName, 10<<20)
	if err != nil {
		return release{}, err
	}
	checksums, err := download(ctx, client, baseURL+"SHA256SUMS", 1<<20)
	if err != nil {
		return release{}, err
	}
	return release{version: version, archive: archive, checksums: checksums}, nil
}

func download(ctx context.Context, client *http.Client, url string, limit int64) (data []byte, resultErr error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "agentmesh-cli")
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

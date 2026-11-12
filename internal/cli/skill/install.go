package skill

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Result struct {
	Version string
	Path    string
}

func Install(ctx context.Context, version string) (Result, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Result{}, fmt.Errorf("resolve user home directory: %w", err)
	}
	return install(ctx, &http.Client{Timeout: 2 * time.Minute}, version, filepath.Join(homeDir, ".agents", "skills", "agentmesh"))
}

func install(ctx context.Context, client *http.Client, version, destination string) (Result, error) {
	if !validTag.MatchString(version) {
		return Result{}, fmt.Errorf("invalid skill version %q", version)
	}
	downloaded, err := downloadRelease(ctx, client, version)
	if err != nil {
		return Result{}, err
	}
	if err = verifyChecksum(downloaded.archive, downloaded.checksums); err != nil {
		return Result{}, err
	}
	if err = installArchive(downloaded.archive, destination); err != nil {
		return Result{}, err
	}
	return Result{Version: downloaded.version, Path: destination}, nil
}

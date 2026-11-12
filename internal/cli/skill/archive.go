package skill

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func installArchive(data []byte, destination string) (resultErr error) {
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return fmt.Errorf("create skills directory: %w", err)
	}
	stage, err := os.MkdirTemp(parent, ".agentmesh-install-*")
	if err != nil {
		return fmt.Errorf("create skill staging directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(stage); removeErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("remove skill staging directory: %w", removeErr))
		}
	}()
	if err = extractArchive(data, stage); err != nil {
		return err
	}
	info, err := os.Lstat(destination)
	if errors.Is(err, os.ErrNotExist) {
		return os.Rename(stage, destination)
	}
	if err != nil {
		return fmt.Errorf("inspect existing skill: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("existing skill path is not a directory: %s", destination)
	}
	backup, err := os.MkdirTemp(parent, ".agentmesh-backup-*")
	if err != nil {
		return fmt.Errorf("create skill backup directory: %w", err)
	}
	keepBackup := false
	defer func() {
		if !keepBackup {
			if removeErr := os.RemoveAll(backup); removeErr != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("remove skill backup directory: %w", removeErr))
			}
		}
	}()
	backupPath := filepath.Join(backup, "agentmesh")
	if err = os.Rename(destination, backupPath); err != nil {
		return fmt.Errorf("back up existing skill: %w", err)
	}
	if err = os.Rename(stage, destination); err != nil {
		if restoreErr := os.Rename(backupPath, destination); restoreErr != nil {
			keepBackup = true
			return errors.Join(err, fmt.Errorf("restore skill backup at %s: %w", backupPath, restoreErr))
		}
		return fmt.Errorf("install skill directory: %w", err)
	}
	return nil
}

func extractArchive(data []byte, destination string) error {
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open skill archive: %w", err)
	}
	if len(archive.File) > 256 {
		return fmt.Errorf("skill archive contains too many entries")
	}
	remaining := int64(20 << 20)
	seen := make(map[string]bool)
	hasSkill := false
	for _, entry := range archive.File {
		name := strings.TrimSuffix(entry.Name, "/")
		if name == "." || !fs.ValidPath(name) || strings.ContainsAny(name, `\:`) {
			return fmt.Errorf("invalid skill archive path %q", entry.Name)
		}
		key := strings.ToLower(name)
		if seen[key] {
			return fmt.Errorf("duplicate skill archive path %q", entry.Name)
		}
		seen[key] = true
		path := filepath.Join(destination, filepath.FromSlash(name))
		if entry.FileInfo().IsDir() {
			if err = os.MkdirAll(path, 0o700); err != nil {
				return err
			}
			continue
		}
		if !entry.Mode().IsRegular() {
			return fmt.Errorf("unsupported skill archive entry %q", entry.Name)
		}
		content, readErr := readArchiveFile(entry, remaining)
		if readErr != nil {
			return fmt.Errorf("extract %s: %w", entry.Name, readErr)
		}
		remaining -= int64(len(content))
		if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
		if err = os.WriteFile(path, content, 0o600); err != nil {
			return err
		}
		if name == "SKILL.md" && len(bytes.TrimSpace(content)) > 0 {
			hasSkill = true
		}
	}
	if !hasSkill {
		return fmt.Errorf("skill archive must contain a nonempty SKILL.md at its root")
	}
	return nil
}

func readArchiveFile(entry *zip.File, limit int64) (data []byte, resultErr error) {
	reader, err := entry.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		resultErr = errors.Join(resultErr, reader.Close())
	}()
	return readLimited(reader, limit)
}

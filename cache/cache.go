package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func GenerateCacheKey(key string, filePatterns []string) (string, error) {
	if len(filePatterns) == 0 {
		return key, nil
	}

	files, err := MatchPatterns(filePatterns)
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	fmt.Println(files)

	h := sha256.New()
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return "", err
		}
		defer f.Close()

		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
	}

	if key == "" {
		return hex.EncodeToString(h.Sum(nil)), nil
	}

	return fmt.Sprintf("%s-%s", key, hex.EncodeToString(h.Sum(nil))), nil

}

func MatchPatterns(filePatterns []string) ([]string, error) {
	matches := []string{}
	for _, pattern := range filePatterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, err
		}

		matches = append(matches, files...)
	}

	return matches, nil
}

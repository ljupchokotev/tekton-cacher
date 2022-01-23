package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"log"

	"github.com/ingka-group-digital/es-tekton-cacher/cache"
	"github.com/ingka-group-digital/es-tekton-cacher/gcs"
	"github.com/ingka-group-digital/es-tekton-cacher/tar"
	"github.com/spf13/pflag"
)

var filePatterns []string
var cacheKeyFiles []string
var cacheKey string

var gcsFlags struct {
	Bucket     string
	PathPrefix string
}

func init() {
	pflag.StringSliceVar(&filePatterns, "file-patterns", []string{}, "")
	pflag.StringVar(&cacheKey, "cache-key", "", "")
	pflag.StringSliceVar(&cacheKeyFiles, "cache-key-files", []string{}, "")

	pflag.StringVar(&gcsFlags.Bucket, "bucket", "", "")
}

func main() {
	pflag.Parse()

	if len(filePatterns) == 0 {
		log.Fatal("--file-patterns is required")
	}
	if cacheKey == "" && len(cacheKeyFiles) == 0 {
		log.Fatal("one of --cache-key or --cache-key-files has to be provided")
	}
	if gcsFlags.Bucket == "" {
		log.Fatal("--bucket is required")
	}

	filesToCache, err := cache.MatchPatterns(filePatterns)
	if err != nil {
		log.Fatal(fmt.Errorf("error matching patterns: %w", err))
	}

	cacheKey, err := cache.GenerateCacheKey(cacheKey, cacheKeyFiles)
	if err != nil {
		log.Fatal(fmt.Errorf("error generating cache key: %w", err))
	}
	fmt.Println(cacheKey)

	ctx := context.Background()

	gcsWriter, err := gcs.NewGCSWriter(ctx, gcsFlags.Bucket, cacheKey)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = gcsWriter.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	gzipw := gzip.NewWriter(gcsWriter)
	defer gzipw.Close()

	err = tar.TarFiles(filesToCache, gzipw)
	if err != nil {
		log.Fatal(err)
	}
}

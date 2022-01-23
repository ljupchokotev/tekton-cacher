package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ingka-group-digital/es-tekton-cacher/cache"
	"github.com/ingka-group-digital/es-tekton-cacher/gcs"
	"github.com/ingka-group-digital/es-tekton-cacher/tar"
	"github.com/spf13/pflag"
)

var cacheKeyFiles []string
var cacheKey string
var resultFile string

var gcsFlags struct {
	Bucket     string
	PathPrefix string
}

func init() {
	pflag.StringSliceVar(&cacheKeyFiles, "cache-key-files", []string{}, "")
	pflag.StringVar(&cacheKey, "cache-key", "", "")

	pflag.StringVar(&gcsFlags.Bucket, "bucket", "", "")

	pflag.StringVar(&resultFile, "result-file", "", "")
}

func main() {
	pflag.Parse()

	if cacheKey == "" && len(cacheKeyFiles) == 0 {
		log.Fatal("one of --cache-key or --cache-key-files has to be provided")
	}
	if gcsFlags.Bucket == "" {
		log.Fatal("--bucket is required")
	}

	cacheKey, err := cache.GenerateCacheKey(cacheKey, cacheKeyFiles)
	if err != nil {
		log.Fatal(fmt.Errorf("error generating cache key: %w", err))
	}
	fmt.Println(cacheKey)

	ctx := context.Background()

	keyExists, err := gcs.KeyExists(ctx, gcsFlags.Bucket, cacheKey)
	if err != nil {
		log.Fatal(fmt.Errorf("error checking if key exists in cache: %w", err))
	}

	if !keyExists {
		if resultFile != "" {
			err := ioutil.WriteFile(resultFile, []byte("miss"), 0600)
			if err != nil {
				log.Fatal(fmt.Errorf("error writing result file: %w", err))
			}
		}

		os.Exit(0)
	}

	gcsReader, err := gcs.NewGCSReader(ctx, gcsFlags.Bucket, cacheKey)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = gcsReader.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	gzipw, err := gzip.NewReader(gcsReader)
	if err != nil {
		log.Fatal(err)
	}
	defer gzipw.Close()

	err = tar.Untar(gzipw)
	if err != nil {
		log.Fatal(err)
	}

	if resultFile != "" {
		err := ioutil.WriteFile(resultFile, []byte("hit"), 0600)
		if err != nil {
			log.Fatal(fmt.Errorf("error writing result file: %w", err))
		}
	}
}

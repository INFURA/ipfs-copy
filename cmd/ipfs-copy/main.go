package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gravitational/configure"

	ipfsCopy "github.com/INFURA/ipfs-copy"

	ipfsApi "github.com/ipfs/go-ipfs-api"
)

const DefaultApiUrl = "https://ipfs.infura.io:5001"
const DefaultWorkersCount = 5

type Config struct {
	ApiUrl        string `env:"IC_API_URL"             cli:"api_url"`
	File          string `env:"IC_CIDS"                cli:"cids"`
	SourceAPI     string `env:"IC_IPFS_SOURCE_API_URL" cli:"ipfs_source_api_url"`
	Workers       int    `env:"IC_WORKERS"             cli:"workers"`
	ProjectID     string `env:"IC_PROJECT_ID"          cli:"project_id"`
	ProjectSecret string `env:"IC_PROJECT_SECRET"      cli:"project_secret"`
}

func main() {
	cfg := mustParseConfigFromEnv()

	file, err := os.Open(cfg.File)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		err := file.Close()
		fmt.Println(err)
		os.Exit(1)
	}()

	shell := ipfsApi.NewShellWithClient(cfg.ApiUrl, NewClient(cfg.ProjectID, cfg.ProjectSecret))

	log.Printf("[INFO] Pinning the CIDs to %v with %v workers...", cfg.ApiUrl, cfg.Workers)
	successPinsCount, failedPinsCount, err := ipfsCopy.PinCIDsFromFile(file, cfg.Workers, shell)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Printf("[INFO] Successfully pinned %d CIDs", successPinsCount)
	log.Printf("[INFO] Failed to pin %d CIDs", failedPinsCount)
}

func mustParseConfigFromEnv() Config {
	var cfg Config

	err := configure.ParseEnv(&cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = configure.ParseCommandLine(&cfg, os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(cfg.File) == 0 && len(cfg.SourceAPI) == 0 {
		fmt.Println("IPFS Copy requires (IC_CIDS, --cids) OR (IC_IPFS_SOURCE_API_URL, --ipfs_source_api_url) to be defined.")
		os.Exit(1)
	}

	if len(cfg.ProjectID) == 0 {
		fmt.Println("IPFS Copy requires IC_PROJECT_ID env var or --project_id flag to be defined.")
		os.Exit(1)
	}

	if len(cfg.ProjectSecret) == 0 {
		fmt.Println("IPFS Copy requires IC_PROJECT_SECRET env var or --project_secret flag to be defined.")
		os.Exit(1)
	}

	if len(cfg.ApiUrl) == 0 {
		cfg.ApiUrl = DefaultApiUrl
	}

	if cfg.Workers == 0 {
		cfg.Workers = DefaultWorkersCount
	}

	return cfg
}

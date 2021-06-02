package main

import (
	"log"
	"os"

	ipfsCopy "github.com/INFURA/ipfs-copy"
	ipfsPump "github.com/INFURA/ipfs-pump/pump"
	"github.com/gravitational/configure"
	ipfsApi "github.com/ipfs/go-ipfs-api"
)

const DefaultApiUrl = "https://ipfs.infura.io:5001"
const DefaultWorkersCount = 5

type Config struct {
	ApiUrl        string `env:"IC_API_URL"        cli:"api-url"`
	File          string `env:"IC_CIDS"           cli:"cids"`
	FileFailed    string `env:"IC_CIDS_FAILED"    cli:"cids-failed"`
	SourceAPI     string `env:"IC_SOURCE_API_URL" cli:"source-api-url"`
	Workers       int    `env:"IC_WORKERS"        cli:"workers"`
	ProjectID     string `env:"IC_PROJECT_ID"     cli:"project-id"`
	ProjectSecret string `env:"IC_PROJECT_SECRET" cli:"project-secret"`

	// Helper settings

	// IsFileCopy is for existing users migration, pinning of files already existing on Infura IPFS nodes
	IsFileCopy bool

	// IsSourceAPICopy is for new users migrating their content from other IPFS nodes to Infura
	IsSourceAPICopy bool
}

func main() {
	cfg := mustParseConfigFromEnv()
	infuraShell := ipfsApi.NewShellWithClient(cfg.ApiUrl, NewClient(cfg.ProjectID, cfg.ProjectSecret))

	// Validates the credentials before spawning all the works
	_, _, err := infuraShell.Version()
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	var failedCIDsWriter ipfsPump.FailedBlocksWriter
	if cfg.FileFailed == "" {
		failedCIDsWriter = ipfsPump.NewNullableFileEnumeratorWriter()
	} else {
		enumWriter, closeWriter, err := ipfsPump.NewFileEnumeratorWriter(cfg.FileFailed)
		if err != nil {
			log.Fatalf("[ERROR] %v\n", err)
		}
		failedCIDsWriter = enumWriter

		defer func() {
			err = closeWriter()
			if err != nil {
				log.Fatalf("[ERROR] %v\n", err)
			}
		}()
	}

	if cfg.IsFileCopy {
		PinCIDsFromFile(cfg, infuraShell, failedCIDsWriter)
		os.Exit(0)
	}

	if cfg.IsSourceAPICopy {
		PumpBlocksAndCopyPins(cfg, infuraShell, failedCIDsWriter)
		os.Exit(0)
	}
}

func PinCIDsFromFile(cfg Config, infuraShell *ipfsApi.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) {
	file, err := os.Open(cfg.File)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}
	defer func() {
		err := file.Close()
		log.Fatalf("[ERROR] %v\n", err)
	}()

	log.Printf("[INFO] Pinning the CIDs to %v with %v workers...\n", cfg.ApiUrl, cfg.Workers)
	successPinsCount, failedPinsCount, err := ipfsCopy.PinCIDsFromFile(file, cfg.Workers, infuraShell, failedPinsWriter)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	log.Printf("[INFO] Successfully pinned %d CIDs\n", successPinsCount)
	log.Printf("[INFO] Failed to pin %d CIDs\n", failedPinsCount)
}

func PumpBlocksAndCopyPins(cfg Config, infuraShell *ipfsApi.Shell, failedCIDsWriter ipfsPump.FailedBlocksWriter) {
	pinEnum := ipfsPump.NewAPIPinEnumerator(cfg.SourceAPI, true)
	blocksColl := ipfsPump.NewAPICollector(cfg.SourceAPI)
	drain := ipfsPump.NewAPIDrainWithShell(infuraShell)

	// Copy all the blocks
	ipfsPump.PumpIt(pinEnum, blocksColl, drain, uint(cfg.Workers), failedCIDsWriter)
	log.Printf("[INFO] Copied %d blocks\n", drain.SuccessfulBlocksCount())

	// Once **all the blocks are copied**, pin the RECURSIVE + DIRECT pins (not before)
	successPinsCount, skippedIndirectPinsCount, failedPinsCount, err := ipfsCopy.PinCIDsFromSource(cfg.SourceAPI, cfg.Workers, infuraShell, failedCIDsWriter)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	log.Printf("[INFO] Successfully pinned %d CIDs\n", successPinsCount)
	log.Printf("[INFO] Skipped indirect %d CIDs\n", skippedIndirectPinsCount)
	log.Printf("[INFO] Failed to pin %d CIDs\n", failedPinsCount)
}

func mustParseConfigFromEnv() Config {
	var cfg Config

	err := configure.ParseEnv(&cfg)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	err = configure.ParseCommandLine(&cfg, os.Args[1:])
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	if len(cfg.File) == 0 && len(cfg.SourceAPI) == 0 {
		log.Fatal("[ERROR] IPFS Copy requires (IC_CIDS, --cids) OR (IC_SOURCE_API_URL, --source-api-url) to be defined.\n")
	}

	if len(cfg.File) > 0 {
		cfg.IsFileCopy = true
		cfg.IsSourceAPICopy = false
	} else {
		cfg.IsFileCopy = false
		cfg.IsSourceAPICopy = true
	}

	if len(cfg.ProjectID) == 0 {
		log.Fatal("[ERROR] IPFS Copy requires IC_PROJECT_ID env var or --project-id flag to be defined.\n")
	}

	if len(cfg.ProjectSecret) == 0 {
		log.Fatal("[ERROR] IPFS Copy requires IC_PROJECT_SECRET env var or --project-secret flag to be defined.\n")
	}

	if len(cfg.ApiUrl) == 0 {
		cfg.ApiUrl = DefaultApiUrl
	}

	if cfg.Workers == 0 {
		cfg.Workers = DefaultWorkersCount
	}

	return cfg
}

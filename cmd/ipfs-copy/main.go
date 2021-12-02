package main

import (
	"context"
	"fmt"
	ipfsCopy "github.com/INFURA/ipfs-copy"
	"github.com/INFURA/ipfs-copy/pump"
	ipfsPump "github.com/INFURA/ipfs-pump/pump"
	"github.com/gravitational/configure"
	ipfsApi "github.com/ipfs/go-ipfs-api"
	"log"
	"os"
)

const DefaultApiUrl = "https://ipfs.infura.io:5001"
const DefaultWorkersCount = 20
const DefaultMaxReqsPerSec = 10
const Version = "1.3.0"

type Config struct {
	ApiUrl        string `env:"IC_API_URL"         cli:"api-url"`
	File          string `env:"IC_CIDS"            cli:"cids"`
	FileFailed    string `env:"IC_CIDS_FAILED"     cli:"cids-failed"`
	SourceAPI     string `env:"IC_SOURCE_API_URL"  cli:"source-api-url"`
	Workers       int    `env:"IC_WORKERS"         cli:"workers"`
	MaxReqsPerSec int    `env:"IC_MAX_REQ_PER_SEC" cli:"max-req-per-sec"`
	ProjectID     string `env:"IC_PROJECT_ID"      cli:"project-id"`
	ProjectSecret string `env:"IC_PROJECT_SECRET"  cli:"project-secret"`

	// Helper settings

	// IsFileCopy is for existing users migration, pinning of files already existing on Infura IPFS nodes
	IsFileCopy bool

	// IsSourceAPICopy is for new users migrating their content from other IPFS nodes to Infura
	IsSourceAPICopy bool
}

func main() {
	AddVersionCmd()

	ctx := context.Background()
	cfg := mustPrepareConfig()
	infuraShell := ipfsApi.NewShellWithClient(cfg.ApiUrl, NewClient(cfg.ProjectID, cfg.ProjectSecret))

	// Validates the credentials before spawning all the workers
	infuraShellVersion, _, err := infuraShell.Version()
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}
	log.Printf("[INFO] Infura IPFS version is: %v", infuraShellVersion)

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

	log.Printf("[INFO] Pinning CIDs to %v with %v workers and maximum %v req/s/worker...\n", cfg.ApiUrl, cfg.Workers, cfg.MaxReqsPerSec)

	if cfg.IsFileCopy {
		PinCIDsFromFile(ctx, cfg, infuraShell, failedCIDsWriter)
		os.Exit(0)
	}

	if cfg.IsSourceAPICopy {
		PumpBlocksAndCopyPins(ctx, cfg, infuraShell, failedCIDsWriter, ipfsPump.NewProgressWriter())
		os.Exit(0)
	}
}

func AddVersionCmd() {
	// Support for `ipfs-copy version` command:
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("ipfs-copy version: %v\n", Version)
		os.Exit(0)
	}
}

func PinCIDsFromFile(ctx context.Context, cfg Config, infuraShell *ipfsApi.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) {
	file, err := os.Open(cfg.File)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatalf("[ERROR] %v\n", err)
		}
	}()

	successPinsCount, failedPinsCount, err := ipfsCopy.PinCIDsFromFile(ctx, file, cfg.Workers, cfg.MaxReqsPerSec, infuraShell, failedPinsWriter)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	log.Printf("[INFO] Successfully pinned %d CIDs\n", successPinsCount)
	log.Printf("[INFO] Failed to pin %d CIDs\n", failedPinsCount)
}

func PumpBlocksAndCopyPins(ctx context.Context, cfg Config, infuraShell *ipfsApi.Shell, failedCIDsWriter ipfsPump.FailedBlocksWriter, progressWriter ipfsPump.ProgressWriter) {
	sourceShell := ipfsApi.NewShell(cfg.SourceAPI)

	// Validate the connection and query the version so we know what features the source has (esp. pin ls --stream)
	sourceShellRawVersion, _, err := sourceShell.Version()
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	log.Printf("[INFO] Source IPFS version is: %v", sourceShellRawVersion)
	isEnumStreamingPossible := hasShellStreamPinListSupport(sourceShellRawVersion)
	log.Printf("[DEBUG] Source IPFS pin/ls --stream support?: %v", isEnumStreamingPossible)

	pinEnum := ipfsPump.NewAPIPinEnumerator(cfg.SourceAPI, isEnumStreamingPossible)
	blocksColl := ipfsPump.NewAPICollector(cfg.SourceAPI)
	drain := pump.NewRateLimitedDrain(ipfsPump.NewAPIDrainWithShell(infuraShell))

	// Copy all the blocks
	ipfsPump.PumpIt(pinEnum, blocksColl, drain, failedCIDsWriter, progressWriter, uint(cfg.Workers))
	log.Printf("[INFO] Copied %d blocks\n", drain.SuccessfulBlocksCount())

	// Once **all the blocks are copied**, pin the RECURSIVE + DIRECT pins (not before)
	successPinsCount, failedPinsCount, err := ipfsCopy.PinCIDsFromSource(ctx, cfg.Workers, cfg.MaxReqsPerSec, isEnumStreamingPossible, sourceShell, infuraShell, failedCIDsWriter)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	log.Printf("[INFO] Successfully pinned %d CIDs\n", successPinsCount)
	log.Printf("[INFO] Failed to pin %d CIDs\n", failedPinsCount)
}

func mustPrepareConfig() Config {
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

	if cfg.MaxReqsPerSec == 0 {
		cfg.MaxReqsPerSec = DefaultMaxReqsPerSec
	}

	return cfg
}

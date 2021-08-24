package ipfscopy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	ipfsPump "github.com/INFURA/ipfs-pump/pump"
	ipfsCid "github.com/ipfs/go-cid"
	ipfsShell "github.com/ipfs/go-ipfs-api"
)

// PinCIDsFromFile will open the file, read a CID from each line separated by LB char and pin them
// in parallel with multiple workers via the pre-configured shell.
func PinCIDsFromFile(ctx context.Context, file io.ReadSeeker, workers int, infuraShell *ipfsShell.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) (successPinsCount uint64, failedPinsCount uint64, err error) {
	cids := make(chan ipfsCid.Cid)

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to seek file to the start. %w", err)
	}

	// Read all the pins from a file
	go func() {
		readCIDs(file, cids)
		close(cids)
	}()

	successPinsCount, failedPinsCount = pinCIDs(cids, workers, infuraShell, failedPinsWriter)

	return successPinsCount, failedPinsCount, nil
}

// PinCIDsFromSource iterates over all pins, filters Recursive + Direct, and pins then in parallel with multiple workers via the pre-configured shell.
func PinCIDsFromSource(ctx context.Context, workers int, hasSourceShellListStreaming bool, sourceShell *ipfsShell.Shell, infuraShell *ipfsShell.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) (successPinsCount uint64, failedPinsCount uint64, err error) {
	pins := make(chan ipfsCid.Cid)
	successPinsCount = 0
	failedPinsCount = 0

	if hasSourceShellListStreaming {
		log.Printf("[INFO] Streaming pins from the source IPFS node...")
		err = streamPinsFromSource(ctx, pins, sourceShell)
		if err != nil {
			return 0, 0, err
		}
	} else {
		log.Printf("[INFO] Fetching pins from the source IPFS node to memory...")
		err = fetchPinsFromSource(pins, sourceShell)
		if err != nil {
			return 0, 0, err
		}
	}

	successPinsCount, failedPinsCount = pinCIDs(pins, workers, infuraShell, failedPinsWriter)

	return successPinsCount, failedPinsCount, nil
}

func streamPinsFromSource(ctx context.Context, cids chan ipfsCid.Cid, sourceShell *ipfsShell.Shell) error {
	pinStream, err := sourceShell.PinsStream(ctx)
	if err != nil {
		return err
	}

	go func() {
		for pinInfo := range pinStream {
			if pinInfo.Type == ipfsShell.IndirectPin {
				continue
			}

			c, err := ipfsCid.Parse(pinInfo.Cid)
			if err != nil {
				log.Printf("[ERROR] Failed parsing pin '%v' from stream. %v\n", pinInfo.Cid, err)
				continue
			}

			cids <- c
		}
		close(cids)
	}()

	return nil
}

// fetchPinsFromSource is a duplicate of streamPinsFromSource but without the streaming logic.
//
// The difference is:
// - IPFS version < 0.5.0 has no stream support
// - The Pins() loads all the pins from source into memory at once
// - The code can't be reused much or wrapped or decorated because the Shell returns quite different responses:
//     - PinsStream() returns <-chan **PinStreamInfo**
//     - Pins() returns map[string]**PinInfo**
//
// Hence it's easier to duplicate than create awkward abstractions in this case.
func fetchPinsFromSource(cids chan ipfsCid.Cid, sourceShell *ipfsShell.Shell) error {
	pins, err := sourceShell.Pins()
	if err != nil {
		return err
	}

	go func() {
		for cid, info := range pins {
			if info.Type == ipfsShell.IndirectPin {
				continue
			}

			c, err := ipfsCid.Parse(cid)
			if err != nil {
				log.Printf("[ERROR] Failed parsing pin '%v' from stream. %v\n", cid, err)
				continue
			}

			cids <- c
		}
		close(cids)
	}()

	return nil
}

func pinCIDs(cids <-chan ipfsCid.Cid, workers int, infuraShell *ipfsShell.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) (successPinsCount uint64, failedPinsCount uint64) {
	successPinsCount = 0
	failedPinsCount = 0

	// 5 workers (by default) will be handling pinning of the entire file
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			for cid := range cids {
				ok := pinCID(cid, infuraShell, failedPinsWriter)
				if ok {
					atomic.AddUint64(&successPinsCount, 1)
				} else {
					atomic.AddUint64(&failedPinsCount, 1)
				}
				// Avoid getting rate limited
				time.Sleep(100 * time.Millisecond)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// Flush all un-persisted (buffered) failed CIDs to the file
	err := failedPinsWriter.Flush()
	if err != nil {
		log.Printf("[ERROR] Unable to flush failed pins to a file. %v\n", err)
	}

	return successPinsCount, failedPinsCount
}

func pinCID(c ipfsCid.Cid, infuraShell *ipfsShell.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) bool {
	err := infuraShell.Pin(c.String())
	if err != nil {
		log.Printf("[ERROR] Failed pinning CID '%v'. %v", c, strings.TrimSpace(err.Error()))

		_, err := failedPinsWriter.Write(c)
		if err != nil {
			log.Printf("[ERROR] Unable to write failed CID '%v' pin to file. %v\n", c, err)
		}

		return false
	}

	log.Printf("[INFO] Pinned: '%v'", c)
	return true
}

func readCIDs(file io.ReadSeeker, cids chan<- ipfsCid.Cid) {
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		row := strings.Fields(fileScanner.Text())

		if len(row) < 1 {
			log.Printf("[ERROR] parsing CID. unexpected line: %s\n", fileScanner.Text())
			continue
		}

		c, err := ipfsCid.Parse(row[0])
		if err != nil {
			log.Printf("[ERROR] parsing CID. %s\n", fileScanner.Text())
			continue
		}

		cids <- c
	}
}

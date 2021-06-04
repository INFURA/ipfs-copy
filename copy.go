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
func PinCIDsFromSource(ctx context.Context, sourceApiUrl string, workers int, infuraShell *ipfsShell.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) (successPinsCount uint64, skippedIndirectPinsCount uint64, failedPinsCount uint64, err error) {
	cids := make(chan ipfsCid.Cid)
	successPinsCount = 0
	failedPinsCount = 0
	skippedIndirectPinsCount = 0

	sourceShell := ipfsShell.NewShell(sourceApiUrl)
	pinStream, err := sourceShell.PinsStream(context.Background())
	if err != nil {
		return 0, 0, 0, err
	}

	// Read all the pins from the source shell
	go func() {
		for pinInfo := range pinStream {
			if pinInfo.Type == ipfsShell.IndirectPin {
				//log.Printf("[DEBUG] Skipping indirect pin from stream: '%v'\n", pinInfo.Cid)
				skippedIndirectPinsCount++
				continue
			}

			c, err := ipfsCid.Parse(pinInfo.Cid)
			if err != nil {
				log.Printf("[ERROR] Failed parsing pin from stream. %v\n", err)
				continue
			}

			cids <- c
		}
		close(cids)
	}()

	successPinsCount, failedPinsCount = pinCIDs(cids, workers, infuraShell, failedPinsWriter)

	return successPinsCount, skippedIndirectPinsCount, failedPinsCount, nil
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

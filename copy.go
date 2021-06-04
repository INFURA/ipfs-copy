package ipfscopy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	ipfsPump "github.com/INFURA/ipfs-pump/pump"
	ipfsCid "github.com/ipfs/go-cid"
	ipfsShell "github.com/ipfs/go-ipfs-api"
)

// PinCIDsFromFile will open the file, read a CID from each line separated by LB char and pin them
// in parallel with multiple workers via the pre-configured shell.
func PinCIDsFromFile(file io.ReadSeeker, workers int, shell *ipfsShell.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) (successPinsCount int, failedPinsCount int, err error) {
	cids := make(chan ipfsCid.Cid)
	successPinsCount = 0
	failedPinsCount = 0

	// 5 workers (by default) will be handling pinning of the entire file
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			for cid := range cids {
				ok := pinCid(cid, shell)
				if ok {
					successPinsCount++
				} else {
					log.Printf("[ERROR] Failed pinning CID: '%v'\n", cid)
					failedPinsCount++
					_, err := failedPinsWriter.Write(cid)
					if err != nil {
						log.Printf("[ERROR] Failed writing pinning error to file for CID: '%v'. %v\n", cid, err)
					}
					continue
				}
			}
			wg.Done()
		}()
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to seek file to the start. %w", err)
	}

	readCIDs(file, cids)
	close(cids)
	wg.Wait()

	return successPinsCount, failedPinsCount, nil
}

// PinCIDsFromSource iterates over all pins, filters Recursive + Direct, and pins then in parallel with multiple workers via the pre-configured shell.
func PinCIDsFromSource(sourceApiUrl string, workers int, infuraShell *ipfsShell.Shell, failedPinsWriter ipfsPump.FailedBlocksWriter) (successPinsCount int, skippedIndirectPinsCount int, failedPinsCount int, err error) {
	cids := make(chan ipfsCid.Cid)
	successPinsCount = 0
	failedPinsCount = 0
	skippedIndirectPinsCount = 0

	sourceShell := ipfsShell.NewShell(sourceApiUrl)
	pinStream, err := sourceShell.PinsStream(context.Background())
	if err != nil {
		return 0, 0, 0, err
	}

	// Read all the pins
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

	// 5 workers (by default) will be handling pinning already filtered direct + recursive pins
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			for cid := range cids {
				ok := pinCid(cid, infuraShell)
				if ok {
					successPinsCount++
				} else {
					log.Printf("[ERROR] Failed pinning CID: '%v'\n", cid)
					failedPinsCount++
					_, err := failedPinsWriter.Write(cid)
					if err != nil {
						log.Printf("[ERROR] Failed writing pinning error to file for CID: '%v'. %v\n", cid, err)
					}
					continue
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return successPinsCount, skippedIndirectPinsCount, failedPinsCount, nil
}

func pinCid(c ipfsCid.Cid, infuraShell *ipfsShell.Shell) bool {
	err := infuraShell.Pin(c.String())
	if err != nil {
		log.Printf("[ERROR] Unable to pin '%v'. %v", c, strings.TrimSpace(err.Error()))
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

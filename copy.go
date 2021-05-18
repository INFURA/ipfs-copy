package ipfscopy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	ipfsCid "github.com/ipfs/go-cid"
	ipfsShell "github.com/ipfs/go-ipfs-api"
)

// PinCIDsFromFile will open the file, read a CID from each line separated by LB char and pin them
// in parallel with multiple workers via the pre-configured shell.
func PinCIDsFromFile(file io.ReadSeeker, workers int, shell *ipfsShell.Shell) (successPinsCount int, failedPinsCount int, err error) {
	cids := make(chan ipfsCid.Cid)
	successPinsCount = 0
	failedPinsCount = 0

	// 5 workers (by default) will be handling pinning of the entire file
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			for c := range cids {
				ok := pinCid(c, shell)
				if ok {
					successPinsCount++
				} else {
					failedPinsCount++
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

func pinCid(c ipfsCid.Cid, shell *ipfsShell.Shell) bool {
	err := shell.Pin(c.String())
	if err != nil {
		log.Printf("[ERROR] Unable to pin '%v'. %v", c, strings.TrimSpace(err.Error()))
		return false
	}

	log.Printf("[INFO] Successfully pinned: '%v'", c)
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

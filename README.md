# IPFS Copy
> Pin your existing IPFS files stored in Infura in 3 steps.

## Step 1 - Create your Infura IPFS Project
Register a new account at https://infura.io (if you don't have one already), and setup your IPFS project following the instructions.

At the end of the process you will be redirected to the settings page where you find your credentials:
- PROJECT_ID
- PROJECT_SECRET

![ipfs-copy Infura credentials settings page](./ipfs-copy-tutorial-creds.png)

## Step 2 - Define a file with CIDs you want to migrate/pin
Prepare a file with CIDs separated by a line-break character `\n`.

Example file:
```
QmaEZGiDrS7kDXMxbmpamrX1sPHZUM61a3YpjDoyaC3yfJ
QmTeRJXx623WwsoDk4371kh3JKCjoDcoWrqrhY9ekRasjE
QmUsQxC5bsgX53WhQ11DkxyB4uPYLEpdgmidFhGgUFK5aK
```

## Step 3 - Execute the `ipfs-copy` command to pin your files
Build the `ipfs-copy` tool:
```bash
go get -u github.com/INFURA/ipfs-copy
make install
```

### Run using flags:
```bash
ipfs-copy --file=/home/xxx/Documents/ipfs-cids.txt --project_id=<YOUR_PROJECT_ID> --project_secret=<YOUR_PROJECT_SECRET>
```
- optional flag `--api_url` defines the target, destination node to pin the files (**default:** https://ipfs.infura.io:5001)
- optional flag `--workers` defines how many CIDs should be pinned in parallel to speed-up files with many CIDs (**default:** 5)

### Run using ENV values (where the IC_API_URL, IC_FILE, IC_PROJECT_ID, IC_PROJECT_SECRET, IC_WORKERS are defined):
```bash
cp sample.env .env

source .env && ipfs-copy
```

What's going to happen?

The `ipfs-copy` command will read your file with all the IPFS hashes (CIDs), and pin them to your Infura IPFS project in parallel with multiple workers for optimal performance.

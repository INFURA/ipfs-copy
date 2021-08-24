# Migration - existing Infura users
## Step 1 - Create your Infura IPFS Project
To interact with the Infura IPFS API, you need to [register your account](https://infura.io/register) and set up your IPFS project.

After the registration, you will be redirected to the settings page, where you find your credentials to authenticate with:
- **PROJECT_ID**
- **PROJECT_SECRET**

![ipfs-copy Infura credentials settings page](./ipfs-copy-tutorial-creds.png)

## Step 2 - Prepare a migration file
Create a file containing IPFS CIDs separated by a line-break character `\n`.

Example file:
```
QmaEZGiDrS7kDXMxbmpamrX1sPHZUM61a3YpjDoyaC3yfJ
QmTeRJXx623WwsoDk4371kh3JKCjoDcoWrqrhY9ekRasjE
QmUsQxC5bsgX53WhQ11DkxyB4uPYLEpdgmidFhGgUFK5aK
```

## Step 3 - Execute the `ipfs-copy` command to pin your data
Build the `ipfs-copy` tool.

Using `go get`:
```bash
go get -u github.com/INFURA/ipfs-copy
```

Cloning the source code manually and compiling it:
```
git clone https://github.com/INFURA/ipfs-copy.git
cd ipfs-copy
make install
```

Or by downloading a [pre-built binary](https://github.com/INFURA/ipfs-copy/releases/tag/v1.0.0) from the release page.

### Run it
The `ipfs-copy` command will read your file with all the IPFS hashes (CIDs) and pin them to your Infura IPFS project in parallel with multiple workers for optimal performance.

#### Using flags
```bash
ipfs-copy --cids=/home/xxx/Documents/ipfs-cids.txt --project-id=<YOUR_PROJECT_ID> --project-secret=<YOUR_PROJECT_SECRET>
```
- optional flag `--workers=1` defines how many CIDs to pin in parallel (**default:** 1)
- optional flag `--max-req-per-sec=10` defines the maximum amount of CIDs a worker can pin per second to avoid getting rate limited (**default:** 10) 
- optional flag `--cids-failed=/tmp/failed_pins.txt` defines an absolute path where failed pins will be logged

#### Using ENV values
The `.env` file contains the following env variables:
- IC_CIDS
- IC_PROJECT_ID
- IC_PROJECT_SECRET
- IC_WORKERS
- IC_MAX_REQ_PER_SEC
- IC_CIDS_FAILED

```bash
cp existing-users-sample.env .env

source .env && ipfs-copy
```

Done! You have completed the migration to the new IPFS service!

---
## Bonus step - limit traffic from a specific IP
Configure your IP inside the **Allowlist -> Limit IP Access** in your project's security settings to restrict access only from a desired list of IPs.

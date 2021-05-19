# IPFS Copy
> Pin your existing IPFS data stored in Infura in 3 steps.

`ipfs-copy` is a migration tool for existing Infura users hosting their data on Infura IPFS nodes in the last months/years and want to migrate to the new, more reliable, performant Infura IPFS service with all the latest features.

You can **pin all your existing data**, currently hosted at Infura, **to the new service** with one command: `ipfs-copy.`

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

### Run it using flags
```bash
ipfs-copy --cids=/home/xxx/Documents/ipfs-cids.txt --project_id=<YOUR_PROJECT_ID> --project_secret=<YOUR_PROJECT_SECRET>
```
- optional flag `--api_url` defines the target, destination node to pin the data (**default:** https://ipfs.infura.io:5001)
- optional flag `--workers` defines how many CIDs to pin in parallel to speed-up data with many CIDs (**default:** 5)

### Run it using ENV values
The `.env` file contains the following env variables:
- IC_API_URL
- IC_CIDS
- IC_PROJECT_ID
- IC_PROJECT_SECRET
- IC_WORKERS

```bash
cp sample.env .env

source .env && ipfs-copy
```

What's going to happen?

The `ipfs-copy` command will read your file with all the IPFS hashes (CIDs) and pin them to your Infura IPFS project in parallel with multiple workers for optimal performance.

Done! You have completed the migration to the new IPFS service!

## Migrate data from self-hosted IPFS node to Infura
If you aren't an existing Infura customer and store your data on a self-hosted IPFS node, you can migrate too! Focus on your business and let Infura handle all the IPFS infrastructure and monitoring.

Tell the `ipfs-copy` command where to find the data to migrate by specifying the `--ipfs-source-url` pointing to your node's IPFS API.

### Example - migrate data from a local IPFS node to Infura IPFS service
```bash
ipfs-copy --ipfs-source-url=http://localhost:5001 --cids=/home/xxx/Documents/ipfs-cids.txt --project_id=<YOUR_PROJECT_ID> --project_secret=<YOUR_PROJECT_SECRET>
```

You can also use the `IC_IPFS_SOURCE_URL` ENV variable.
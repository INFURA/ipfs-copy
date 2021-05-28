# Migration - new Infura users
If you aren't an existing Infura customer and store your data on a self-hosted IPFS node, you can migrate too! Focus on your business and let Infura handle all the IPFS infrastructure and monitoring.

## Step 1 - Create your Infura IPFS Project
To interact with the Infura IPFS API, you need to [register your account](https://infura.io/register) and set up your IPFS project.

After the registration, you will be redirected to the settings page, where you find your credentials to authenticate with:
- **PROJECT_ID**
- **PROJECT_SECRET**

![ipfs-copy Infura credentials settings page](./ipfs-copy-tutorial-creds.png)

## Step 2 - Build the `ipfs-copy` command
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

## Step 3 - Migrate your content
### Run `ipfs-copy` using flags
```bash
ipfs-copy --ipfs-source-api-url=http://localhost:5001 --project_id=<YOUR_PROJECT_ID> --project_secret=<YOUR_PROJECT_SECRET>
```
- optional flag `--workers` defines how many pins to copy in parallel (**default:** 5)

### Run `ipfs-copy` using ENV variables
The `.env` contains:
- IC_IPFS_SOURCE_API_URL
- IC_PROJECT_ID
- IC_PROJECT_SECRET
- IC_WORKERS

```bash
cp new-self-hosted-users-sample.env .env

source .env && ipfs-copy
```

What's going to happen?

The `ipfs-copy` command will iterate all pins from the source node, copy the blocks and then pin them to your Infura IPFS project in parallel with multiple workers for optimal performance.

Done! You have migrated your content to Infura IPFS service!
# ZIGChain

**ZIGChain** is a blockchain built using Cosmos SDK and CometBFT.

## Prerequisites

Before you can build and run zigchain, you need to install the following dependencies:

### 1. Install Go (Latest Version)

**macOS:**
```bash
# Using Homebrew (recommended)
brew install go

# Or download from golang.org
# Visit https://golang.org/dl/ and download the latest version for macOS
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Or download from golang.org
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
```

**Verify Go installation:**
```bash
go version
```

### 2. Install Ignite v28

**macOS:**
```bash
# Download and install Ignite v28.10.0 for macOS (Apple Silicon)
curl -L https://github.com/ignite/cli/releases/download/v28.10.0/ignite_28.10.0_darwin_arm64.tar.gz -o ignite.tar.gz
tar -xzf ignite.tar.gz
sudo mv ignite /usr/local/bin/
rm ignite.tar.gz
```

**Linux:**
```bash
# Download and install Ignite v28.10.0 for Linux
curl -L https://github.com/ignite/cli/releases/download/v28.10.0/ignite_28.10.0_linux_amd64.tar.gz -o ignite.tar.gz
tar -xzf ignite.tar.gz
sudo mv ignite /usr/local/bin/
rm ignite.tar.gz
```

**Verify Ignite installation:**
```bash
ignite version
# Should show v28.10.0
```

### 3. Set up Go Environment

You need to add Go's bin directory to your PATH so you can run Go binaries from anywhere.

**macOS:**
```bash
# Add to your shell profile (~/.zshrc for zsh, ~/.bash_profile for bash)
echo 'export GOPATH=$HOME/go' >> ~/.zshrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.zshrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc

# Reload your shell configuration
source ~/.zshrc
```

**Linux:**
```bash
# Add to your shell profile (~/.bashrc for bash, ~/.zshrc for zsh)
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Reload your shell configuration
source ~/.bashrc
```

**Verify PATH setup:**
```bash
echo $PATH | grep go
# Should show /usr/local/go/bin and ~/go/bin in the output
```

## Building from Source

Once you have all dependencies installed, you can build the zigchain binary:

```bash
# Clone the repository (if you haven't already)
git clone <repository-url>
cd zigchain

# Build and install the binary
make install
```

This will:
1. Check that you have the required Go version (1.24+)
2. Download and verify dependencies
3. Build the `zigchaind` binary
4. Install it to `~/go/bin/zigchaind`

**Verify the installation:**
```bash
zigchaind version
```

## Development

```
ignite chain serve
```

`serve` command installs dependencies, builds, initializes, and starts your blockchain in development.

## Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Ignite CLI docs](https://docs.ignite.com).

## Full Documentation 

You can see full documentation in docs subdirectory.

## Release
To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

## Software Upgrade Governance Proposals

This section explains how to submit a new software upgrade governance proposal for ZIGChain.

### Prerequisites

- Account with sufficient tokens for proposal deposit
- Understanding of the upgrade process and requirements

### Step-by-Step Upgrade Process

#### 1. Define the New Version

First, define the new version you want to upgrade to (e.g., `v2`).

#### 2. Create Upgrade Directory Structure

Copy the existing upgrade folder and rename it for the new version:

```bash
# Copy the existing upgrade folder
cp -r app/upgrades/v1 app/upgrades/v2
```

#### 3. Update Constants

Edit the `constants.go` file in the new upgrade directory:

```bash
# Edit app/upgrades/v2_0_0/constants.go
```

Update the `UpgradeName` to match your new version:

```go
const UpgradeName = "v2"
```

#### 4. Add Migration Logic

Edit the `upgrades.go` file in the new upgrade directory to add any additional migration logic needed for the upgrade:

```bash
# Edit app/upgrades/v2/upgrades.go
```

Add your specific migration logic as needed.

#### 5. Update Setup Handlers

Edit `app/setup_handlers.go` and update the upgrades array:

```go
// Change from:
Upgrades = []upgrades.Upgrade{v1.Upgrade}

// To:
Upgrades = []upgrades.Upgrade{v2.Upgrade}
```

Make sure to import the new upgrade version:

```go
"zigchain/app/upgrades/v2"
```

#### 6. Build the New Binary

Build the new binary with the updated version:

```bash
VERSION=v2.0.0 make install
```

#### 7. Submit the Governance Proposal

Create a proposal file `/tmp/proposal.json`:

```json
{
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "zig10d07y265gmmuvt4z0w9aw880jnsr700jmgkh5m",
      "plan": {
        "name": "v2",
        "height": "1000000",
        "info": "{}"
      }
    }
  ],
  "metadata": "{\"title\":\"Upgrade\",\"authors\":[],\"summary\":\"Upgrade\",\"details\":\"\",\"proposal_forum_url\":\"\",\"vote_option_context\":\"\"}",
  "deposit": "500uzig",
  "title": "Upgrade",
  "summary": "Upgrade",
  "expedited": false
}
```

Submit the proposal:

```bash
zigchaind tx gov submit-proposal /tmp/proposal.json --from $ACCOUNT_NAME -y
```

#### 8. Vote on the Proposal

Vote on the new proposal:

```bash
zigchaind tx gov vote 1 yes --from $ACCOUNT_NAME -y
```

#### 9. Monitor the Upgrade

When the upgrade height is reached, the node logs should show:

```
[...]
[ZIGCHAIND] 11:05PM ERR CONSENSUS FAILURE!!! err="failed to apply block; error UPGRADE \"v2\" NEEDED at height: 106
[...]
```

#### 10. Apply the Upgrade

1. **Stop the node**
2. **Switch the binary** from the old version to the new one
3. **Start the node again**:

```bash
zigchaind start
```

#### 11. Verify Successful Upgrade

The node logs should show successful upgrade completion:

```
[...]
11:09PM INF applying upgrade "v2" at height: 106 module=x/upgrade
11:09PM INF Starting module migrations... module=server
11:09PM INF migrating module dex from version 1 to version 2 module=server
11:09PM INF migrating module factory from version 1 to version 2 module=server
11:09PM INF adding a new module: packetfowardmiddleware module=server
11:09PM INF adding a new module: ratelimit module=server
11:09PM INF migrating module tokenwrapper from version 1 to version 2 module=server
11:09PM INF Upgrade v2 complete module=server
[...]
```

### Troubleshooting

#### Missing Upgrade Handler Error

If the new upgrade plan is missing in the setup handlers function, starting the node will result in this error:

```
[...]
11:13PM ERR CONSENSUS FAILURE!!! err="failed to apply block; error wrong app version 0, upgrade handler is missing for v2 upgrade plan"
[...]
```

**Solution**: Ensure you've properly updated `app/setup_handlers.go` with the new upgrade version and imported the correct package.

#### Common Issues

1. **Incorrect upgrade name**: Ensure the upgrade name in the proposal matches exactly with the `UpgradeName` constant
2. **Missing import**: Verify that the new upgrade package is properly imported in `setup_handlers.go`
3. **Insufficient deposit**: Make sure your account has enough tokens for the proposal deposit
4. **Wrong authority**: Ensure you're using the correct authority address for the proposal

## Development Tools

To install `mockgen`, which is used for generating mocks in Go, follow these steps:

1. **Ensure Go is installed**: Verify that you have Go installed by running `go version`. If not, install it from [golang.org](https://golang.org/dl/).

2. **Install mockgen**:
   Run the following command to install `mockgen`:
   ```bash
   go install go.uber.org/mock/mockgen@latest
   ```

3. **Verify installation**:
   Check if `mockgen` is installed by running:
   ```bash
   mockgen --version
   ```
   This should display the version of `mockgen`.

4. **Run the following script to generate mock files**:
   ```bash
   ./sh/mockgen.sh
   ```

### Notes:
- If you encounter issues, ensure your Go environment is set up correctly (`GOPATH`, `GOBIN`, etc.).
- Make sure `~/go/bin` is in your PATH as described in the prerequisites section.

## Deterministic and Cross-Platform Builds

This guide explains how to build deterministic binaries across multiple architectures using the ZIGChain source code. This process ensures that the same source code produces identical binaries across different platforms and architectures.

### Prerequisites

- Git installed on your system
- Go development environment set up
- Docker installed on your system
- Make utility available
- Access to the ZIGChain GitHub repository

### Step-by-Step Build Process

#### 1. Clone the Public ZIGChain Repository

```bash
git clone git@github.com:ZIGChain/zigchain.git
cd zigchain
```

#### 2. Fetch All Tags

```bash
git fetch --all --tags
```

This ensures you have access to all available release tags in the repository.

#### 3. Checkout the Desired Version

```bash
git checkout v2.0.0
```

Replace `v2.0.0` with the actual version you want to build. You can list available tags with:

```bash
git tag -l
```

#### 4. Run the Go Releaser Build Command

```bash
VERSION=v2.0.0 make goreleaser-build-local
```

Replace `v2.0.0` with the version number you want to build (with the 'v' prefix).

#### 5. Build Output

The build process will create a `dist` folder containing all binaries for different architectures:

```
$ tree dist
./dist
├── SHA256SUMS-v2.0.0.txt
├── artifacts.json
├── config.yaml
├── metadata.json
├── zigchaind-darwin-amd64_darwin_amd64_v1
│   └── zigchaind
├── zigchaind-darwin-arm64_darwin_arm64_v8.0
│   └── zigchaind
├── zigchaind-darwin-universal_darwin_all
│   └── zigchain
├── zigchaind-linux-amd64_linux_amd64_v1
│   └── zigchaind
├── zigchaind-v2.0.0-darwin-amd64.tar.gz
├── zigchaind-v2.0.0-darwin-arm64.tar.gz
└── zigchaind-v2.0.0-linux-amd64.tar.gz
```

#### 6. Available Architectures

The build process generates binaries for the following architectures:

- **darwin-amd64**: macOS on Intel processors
- **darwin-arm64**: macOS on Apple Silicon (M1/M2/M3)
- **darwin-universal**: Universal binary for macOS (supports both Intel and Apple Silicon)
- **linux-amd64**: Linux on Intel/AMD processors

#### 7. Verify Checksums

To ensure the integrity of your builds, compare the generated checksums with the official ones:

```bash
$ cat ./dist/SHA256SUMS-v2.0.0.txt 
ff7be08285d798ec6aa7d8ab6841d98b9ae9e2a83ba1f7ce3125092718caec76  zigchaind-v2.0.0-darwin-amd64
b7adaaee85eb51c38b77f135dc0eda70b3071abfe18f1b196beb964b0de23434  zigchaind-v2.0.0-darwin-amd64.tar.gz
cd9b65749b5edf38f88835afc060d3c48a02f267ca5ddb264d42561d4b3664a1  zigchaind-v2.0.0-darwin-arm64
9795fe8b25f65e508d3e027bc58070ff2f96db47c9636dbd71e4ee7958d95f02  zigchaind-v2.0.0-darwin-arm64.tar.gz
600ccd9432b44648946671c0c325ad912aec02126e2d9549317218bc03a2e207  zigchaind-v2.0.0-linux-amd64
3d7d9a823e92f4b6f0ca9e260e2df74af31f7c1eaaad47f39c464ed748ce11d3  zigchaind-v2.0.0-linux-amd64.tar.gz
```

**Important**: These checksums should match the ones published in the official repository:
https://github.com/ZIGChain/networks/blob/main/zig-test-2/binaries/SHA256SUMS-v2.0.0.txt

#### 8. Deployment

Select the appropriate binary for your target architecture:

- For **macOS Intel**: Use `zigchaind-darwin-amd64_darwin_amd64_v1/zigchaind`
- For **macOS Apple Silicon**: Use `zigchaind-darwin-arm64_darwin_arm64_v8.0/zigchaind`
- For **macOS Universal**: Use `zigchaind-darwin-universal_darwin_all/zigchain`
- For **Linux**: Use `zigchaind-linux-amd64_linux_amd64_v1/zigchaind`

Or use the corresponding `.tar.gz` archives for easier distribution.

### Troubleshooting

#### Build Failures

If the build fails, ensure:

1. You have the correct Go version installed
2. All dependencies are properly installed
3. You're using the correct version tag
4. Your system has sufficient disk space

#### Checksum Mismatches

If checksums don't match the official ones:

1. Verify you're building from the correct commit
2. Ensure no local modifications to the source code
3. Check that you're using the exact same version number
4. Rebuild from a clean repository clone

#### Cross-Platform Building

For building on different platforms:

- **macOS**: Can build all architectures natively
- **Linux**: Can build Linux binaries natively, may need Docker for other platforms
- **Windows**: May require WSL or Docker for cross-compilation

### Security Considerations

- Always verify checksums before deployment
- Use official release tags only
- Avoid building from custom branches unless necessary
- Keep your build environment clean and up-to-date

### Additional Resources

- [ZIGChain GitHub Repository](https://github.com/ZIGChain/zigchain)
- [Official Network Binaries](https://github.com/ZIGChain/networks)

# Homebrew Installation Guide

This guide explains how to set up your `ask` CLI tool as a Homebrew package.

## Prerequisites

1. Your Go project should be hosted on GitHub
2. You need to create releases with binaries for multiple platforms

## Setup Steps

### 1. Build Binaries for All Platforms

Run the following command to build binaries for all supported platforms:

```bash
make build-all
```

This will create:
- `ask-darwin-amd64` (macOS Intel)
- `ask-darwin-arm64` (macOS Apple Silicon)
- `ask-linux-amd64` (Linux Intel)
- `ask-linux-arm64` (Linux ARM)

### 2. Calculate SHA256 Hashes

Run the following command to get the SHA256 hashes needed for the Homebrew formula:

```bash
make sha256
```

### 3. Update the Homebrew Formula

1. Update `Formula/ask.rb` with:
   - Your actual GitHub username in the `homepage` and `url` fields
   - The correct SHA256 hashes from step 2
   - The correct version number

2. Example of updated formula:
```ruby
class Ask < Formula
  desc "A CLI tool to get ChatGPT answers from the command line"
  homepage "https://github.com/yourusername/ask"
  version "1.0.0"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/yourusername/ask/releases/download/v1.0.0/ask-darwin-arm64"
    sha256 "actual_sha256_hash_here"
  # ... rest of the formula
end
```

### 4. Create a GitHub Release

1. Tag your release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. The GitHub Actions workflow will automatically:
   - Build binaries for all platforms
   - Create a release with the binaries attached

### 5. Set Up Homebrew Tap

#### Option A: Personal Tap (Recommended for testing)

1. Create a new repository named `homebrew-tap` on GitHub
2. Add the formula to the repository:
```bash
mkdir homebrew-tap
cp Formula/ask.rb homebrew-tap/
cd homebrew-tap
git init
git add ask.rb
git commit -m "Add ask formula"
git remote add origin https://github.com/yourusername/homebrew-tap.git
git push -u origin main
```

3. Install from your tap:
```bash
brew tap yourusername/tap
brew install ask
```

#### Option B: Submit to Homebrew Core

For wider distribution, you can submit your formula to Homebrew Core:

1. Fork the [homebrew-core](https://github.com/Homebrew/homebrew-core) repository
2. Add your formula to `Formula/ask.rb`
3. Submit a pull request

### 6. Test the Installation

```bash
# Install the formula
brew install yourusername/tap/ask

# Test it works
ask "Hello, world!"
```

## Updating the Formula

When you release a new version:

1. Update the version in `Formula/ask.rb`
2. Update the SHA256 hashes
3. Create a new GitHub release
4. Update your tap repository

## Troubleshooting

- If the formula fails to install, check that the SHA256 hashes are correct
- Ensure all platform binaries are available in the GitHub release
- Verify that the URLs in the formula point to the correct release assets 
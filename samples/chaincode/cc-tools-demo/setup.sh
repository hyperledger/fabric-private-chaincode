#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Set the CC_TOOLS_DEMO_PATH environment variable
export CC_TOOLS_DEMO_PATH="$FPC_PATH/samples/chaincode/cc-tools-demo"

# Clone the repository with sparse checkout
git clone -n --no-checkout --depth=1 --filter=tree:0 https://github.com/hyperledger-labs/cc-tools-demo.git "$CC_TOOLS_DEMO_PATH/chaincode"

# Navigate to the cloned directory
cd "$CC_TOOLS_DEMO_PATH/chaincode" || { echo "$CC_TOOLS_DEMO_PATH/chaincode does not exist. Exiting." >&2; exit 1; }

# Configure Git
git config --global --add safe.directory /src/github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/chaincode

# Enable sparse checkout
git sparse-checkout set --no-cone chaincode/*

# Checkout the sparse files
git checkout

# Move the chaincode files to the destination directory
mv chaincode/* "$CC_TOOLS_DEMO_PATH"

# Navigate to the CC_TOOLS_DEMO_PATH directory
cd "$CC_TOOLS_DEMO_PATH"

# Remove the now-empty chaincode directory
rm -r "$CC_TOOLS_DEMO_PATH/chaincode"

echo "Script execution completed successfully."

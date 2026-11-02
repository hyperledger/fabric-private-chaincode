#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Set the CC_TOOLS_DEMO_PATH environment variable
export CC_TOOLS_DEMO_PATH="$FPC_PATH/samples/chaincode/cc-tools-demo"

# The cc-tools-demo commit this tutorial is tested with; it uses cc-tools/v2 and fabric-chaincode-go/v2, as FPC does
CC_TOOLS_DEMO_COMMIT=18a9af486de0f0313358ec585d601c28857270c6

# Fetch the chaincode folder of that commit with a sparse checkout
git init -q "$CC_TOOLS_DEMO_PATH/chaincode"

# Navigate to the cloned directory
cd "$CC_TOOLS_DEMO_PATH/chaincode" || { echo "$CC_TOOLS_DEMO_PATH/chaincode does not exist. Exiting." >&2; exit 1; }

# Configure Git
git config --global --add safe.directory /src/github.com/hyperledger/fabric-private-chaincode/samples/chaincode/cc-tools-demo/chaincode

git remote add origin https://github.com/hyperledger-labs/cc-tools-demo.git
git fetch -q --depth=1 --filter=tree:0 origin "$CC_TOOLS_DEMO_COMMIT"

# Enable sparse checkout
git sparse-checkout set --no-cone 'chaincode/*'

# Checkout the sparse files
git checkout -q FETCH_HEAD

# Move the chaincode files to the destination directory
mv chaincode/* "$CC_TOOLS_DEMO_PATH"

# Navigate to the CC_TOOLS_DEMO_PATH directory
cd "$CC_TOOLS_DEMO_PATH"

# Remove the now-empty chaincode directory
rm -rf "$CC_TOOLS_DEMO_PATH/chaincode"

echo "Script execution completed successfully."

#!/bin/bash

# FPC Multi-User Testing System
# Description: Interactive testing system for Alice, Bob, and Monitor

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Global variables
STATE_DIR="/tmp/fpc_test_state"
RUNNING_IN_DOCKER="false"
MAIN_SCRIPT_SOURCED="false"
USER_MODE=""
USER_NAME=""
USER_ORG=""
CERT_HASH=""

# State files
ACTIVITY_LOG="$STATE_DIR/activity.log"
ALICE_STATE="$STATE_DIR/alice.state"
BOB_STATE="$STATE_DIR/bob.state"
SHARED_STATE="$STATE_DIR/shared.state"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_activity() {
    local timestamp=$(date '+%H:%M:%S')
    local user="$1"
    local action="$2"
    echo "[$timestamp] $user: $action" >>"$ACTIVITY_LOG"
}

# Initialize state directory and files
init_state() {
    mkdir -p "$STATE_DIR"
    touch "$ACTIVITY_LOG" "$ALICE_STATE" "$BOB_STATE" "$SHARED_STATE"

    # Initialize shared state if empty
    if [ ! -s "$SHARED_STATE" ]; then
        cat >"$SHARED_STATE" <<'EOF'
DIGITAL_ASSET_UUID=""
DIGITAL_ASSET_JSON=""
ESCROWS=()
SYSTEM_INITIALIZED="false"
EOF
    fi
}

# Source state files
load_state() {
    if [ -f "$SHARED_STATE" ]; then
        source "$SHARED_STATE"
    fi

    if [ -n "$USER_MODE" ] && [ -f "${STATE_DIR}/${USER_MODE}.state" ]; then
        source "${STATE_DIR}/${USER_MODE}.state"
    fi
}

# Save user-specific state
save_user_state() {
    cat >"${STATE_DIR}/${USER_MODE}.state" <<EOF
WALLET_UUID="${WALLET_UUID:-}"
WALLET_ID="${WALLET_ID:-}"
WALLET_BALANCE="${WALLET_BALANCE:-0}"
ESCROW_BALANCE="${ESCROW_BALANCE:-0}"
LAST_ESCROW_UUID="${LAST_ESCROW_UUID:-}"
LAST_ESCROW_SECRET="${LAST_ESCROW_SECRET:-}"
LAST_PARCEL_ID="${LAST_PARCEL_ID:-}"
EOF
}

# Save shared state
save_shared_state() {
    cat >"$SHARED_STATE" <<EOF
DIGITAL_ASSET_UUID="${DIGITAL_ASSET_UUID:-}"
DIGITAL_ASSET_JSON='${DIGITAL_ASSET_JSON:-}'
ESCROWS=(${ESCROWS[@]})
SYSTEM_INITIALIZED="${SYSTEM_INITIALIZED:-false}"
EOF
}

# Check if FPC_PATH is set
check_fpc_path() {
    if [ -z "$FPC_PATH" ]; then
        log_error "FPC_PATH is not set. Please export FPC_PATH=/path/to/your/fpc"
        exit 1
    fi
}

# Setup user environment
setup_user_env() {
    local user="$1"
    local env_file

    case "$user" in
    "alice")
        USER_MODE="alice"
        USER_NAME="alice"
        USER_ORG="Org1MSP"
        CERT_HASH="sha256:alice_cert"
        env_file="$FPC_PATH/samples/chaincode/confidential-escrow/.env.alice"
        ;;
    "bob")
        USER_MODE="bob"
        USER_NAME="bob"
        USER_ORG="Org2MSP"
        CERT_HASH="sha256:bob_cert"
        env_file="$FPC_PATH/samples/chaincode/confidential-escrow/.env.bob"
        ;;
    *)
        log_error "Invalid user. Use 'alice' or 'bob'"
        exit 1
        ;;
    esac

    # Check if environment file exists
    if [ ! -f "$env_file" ]; then
        log_error "Environment file not found at: $env_file"
        exit 1
    fi

    # Source the user's environment
    source "$env_file"
    log_info "Loaded $USER_NAME environment from $(basename $env_file)"

    # Check if this peer's enclave is initialized
    # local enclave_marker="/tmp/fpc_enclave_${USER_MODE}_initialized"
    local enclave_marker="/tmp/fpc_enclave_initialized"

    if [ ! -f "$enclave_marker" ]; then
        log_info "$USER_NAME's enclave not initialized. Setting up now..."

        # Source main.sh if not already done
        if [ "$MAIN_SCRIPT_SOURCED" = "false" ]; then
            source_main_script
            MAIN_SCRIPT_SOURCED="true"
        fi

        # Setup docker environment for this specific user
        setup_docker "$env_file"

        # Mark this peer as initialized
        touch "$enclave_marker"
        log_success "$USER_NAME's environment initialized!"
        echo
        read -p "Press Enter to continue to $USER_NAME's interface..."
    else
        log_info "$USER_NAME's environment already initialized"
    fi

    # Navigate to fpcclient directory
    cd "$FPC_PATH/samples/application/simple-cli-go"
}

# Extract data from chaincode responses
extract_uuid() {
    local output="$1"
    local asset_type="$2"
    echo "$output" | grep '^>' | sed 's/^> //' | grep -o "\"@key\":\"${asset_type}:[^\"]*\"" | cut -d':' -f3 | tr -d '"'
}

# Extract only the asset UUID, not the full JSON with metadata
extract_asset_uuid() {
    local output="$1"
    echo "$output" | grep '^>' | sed 's/^> //' | grep -o '"@key":"digitalAsset:[^"]*"' | cut -d':' -f3 | tr -d '"'
}

# Extract balance from response
extract_balance() {
    local output="$1"
    echo "$output" | grep -o '"balance":[0-9]*' | head -1 | cut -d':' -f2
}

extract_escrow_balance() {
    local output="$1"
    echo "$output" | grep -o '"escrowBalance":[0-9]*' | head -1 | cut -d":" -f2
}

# Run fpcclient command with error handling
# Run fpcclient command with proper error handling
run_fpcclient() {
    local cmd_type="$1"
    local function_name="$2"
    local args="$3"
    local desc="$4"

    log_info "$desc"
    echo -e "${CYAN}Command: ./fpcclient $cmd_type $function_name${NC}"
    echo -e "${CYAN}Payload: $args${NC}"
    echo

    local output
    if [ "$cmd_type" = "invoke" ]; then
        output=$(./fpcclient invoke "$function_name" "$args" 2>&1)
    else
        output=$(./fpcclient query "$function_name" "$args" 2>&1)
    fi

    local exit_code=$?

    echo -e "${YELLOW}════════ FPCCLIENT OUTPUT START ════════${NC}"
    echo "$output"
    echo -e "${YELLOW}════════ FPCCLIENT OUTPUT END ══════════${NC}"
    echo

    if [ $exit_code -eq 0 ]; then
        log_success "$desc - COMPLETED"
        return 0
    else
        log_error "$desc - FAILED (Exit Code: $exit_code)"
        return 1
    fi
}

# Dashboard functions
show_dashboard() {
    clear
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}                 ${YELLOW}FPC MULTI-USER TEST SYSTEM${NC}                    ${CYAN}║${NC}"
    echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"

    if [ "$USER_MODE" = "monitor" ]; then
        show_monitor_dashboard
    else
        show_user_dashboard
    fi

    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
}

show_user_dashboard() {
    load_state

    local wallet_display="${WALLET_UUID}"
    [ -n "$WALLET_UUID" ] && wallet_display="${wallet_display}" || wallet_display="Not created"

    echo -e "${CYAN}║${NC} ${GREEN}User:${NC} $USER_NAME ($USER_ORG)                                    ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC} ${GREEN}Wallet UUID:${NC} ${wallet_display}                                   ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC} ${GREEN}Balance:${NC} ${WALLET_BALANCE:-0} CBDC                                  ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC} ${GREEN}Escrow Balance:${NC} ${ESCROW_BALANCE:-0} CBDC                           ${CYAN}║${NC}"
    echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"

    # Show recent activity
    echo -e "${CYAN}║${NC} ${YELLOW}Recent Activity:${NC}                                           ${CYAN}║${NC}"
    if [ -f "$ACTIVITY_LOG" ] && [ -s "$ACTIVITY_LOG" ]; then
        tail -n 3 "$ACTIVITY_LOG" | while IFS= read -r line; do
            printf "${CYAN}║${NC} %-58s ${CYAN}║${NC}\n" "${line:0:58}"
        done
    else
        printf "${CYAN}║${NC} %-58s ${CYAN}║${NC}\n" "No activity yet"
    fi
}

show_monitor_dashboard() {
    load_state

    local asset_display="${DIGITAL_ASSET_UUID}"
    [ -n "$DIGITAL_ASSET_UUID" ] && asset_display="${asset_display}..." || asset_display="Not created"

    echo -e "${CYAN}║${NC}                     ${YELLOW}SYSTEM MONITOR${NC}                           ${CYAN}║${NC}"
    echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${CYAN}║${NC} ${GREEN}System Status:${NC} ${SYSTEM_INITIALIZED:-Not initialized}              ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC} ${GREEN}Digital Asset:${NC} ${DIGITAL_ASSET_UUID:-'Not created'}               ${CYAN}║${NC}"
    echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"

    # Show user states
    echo -e "${CYAN}║${NC} ${YELLOW}Users Status:${NC}                                            ${CYAN}║${NC}"

    if [ -f "$ALICE_STATE" ]; then
        source "$ALICE_STATE"
        local alice_wallet="${WALLET_UUID}"
        [ -n "$WALLET_UUID" ] && alice_wallet="${alice_wallet}" || alice_wallet="No wallet"
        printf "${CYAN}║${NC} Alice: Wallet: %-20s Balance: %-10s ${CYAN}║${NC}\n" "$alice_wallet" "${WALLET_BALANCE:-0}"
    else
        echo -e "${CYAN}║${NC} Alice: Not active                                        ${CYAN}║${NC}"
    fi

    if [ -f "$BOB_STATE" ]; then
        source "$BOB_STATE"
        local bob_wallet="${WALLET_UUID}"
        [ -n "$WALLET_UUID" ] && bob_wallet="${bob_wallet}" || bob_wallet="No wallet"
        printf "${CYAN}║${NC} Bob:   Wallet: %-20s Balance: %-10s ${CYAN}║${NC}\n" "$bob_wallet" "${WALLET_BALANCE:-0}"
    else
        echo -e "${CYAN}║${NC} Bob:   Not active                                        ${CYAN}║${NC}"
    fi

    echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"

    # Show recent activity
    echo -e "${CYAN}║${NC} ${YELLOW}Live Activity Log:${NC}                                       ${CYAN}║${NC}"
    if [ -f "$ACTIVITY_LOG" ] && [ -s "$ACTIVITY_LOG" ]; then
        tail -n 5 "$ACTIVITY_LOG" | while IFS= read -r line; do
            printf "${CYAN}║${NC} %-58s ${CYAN}║${NC}\n" "${line:0:58}"
        done
    else
        printf "${CYAN}║${NC} %-58s ${CYAN}║${NC}\n" "No activity yet"
    fi
}

# Menu functions
show_main_menu() {
    echo
    echo -e "${YELLOW}═══ MAIN MENU ═══${NC}"
    # echo "1. Setup System (Initialize Digital Asset)"
    echo "1. Wallet Operations"
    echo "2. Token Operations"
    echo "3. Escrow Operations"
    echo "4. Query Operations"
    echo "5. Refresh Dashboard"
    echo "6. View Activity Log"
    echo "0. Exit"
    echo
}

show_wallet_menu() {
    echo
    echo -e "${YELLOW}═══ WALLET OPERATIONS ═══${NC}"
    echo "1. Create Wallet"
    echo "2. Check Balance"
    echo "3. Query My Wallet Details"
    echo "0. Back to Main Menu"
    echo
}

show_token_menu() {
    echo
    echo -e "${YELLOW}═══ TOKEN OPERATIONS ═══${NC}"
    echo "1. Mint Tokens"
    echo "2. Transfer Tokens"
    echo "3. Burn Tokens"
    echo "0. Back to Main Menu"
    echo
}

show_escrow_menu() {
    echo
    echo -e "${YELLOW}═══ ESCROW OPERATIONS ═══${NC}"
    echo "1. Create Escrow"
    echo "2. Verify Escrow Condition"
    echo "3. Release Escrow"
    echo "4. Refund Escrow"
    echo "5. Check Escrow Balance"
    echo "6. Query Escrow Details"
    echo "0. Back to Main Menu"
    echo
}

show_query_menu() {
    echo
    echo -e "${YELLOW}═══ QUERY OPERATIONS ═══${NC}"
    echo "1. Query My Wallet"
    echo "2. Query Digital Asset"
    echo "3. Get System Schema"
    echo "4. Get My Balance"
    echo "5. Get Wallet By Owner"
    echo "0. Back to Main Menu"
    echo
}

show_monitor_menu() {
    echo
    echo -e "${YELLOW}═══ MONITOR MENU ═══${NC}"
    echo "1. Refresh Dashboard"
    echo "2. View Full Activity Log"
    echo "3. Clear Activity Log"
    echo "4. Export System State"
    echo "5. View Alice State"
    echo "6. View Bob State"
    echo "0. Exit"
    echo
}

# System operations
# setup_system() {
#     log_info "Setting up digital asset..."
#     load_state
#
#     if [ "$SYSTEM_INITIALIZED" = "true" ]; then
#         log_info "System already initialized with asset: ${DIGITAL_ASSET_UUID}..."
#         return 0
#     fi
#
#     local output
#     if output=$(run_fpcclient "invoke" "createDigitalAsset" '{
#         "name": "CBDC",
#         "symbol": "CBDC",
#         "decimals": 2,
#         "totalSupply": 1000000,
#         "owner": "central_bank",
#         "issuerHash": "sha256:central_bank_cert"
#     }' "Creating digital asset CBDC"); then
#         DIGITAL_ASSET_UUID=$(extract_uuid "$output" "digitalAsset")
#         DIGITAL_ASSET_JSON=$(echo "$output" | grep '^>' | sed 's/^> //')
#         SYSTEM_INITIALIZED="true"
#         save_shared_state
#         log_activity "$USER_NAME" "Created digital asset CBDC (UUID: ${DIGITAL_ASSET_UUID}...)"
#         log_success "Digital asset created successfully!"
#         echo "UUID: $DIGITAL_ASSET_UUID"
#     else
#         log_error "Failed to create digital asset"
#         return 1
#     fi
# }

# Wallet operations
create_wallet() {
    load_state

    if [ -n "$WALLET_UUID" ]; then
        log_info "Wallet already exists: ${WALLET_UUID}..."
        return 0
    fi

    # Create digital asset if not exists
    if [ "$SYSTEM_INITIALIZED" != "true" ]; then
        log_info "Digital asset not found. Creating automatically..."
        local output
        if output=$(run_fpcclient "invoke" "createDigitalAsset" '{
            "name": "CBDC",
            "symbol": "CBDC",
            "decimals": 2,
            "totalSupply": 1000000,
            "owner": "central_bank",
            "issuerHash": "sha256:central_bank_cert"
        }' "Creating digital asset CBDC"); then
            DIGITAL_ASSET_UUID=$(extract_uuid "$output" "digitalAsset")
            DIGITAL_ASSET_JSON=$(echo "$output" | grep '^>' | sed 's/^> //')
            SYSTEM_INITIALIZED="true"
            save_shared_state
            log_activity "$USER_NAME" "Created digital asset CBDC (UUID: ${DIGITAL_ASSET_UUID}...)"
            log_success "Digital asset created automatically!"
        else
            echo "$output"
            log_error "Failed to create digital asset"
            return 1
        fi
    fi

    local wallet_id="${USER_MODE}-wallet-$(date +%s)"

    local json_payload="{
        \"walletId\": \"$wallet_id\",
        \"ownerPubKey\": \"${USER_NAME}_public_key\",
        \"ownerCertHash\": \"$CERT_HASH\",
        \"balances\": [0],
        \"digitalAssetTypes\": [$DIGITAL_ASSET_JSON]
    }"

    local output
    if output=$(run_fpcclient "invoke" "createWallet" "$json_payload" "Creating wallet for $USER_NAME"); then
        echo "$output"

        WALLET_UUID=$(extract_uuid "$output" "wallet")

        if [ -z "$WALLET_UUID" ]; then
            log_error "Failed to extract wallet UUID from response"
            return 1
        fi

        WALLET_ID="$wallet_id"
        WALLET_BALANCE=0
        ESCROW_BALANCE=0
        save_user_state
        log_activity "$USER_NAME" "Created wallet: ${WALLET_UUID}..."
        log_success "Wallet created successfully!"

        # Auto-mint 10000 CBDC
        log_info "Auto-funding wallet with 10000 CBDC..."
        if output=$(run_fpcclient "invoke" "mintTokens" "{
            \"assetId\": \"$DIGITAL_ASSET_UUID\",
            \"walletUUID\": \"$WALLET_UUID\",
            \"amount\": 10000,
            \"issuerCertHash\": \"sha256:central_bank_cert\"
        }" "Minting 10000 CBDC tokens"); then
            WALLET_BALANCE=10000
            save_user_state
            log_activity "$USER_NAME" "Auto-funded wallet with 10000 CBDC"
            log_success "Wallet funded with 10000 CBDC!"
        else
            echo "$output"
            log_error "Failed to auto-fund wallet"
        fi
    else
        echo "$output"
        log_error "Failed to create wallet"
        return 1
    fi
    # load_state
    #
    # if [ -n "$WALLET_UUID" ]; then
    #     log_info "Wallet already exists: ${WALLET_UUID}..."
    #     return 0
    # fi
    #
    # if [ -z "$DIGITAL_ASSET_JSON" ]; then
    #     log_error "Please setup the system first (Option 1 in Main Menu)!"
    #     return 1
    # fi
    #
    # local wallet_id="${USER_MODE}-wallet-$(date +%s)"
    #
    # # =======
    # # echo -e "${CYAN}Creating wallet with:${NC}"
    # # echo "  Wallet ID: $wallet_id"
    # # echo "  Owner: $USER_NAME"
    # # echo "  Cert Hash: $CERT_HASH"
    # # echo "  Digital Asset: ${DIGITAL_ASSET_UUID:0:12}..."
    # # echo "  DIGITAL_ASSET_JSON: $DIGITAL_ASSET_JSON"
    # local json_payload="{
    #     \"walletUUID\": \"$wallet_id\",
    #     \"ownerId\": \"$USER_NAME\",
    #     \"ownerCertHash\": \"$CERT_HASH\",
    #     \"balances\": [0],
    #     \"digitalAssetTypes\": [$DIGITAL_ASSET_JSON]
    # }"
    # #
    # # echo -e "${YELLOW}JSON Payload:${NC}"
    # # echo "$json_payload" | jq '.' 2>/dev/null || echo "$json_payload"
    # # echo
    # # ======
    #
    # local output
    # if output=$(run_fpcclient "invoke" "createWallet" "$json_payload" "Creating wallet for $USER_NAME"); then
    #     WALLET_UUID=$(extract_uuid "$output" "wallet")
    #
    #     if [ -z "$WALLET_UUID" ]; then
    #         log_error "Failed to extract wallet UUID from response"
    #         echo -e "${YELLOW}Response was:${NC}"
    #         echo "$output"
    #         return 1
    #     fi
    #
    #     WALLET_ID="$wallet_id"
    #     WALLET_BALANCE=0
    #     ESCROW_BALANCE=0
    #     save_user_state
    #     log_activity "$USER_NAME" "Created wallet: ${WALLET_UUID}..."
    #     log_success "Wallet created successfully!"
    #     echo "Wallet UUID: $WALLET_UUID"
    #     echo "Wallet ID: $WALLET_ID"
    # else
    #     log_error "Failed to create wallet"
    #     echo "$output"
    #     echo
    #     echo -e "${YELLOW}Troubleshooting:${NC}"
    #     echo "1. Make sure you ran 'Setup System' first (Option 1)"
    #     echo "2. Check if the network is running properly"
    #     echo "3. Verify your environment variables are correct"
    #     return 1
    # fi
}

check_balance() {
    load_state

    if [ -z "$WALLET_UUID" ]; then
        log_error "Please create a wallet first!"
        return 1
    fi

    local output
    if output=$(run_fpcclient "query" "getBalance" "{
        \"walletUUID\": \"$WALLET_UUID\",
        \"assetSymbol\": \"CBDC\",
        \"ownerCertHash\": \"$CERT_HASH\"
    }" "Checking balance for $USER_NAME"); then
        local balance=$(extract_balance "$output")
        WALLET_BALANCE=${balance:-0}
        save_user_state
        log_success "Current balance: $WALLET_BALANCE CBDC"
    else
        log_error "Failed to check balance"
        echo "$output"
        return 1
    fi
}

query_wallet() {
    load_state

    if [ -z "$WALLET_UUID" ]; then
        log_error "Please create a wallet first!"
        return 1
    fi

    run_fpcclient "query" "readWallet" "{\"uuid\": \"$WALLET_UUID\"}" "Querying wallet details"
}

mint_tokens() {
    load_state

    if [ -z "$WALLET_UUID" ] || [ -z "$DIGITAL_ASSET_UUID" ]; then
        log_error "Please create wallet and setup system first!"
        return 1
    fi

    echo -n "Enter amount to mint: "
    read amount

    if ! [[ "$amount" =~ ^[0-9]+$ ]] || [ "$amount" -le 0 ]; then
        log_error "Invalid amount. Please enter a positive number."
        return 1
    fi

    local output
    if output=$(run_fpcclient "invoke" "mintTokens" "{
        \"assetId\": \"$DIGITAL_ASSET_UUID\",
        \"walletUUID\": \"$WALLET_UUID\",
        \"amount\": $amount,
        \"issuerCertHash\": \"sha256:central_bank_cert\"
    }" "Minting $amount CBDC tokens"); then
        log_activity "$USER_NAME" "Minted $amount CBDC tokens"
        check_balance
        log_success "$amount tokens minted successfully!"
    else
        log_error "Failed to mint tokens"
        return 1
    fi
}

transfer_tokens() {
    load_state

    if [ -z "$WALLET_UUID" ]; then
        log_error "Please create a wallet first!"
        return 1
    fi

    # Get other user's wallet
    local other_user_state
    local other_user

    if [ "$USER_MODE" = "alice" ]; then
        other_user_state="$BOB_STATE"
        other_user="bob"
    else
        other_user_state="$ALICE_STATE"
        other_user="alice"
    fi

    if [ ! -f "$other_user_state" ] || [ ! -s "$other_user_state" ]; then
        log_error "$other_user hasn't created a wallet yet!"
        return 1
    fi

    # Save current state
    local my_wallet="$WALLET_UUID"

    # Load other user's wallet
    source "$other_user_state"
    local to_wallet="$WALLET_UUID"

    if [ -z "$to_wallet" ]; then
        log_error "$other_user's wallet UUID not found!"
        return 1
    fi

    # Reload our state
    source "${STATE_DIR}/${USER_MODE}.state"
    source "$SHARED_STATE"

    echo -n "Enter amount to transfer to $other_user: "
    read amount

    if ! [[ "$amount" =~ ^[0-9]+$ ]] || [ "$amount" -le 0 ]; then
        log_error "Invalid amount. Please enter a positive number."
        return 1
    fi

    local output
    if output=$(run_fpcclient "invoke" "transferTokens" "{
        \"fromWalletId\": \"$my_wallet\",
        \"toWalletId\": \"$to_wallet\",
        \"assetId\": \"$DIGITAL_ASSET_UUID\",
        \"amount\": $amount,
        \"senderCertHash\": \"$CERT_HASH\"
    }" "Transferring $amount CBDC to $other_user"); then
        log_activity "$USER_NAME" "Transferred $amount CBDC to $other_user"
        check_balance
        log_success "$amount tokens transferred successfully to $other_user!"
    else
        log_error "Failed to transfer tokens"
        return 1
    fi
}

burn_tokens() {
    load_state

    if [ -z "$WALLET_UUID" ] || [ -z "$DIGITAL_ASSET_UUID" ]; then
        log_error "Please create wallet first!"
        return 1
    fi

    echo -n "Enter amount to burn: "
    read amount

    if ! [[ "$amount" =~ ^[0-9]+$ ]] || [ "$amount" -le 0 ]; then
        log_error "Invalid amount. Please enter a positive number."
        return 1
    fi

    local output
    if output=$(run_fpcclient "invoke" "burnTokens" "{
        \"assetId\": \"$DIGITAL_ASSET_UUID\",
        \"walletUUID\": \"$WALLET_UUID\",
        \"amount\": $amount,
        \"issuerCertHash\": \"sha256:central_bank_cert\"
    }" "Burning $amount CBDC tokens"); then
        log_activity "$USER_NAME" "Burned $amount CBDC tokens"
        check_balance
        log_success "$amount tokens burned successfully!"
    else
        log_error "Failed to burn tokens"
        return 1
    fi
}

# Escrow operations
create_escrow() {
    load_state

    if [ -z "$WALLET_UUID" ] || [ -z "$DIGITAL_ASSET_JSON" ]; then
        log_error "Please create wallet and setup system first!"
        return 1
    fi

    local other_user
    local other_cert_hash

    if [ "$USER_MODE" = "alice" ]; then
        other_user="bob"
        other_cert_hash="sha256:bob_cert"
        other_user_state="$BOB_STATE"
    else
        other_user="alice"
        other_cert_hash="sha256:alice_cert"
        other_user_state="$ALICE_STATE"
    fi

    echo -n "Enter parcel ID: "
    read parcel_id

    if [ -z "$parcel_id" ]; then
        log_error "Parcel ID cannot be empty!"
        return 1
    fi

    echo -n "Enter escrow amount: "
    read amount

    if ! [[ "$amount" =~ ^[0-9]+$ ]] || [ "$amount" -le 0 ]; then
        log_error "Invalid amount. Please enter a positive number."
        return 1
    fi

    echo -n "Enter secret for escrow: "
    read -s secret
    echo

    if [ -z "$secret" ]; then
        log_error "Secret cannot be empty!"
        return 1
    fi

    local escrow_id="${USER_MODE}-escrow-$(date +%s)"
    local buyer_wallet="$WALLET_UUID"

    source "$other_user_state"
    local seller_wallet="$WALLET_UUID"

    source "${STATE_DIR}/${USER_MODE}.state"
    source "$SHARED_STATE"

    local output
    if output=$(run_fpcclient "invoke" "createAndLockEscrow" "{
        \"escrowId\": \"$escrow_id\",
        \"buyerPubKey\": \"${USER_MODE}_public_key\",
        \"sellerPubKey\": \"${other_user}_public_key\",
        \"amount\": $amount,
        \"assetType\": $DIGITAL_ASSET_JSON,
        \"parcelId\": \"$parcel_id\",
        \"secret\": \"$secret\",
        \"buyerCertHash\": \"$CERT_HASH\",
        \"buyerWalletUUID\": \"$buyer_wallet\",
        \"sellerWalletUUID\": \"$seller_wallet\"
    }" "Creating escrow for parcel $parcel_id"); then
        local escrow_uuid=$(extract_uuid "$output" "escrow")
        LAST_ESCROW_UUID="$escrow_uuid"
        LAST_ESCROW_SECRET="$secret"
        LAST_PARCEL_ID="$parcel_id"
        save_user_state
        log_activity "$USER_NAME" "Created escrow for parcel $parcel_id with $other_user: ${escrow_uuid}..."
        check_balance
        log_success "Escrow created and funds locked successfully!"
        echo "Escrow UUID: $escrow_uuid"
        echo "Parcel ID: $parcel_id"
        echo "Remember your secret!"
    else
        echo "$output"
        log_error "Failed to create escrow"
        return 1
    fi
    # load_state
    #
    # if [ -z "$WALLET_UUID" ] || [ -z "$DIGITAL_ASSET_JSON" ]; then
    #     log_error "Please create wallet and setup system first!"
    #     return 1
    # fi
    #
    # # Get other user's cert hash (for seller)
    # local other_user
    # local other_cert_hash
    #
    # if [ "$USER_MODE" = "alice" ]; then
    #     other_user="bob"
    #     other_cert_hash="sha256:bob_cert"
    #     other_user_state="$BOB_STATE"
    # else
    #     other_user="alice"
    #     other_cert_hash="sha256:alice_cert"
    #     other_user_state="$ALICE_STATE"
    # fi
    #
    # echo -n "Enter escrow amount: "
    # read amount
    #
    # if ! [[ "$amount" =~ ^[0-9]+$ ]] || [ "$amount" -le 0 ]; then
    #     log_error "Invalid amount. Please enter a positive number."
    #     return 1
    # fi
    #
    # echo -n "Enter secret for escrow condition: "
    # read -s secret
    # echo
    #
    # if [ -z "$secret" ]; then
    #     log_error "Secret cannot be empty!"
    #     return 1
    # fi
    #
    # local escrow_id="${USER_MODE}-escrow-$(date +%s)"
    # local condition_hash=$(echo -n "$secret" | sha256sum | cut -d' ' -f1)
    #
    # # Get seller's wallet UUID
    # source "$other_user_state"
    # local seller_wallet="$WALLET_UUID"
    #
    # local output
    # if output=$(run_fpcclient "invoke" "createAndLockEscrow" "{
    #     \"escrowId\": \"$escrow_id\",
    #     \"buyerPubKey\": \"${USER_MODE}_public_key\",
    #     \"sellerPubKey\": \"${other_user}_public_key\",
    #     \"amount\": $amount,
    #     \"assetType\": $DIGITAL_ASSET_JSON,
    #     \"conditionValue\": \"$condition_hash\",
    #     \"buyerCertHash\": \"$CERT_HASH\",
    #     \"buyerWalletId\": \"$WALLET_UUID\",
    #     \"sellerWalletId\": \"$seller_wallet\"
    # }" "Creating escrow with $other_user and locking $amount CBDC"); then
    #     # \"buyerWalletId\": \"$WALLET_UUID\",
    #     echo "$output"
    #     local escrow_uuid=$(extract_uuid "$output" "escrow")
    #     LAST_ESCROW_UUID="$escrow_uuid"
    #     LAST_ESCROW_SECRET="$secret"
    #     save_user_state
    #     log_activity "$USER_NAME" "Created escrow with $other_user, locked $amount CBDC: ${escrow_uuid}..."
    #     check_balance
    #     log_success "Escrow created and funds locked successfully!"
    #     echo "Escrow UUID: $escrow_uuid"
    #     echo "Seller: $other_user"
    #     echo "Remember your secret for verification!"
    # else
    #     echo "$output"
    #     log_error "Failed to create escrow"
    #     return 1
    # fi
}

lock_funds_in_escrow() {
    load_state

    if [ -z "$WALLET_UUID" ]; then
        log_error "Please create a wallet first!"
        return 1
    fi

    if [ -z "$LAST_ESCROW_UUID" ]; then
        echo -n "Enter escrow UUID to lock funds in: "
        read escrow_uuid
    else
        echo "Last created escrow: ${LAST_ESCROW_UUID}..."
        echo -n "Use this escrow? (y/n): "
        read use_last

        if [ "$use_last" = "y" ] || [ "$use_last" = "Y" ]; then
            escrow_uuid="$LAST_ESCROW_UUID"
        else
            echo -n "Enter escrow UUID: "
            read escrow_uuid
        fi
    fi

    echo -n "Enter amount to lock: "
    read amount

    if ! [[ "$amount" =~ ^[0-9]+$ ]] || [ "$amount" -le 0 ]; then
        log_error "Invalid amount. Please enter a positive number."
        return 1
    fi

    local output
    if output=$(run_fpcclient "invoke" "lockFundsInEscrow" "{
        \"escrowId\": \"$escrow_uuid\",
        \"payerWalletId\": \"$WALLET_UUID\",
        \"amount\": $amount,
        \"assetId\": \"$DIGITAL_ASSET_UUID\",
        \"payerCertHash\": \"$CERT_HASH\"
    }" "Locking $amount CBDC in escrow"); then
        log_activity "$USER_NAME" "Locked $amount CBDC in escrow ${escrow_uuid}..."
        check_balance
        log_success "Funds locked successfully in escrow!"
    else
        log_error "Failed to lock funds in escrow"
        return 1
    fi
}

verify_escrow_condition() {
    load_state

    if [ -z "$LAST_ESCROW_UUID" ]; then
        echo -n "Enter escrow UUID to verify: "
        read escrow_uuid
    else
        echo "Last escrow: ${LAST_ESCROW_UUID}..."
        echo -n "Use this escrow? (y/n): "
        read use_last

        if [ "$use_last" = "y" ] || [ "$use_last" = "Y" ]; then
            escrow_uuid="$LAST_ESCROW_UUID"
        else
            echo -n "Enter escrow UUID: "
            read escrow_uuid
        fi
    fi

    echo -n "Enter secret: "
    read -s secret
    echo

    echo -n "Enter parcel ID: "
    read parcel_id

    local output
    if output=$(run_fpcclient "invoke" "verifyEscrowCondition" "{
        \"escrowId\": \"$escrow_uuid\",
        \"secret\": \"$secret\",
        \"parcelId\": \"$parcel_id\"
    }" "Verifying escrow condition"); then
        log_activity "$USER_NAME" "Verified escrow condition for ${escrow_uuid}..."
        log_success "Escrow condition verified successfully!"
    else
        log_error "Failed to verify escrow condition"
        return 1
    fi
}

release_escrow() {
    load_state

    if [ -z "$WALLET_UUID" ]; then
        log_error "Please create a wallet first!"
        return 1
    fi

    echo -n "Enter escrow UUID to release: "
    read escrow_uuid

    echo -n "Enter parcel ID: "
    read parcel_id

    echo -n "Enter secret provided by buyer: "
    read -s secret
    echo

    local output
    if output=$(run_fpcclient "invoke" "releaseEscrow" "{
        \"escrowUUID\": \"$escrow_uuid\",
        \"secret\": \"$secret\",
        \"parcelId\": \"$parcel_id\",
        \"sellerCertHash\": \"$CERT_HASH\"
    }" "Releasing escrow"); then
        log_activity "$USER_NAME" "Released escrow ${escrow_uuid}..."
        check_balance
        log_success "Escrow released! Funds transferred to your wallet."
    else
        echo "$output"
        log_error "Failed to release escrow"
        return 1
    fi
    # load_state
    #
    # if [ -z "$WALLET_UUID" ]; then
    #     log_error "Please create a wallet first!"
    #     return 1
    # fi
    #
    # if [ -z "$LAST_ESCROW_UUID" ]; then
    #     echo -n "Enter escrow UUID to release: "
    #     read escrow_uuid
    # else
    #     echo "Last escrow: ${LAST_ESCROW_UUID}..."
    #     echo -n "Use this escrow? (y/n): "
    #     read use_last
    #
    #     if [ "$use_last" = "y" ] || [ "$use_last" = "Y" ]; then
    #         escrow_uuid="$LAST_ESCROW_UUID"
    #     else
    #         echo -n "Enter escrow UUID: "
    #         read escrow_uuid
    #     fi
    # fi
    #
    # local output
    # if output=$(run_fpcclient "invoke" "releaseEscrow" "{
    #     \"escrowId\": \"$escrow_uuid\",
    #     \"payerWalletId\": \"$WALLET_UUID\",
    # }" "Releasing escrow to $other_user"); then
    #     log_activity "$USER_NAME" "Released escrow ${escrow_uuid}... to $other_user"
    #     check_balance
    #     log_success "Escrow released successfully to $other_user!"
    # else
    #     log_error "Failed to release escrow"
    #     return 1
    # fi
}

refund_escrow() {
    load_state

    if [ -z "$WALLET_UUID" ]; then
        log_error "Please create a wallet first!"
        return 1
    fi

    if [ -z "$LAST_ESCROW_UUID" ]; then
        echo -n "Enter escrow UUID to refund: "
        read escrow_uuid
    else
        echo "Last escrow: ${LAST_ESCROW_UUID}"
        echo -n "Use this escrow? (y/n): "
        read use_last

        if [ "$use_last" = "y" ] || [ "$use_last" = "Y" ]; then
            escrow_uuid="$LAST_ESCROW_UUID"
        else
            echo -n "Enter escrow UUID: "
            read escrow_uuid
        fi
    fi

    local output
    if output=$(run_fpcclient "invoke" "refundEscrow" "{
        \"escrowUUID\": \"$escrow_uuid\",
        \"buyerWalletUUID\": \"$WALLET_UUID\",
        \"buyerCertHash\": \"$CERT_HASH\"
    }" "Refunding escrow"); then
        log_activity "$USER_NAME" "Refunded escrow ${escrow_uuid}..."
        check_balance
        log_success "Escrow refunded! Funds returned to your wallet."
    else
        echo "$output"
        log_error "Failed to refund escrow"
        return 1
    fi
    # load_state
    #
    # if [ -z "$WALLET_UUID" ]; then
    #     log_error "Please create a wallet first!"
    #     return 1
    # fi
    #
    # if [ -z "$LAST_ESCROW_UUID" ]; then
    #     echo -n "Enter escrow UUID to refund: "
    #     read escrow_uuid
    # else
    #     echo "Last escrow: ${LAST_ESCROW_UUID}..."
    #     echo -n "Use this escrow? (y/n): "
    #     read use_last
    #
    #     if [ "$use_last" = "y" ] || [ "$use_last" = "Y" ]; then
    #         escrow_uuid="$LAST_ESCROW_UUID"
    #     else
    #         echo -n "Enter escrow UUID: "
    #         read escrow_uuid
    #     fi
    # fi
    #
    # local output
    # if output=$(run_fpcclient "invoke" "refundEscrow" "{
    #     \"escrowId\": \"$escrow_uuid\",
    #     \"payerWalletId\": \"$WALLET_UUID\"
    # }" "Refunding escrow"); then
    #     log_activity "$USER_NAME" "Refunded escrow ${escrow_uuid}..."
    #     check_balance
    #     log_success "Escrow refunded successfully!"
    # else
    #     log_error "Failed to refund escrow"
    #     return 1
    # fi
}

check_escrow_balance() {
    load_state

    if [ -z "$WALLET_UUID" ]; then
        log_error "Please create a wallet first!"
        return 1
    fi

    local output
    if output=$(run_fpcclient "query" "getEscrowBalance" "{
        \"walletUUID\": \"$WALLET_UUID\",
        \"assetSymbol\": \"CBDC\",
        \"ownerCertHash\": \"$CERT_HASH\"
    }" "Checking escrow balance"); then
        echo "$output"
        local balance=$(extract_escrow_balance "$output")
        ESCROW_BALANCE=${balance:-0}
        save_user_state
        log_success "Current escrow balance: $ESCROW_BALANCE CBDC"
    else
        log_error "Failed to check escrow balance"
        return 1
    fi
}

query_escrow() {
    load_state

    if [ -z "$LAST_ESCROW_UUID" ]; then
        echo -n "Enter escrow UUID to query: "
        read escrow_uuid
    else
        echo "Last escrow: ${LAST_ESCROW_UUID}..."
        echo -n "Use this escrow? (y/n): "
        read use_last

        if [ "$use_last" = "y" ] || [ "$use_last" = "Y" ]; then
            escrow_uuid="$LAST_ESCROW_UUID"
        else
            echo -n "Enter escrow UUID: "
            read escrow_uuid
        fi
    fi

    run_fpcclient "query" "readEscrow" "{\"uuid\": \"$escrow_uuid\"}" "Querying escrow details"
}

# Query operations
query_digital_asset() {
    load_state

    if [ -z "$DIGITAL_ASSET_UUID" ]; then
        log_error "Digital asset not created yet!"
        return 1
    fi

    run_fpcclient "query" "readDigitalAsset" "{\"uuid\": \"$DIGITAL_ASSET_UUID\"}" "Querying digital asset"
}

get_schema() {
    run_fpcclient "invoke" "getSchema" '{}' "Getting system schema"
}

get_wallet_by_owner() {
    load_state

    if [ -z "$WALLET_UUID" ]; then
        log_error "Please create a wallet first!"
        return 1
    fi

    run_fpcclient "query" "getWalletByOwner" "{
        \"walletUuid\": \"$WALLET_UUID\",
        \"ownerCertHash\": \"$CERT_HASH\"
    }" "Getting wallet by owner"
}

# Monitor operations
run_monitor() {
    local last_activity_hash=""
    local last_alice_hash=""
    local last_bob_hash=""

    while true; do
        # Calculate hashes of current state
        local current_activity_hash=$([ -f "$ACTIVITY_LOG" ] && md5sum "$ACTIVITY_LOG" | cut -d' ' -f1 || echo "")
        local current_alice_hash=$([ -f "$ALICE_STATE" ] && md5sum "$ALICE_STATE" | cut -d' ' -f1 || echo "")
        local current_bob_hash=$([ -f "$BOB_STATE" ] && md5sum "$BOB_STATE" | cut -d' ' -f1 || echo "")

        # Check if anything changed
        if [ "$current_activity_hash" != "$last_activity_hash" ] ||
            [ "$current_alice_hash" != "$last_alice_hash" ] ||
            [ "$current_bob_hash" != "$last_bob_hash" ]; then

            show_dashboard
            echo
            echo -e "${YELLOW}[Auto-refreshing... Press Ctrl+C to exit]${NC}"

            # Update hashes
            last_activity_hash="$current_activity_hash"
            last_alice_hash="$current_alice_hash"
            last_bob_hash="$current_bob_hash"
        fi

        sleep 2 # Refresh every 2 seconds
    done
    # while true; do
    #     show_dashboard
    #     show_monitor_menu
    #     read -p "Choose option: " choice
    #
    #     case $choice in
    #     1)
    #         sleep 1
    #         continue
    #         ;;
    #     2) show_full_activity_log ;;
    #     3) clear_activity_log ;;
    #     4) export_system_state ;;
    #     5) view_user_state "alice" ;;
    #     6) view_user_state "bob" ;;
    #     0) exit 0 ;;
    #     *) log_error "Invalid option" ;;
    #     esac
    #
    #     if [ $choice -ne 1 ]; then
    #         read -p "Press Enter to continue..."
    #     fi
    # done
}

show_full_activity_log() {
    clear
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}                    FULL ACTIVITY LOG                      ${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
    echo

    if [ -f "$ACTIVITY_LOG" ] && [ -s "$ACTIVITY_LOG" ]; then
        cat "$ACTIVITY_LOG"
    else
        echo "No activity logged yet."
    fi
    echo
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
}

clear_activity_log() {
    echo -n "Are you sure you want to clear the activity log? (y/N): "
    read confirm

    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        >"$ACTIVITY_LOG"
        log_success "Activity log cleared!"
    else
        log_info "Activity log not cleared."
    fi
}

export_system_state() {
    local export_file="/tmp/fpc_system_export_$(date +%Y%m%d_%H%M%S).txt"

    {
        echo "=========================================="
        echo "     FPC System State Export"
        echo "=========================================="
        echo "Export Time: $(date)"
        echo
        echo "=========================================="
        echo "         Shared State"
        echo "=========================================="
        if [ -f "$SHARED_STATE" ]; then
            cat "$SHARED_STATE"
        else
            echo "No shared state"
        fi
        echo
        echo "=========================================="
        echo "         Alice State"
        echo "=========================================="
        if [ -f "$ALICE_STATE" ] && [ -s "$ALICE_STATE" ]; then
            cat "$ALICE_STATE"
        else
            echo "Alice not active"
        fi
        echo
        echo "=========================================="
        echo "         Bob State"
        echo "=========================================="
        if [ -f "$BOB_STATE" ] && [ -s "$BOB_STATE" ]; then
            cat "$BOB_STATE"
        else
            echo "Bob not active"
        fi
        echo
        echo "=========================================="
        echo "         Activity Log"
        echo "=========================================="
        if [ -f "$ACTIVITY_LOG" ] && [ -s "$ACTIVITY_LOG" ]; then
            cat "$ACTIVITY_LOG"
        else
            echo "No activity logged"
        fi
        echo
        echo "=========================================="
    } >"$export_file"

    log_success "System state exported to: $export_file"
}

view_user_state() {
    local user="$1"
    local state_file="${STATE_DIR}/${user}.state"

    clear
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}                    ${user^^} STATE                           ${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
    echo

    if [ -f "$state_file" ] && [ -s "$state_file" ]; then
        cat "$state_file"
    else
        echo "$user is not active yet."
    fi

    echo
    echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
}

# User interaction loops
run_user_interface() {
    while true; do
        show_dashboard
        show_main_menu
        read -p "Choose option: " choice

        case $choice in
        # 1) setup_system ;;
        1) handle_wallet_operations ;;
        2) handle_token_operations ;;
        3) handle_escrow_operations ;;
        4) handle_query_operations ;;
        5)
            sleep 1
            continue
            ;;
        6) show_full_activity_log ;;
        0) exit 0 ;;
        *) log_error "Invalid option" ;;
        esac

        if [ $choice -ne 6 ]; then
            read -p "Press Enter to continue..."
        fi
    done
}

handle_wallet_operations() {
    while true; do
        show_wallet_menu
        read -p "Choose option: " choice

        case $choice in
        1) create_wallet ;;
        2) check_balance ;;
        3) query_wallet ;;
        0) return ;;
        *) log_error "Invalid option" ;;
        esac

        if [ $choice -ne 0 ]; then
            read -p "Press Enter to continue..."
        fi
    done
}

handle_token_operations() {
    while true; do
        show_token_menu
        read -p "Choose option: " choice

        case $choice in
        1) mint_tokens ;;
        2) transfer_tokens ;;
        3) burn_tokens ;;
        0) return ;;
        *) log_error "Invalid option" ;;
        esac

        if [ $choice -ne 0 ]; then
            read -p "Press Enter to continue..."
        fi
    done
}

handle_escrow_operations() {
    while true; do
        show_escrow_menu
        read -p "Choose option: " choice

        case $choice in
        1) create_escrow ;;
        2) verify_escrow_condition ;;
        3) release_escrow ;;
        4) refund_escrow ;;
        5) check_escrow_balance ;;
        6) query_escrow ;;
        0) return ;;
        *) log_error "Invalid option" ;;
        esac

        if [ $choice -ne 0 ]; then
            read -p "Press Enter to continue..."
        fi
    done
}

handle_query_operations() {
    while true; do
        show_query_menu
        read -p "Choose option: " choice

        case $choice in
        1) query_wallet ;;
        2) query_digital_asset ;;
        3) get_schema ;;
        4) check_balance ;;
        5) get_wallet_by_owner ;;
        0) return ;;
        *) log_error "Invalid option" ;;
        esac

        if [ $choice -ne 0 ]; then
            read -p "Press Enter to continue..."
        fi
    done
}

# Main Execution
show_startup_menu() {
    clear
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}                 ${YELLOW}FPC SETUP & TEST SYSTEM${NC}                       ${CYAN}║${NC}"
    echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"

    # Dynamic environment display with proper spacing
    local env_text
    if [ "$RUNNING_IN_DOCKER" = "true" ]; then
        env_text="Docker Container"
    else
        env_text="Host System     " # Padded to match "Docker Container" length
    fi
    echo -e "${CYAN}║${NC}  Environment: ${env_text}                          ${CYAN}║${NC}"

    echo -e "${CYAN}╠══════════════════════════════════════════════════════════════╣${NC}"
    echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  ${YELLOW}SETUP OPTIONS:${NC}                                              ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  1. Full Setup (ERCC + Network + Install)                   ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  2. Quick Setup (Skip ERCC build)                           ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  ${YELLOW}MULTI-USER TEST:${NC}                                           ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  3. Run as Alice (Org1MSP)                                  ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  4. Run as Bob (Org2MSP)                                    ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  5. Run as Monitor (Read-only)                              ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  ${YELLOW}UTILITIES:${NC}                                                 ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  6. Reset System (Clear all state)                          ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  7. Run All Tests (Original test script)                    ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}  0. Exit                                                    ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}                                                              ${CYAN}║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
}

reset_system() {
    echo -n "Are you sure you want to reset the entire system? (y/N): "
    read confirm

    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        rm -rf "$STATE_DIR"
        rm -rf "tmp/fpc_enclave_initialized"
        # rm -f "/tmp/fpc_enclave_alice_initialized"
        # rm -f "/tmp/fpc_enclave_bob_initialized"
        init_state
        log_success "System reset complete!"
        sleep 2
    else
        log_info "Reset cancelled."
        sleep 1
    fi
}

###########
# MAIN.SH #
###########

# Source main.sh for setup functions
source_main_script() {
    local main_script="$FPC_PATH/samples/chaincode/confidential-escrow/main.sh"

    if [ -f "$main_script" ]; then
        # Source only the functions, don't execute main
        source "$main_script"
        log_info "Loaded setup functions from main.sh"
    else
        log_error "main.sh not found at: $main_script"
        exit 1
    fi
}

# Detect if running inside Docker container
detect_environment() {
    if [ -f "/.dockerenv" ] || grep -q docker /proc/1/cgroup 2>/dev/null; then
        RUNNING_IN_DOCKER="true"
        log_info "Detected: Running inside Docker container"
    else
        RUNNING_IN_DOCKER="false"
        log_info "Detected: Running on host system"
    fi
}

#########
# SETUP #
#########

# Wrapper functions that call main.sh functions
do_full_setup() {
    if [ "$MAIN_SCRIPT_SOURCED" = "false" ]; then
        source_main_script
        MAIN_SCRIPT_SOURCED="true"
    fi

    log_info "=== RUNNING FULL SETUP ==="
    build_ercc
    build_chaincode
    initial_setup
    setup_network
    install_fpc
    start_ercc
    log_success "Full setup completed!"
}

do_quick_setup() {
    if [ "$MAIN_SCRIPT_SOURCED" = "false" ]; then
        source_main_script
        MAIN_SCRIPT_SOURCED="true"
    fi

    log_info "=== RUNNING QUICK SETUP ==="
    build_chaincode
    setup_network
    install_fpc
    start_ercc
    log_success "Quick setup completed!"
}

run_original_tests() {
    if [ "$MAIN_SCRIPT_SOURCED" = "false" ]; then
        source_main_script
        MAIN_SCRIPT_SOURCED="true"
    fi

    log_info "=== RUNNING ORIGINAL TEST SUITE ==="
    run_tests
    log_success "Original tests completed!"
}

main() {
    check_fpc_path
    detect_environment
    init_state

    # If argument provided, use it directly
    if [ $# -gt 0 ]; then
        case "$1" in
        "full")
            source_main_script
            MAIN_SCRIPT_SOURCED="true"
            do_full_setup
            ;;
        "quick")
            source_main_script
            MAIN_SCRIPT_SOURCED="true"
            do_quick_setup
            ;;
        "test-all")
            source_main_script
            MAIN_SCRIPT_SOURCED="true"
            run_original_tests
            ;;
        "alice" | "bob")
            setup_user_env "$1"
            run_user_interface
            ;;
        "monitor")
            USER_MODE="monitor"
            run_monitor
            ;;
        "reset")
            reset_system
            ;;
        *)
            echo "Usage: $0 [full|quick|docker-env|alice|bob|monitor|reset|test-all]"
            echo
            echo "Setup Options:"
            echo "  full       - Full setup (ERCC + Network + Install)"
            echo "  quick      - Quick setup (Skip ERCC build)"
            echo
            echo "Multi-User Options:"
            echo "  alice      - Run as Alice (Org1MSP)"
            echo "  bob        - Run as Bob (Org2MSP)"
            echo "  monitor    - Run as Monitor (read-only)"
            echo
            echo "Utilities:"
            echo "  reset      - Reset system state"
            echo "  test-all   - Run original test suite"
            exit 1
            ;;
        esac
        return
    fi

    # Interactive mode with integrated menu
    while true; do
        show_startup_menu
        read -p "Choose option (0-7): " choice

        case $choice in
        1 | 2 | 7)
            # Source main.sh once for setup operations
            if [ "$MAIN_SCRIPT_SOURCED" = "false" ]; then
                source_main_script
                MAIN_SCRIPT_SOURCED="true"
            fi
            ;;
        esac

        case $choice in
        1)
            do_full_setup
            read -p "Press Enter to continue..."
            ;;
        2)
            do_quick_setup
            read -p "Press Enter to continue..."
            ;;
        3)
            setup_user_env "alice"
            run_user_interface
            ;;
        4)
            setup_user_env "bob"
            run_user_interface
            ;;
        5)
            USER_MODE="monitor"
            run_monitor
            ;;
        6)
            reset_system
            ;;
        7)
            run_original_tests
            read -p "Press Enter to continue..."
            ;;
        0)
            exit 0
            ;;
        *)
            log_error "Invalid option"
            sleep 1
            ;;
        esac
    done
}

main "$@"

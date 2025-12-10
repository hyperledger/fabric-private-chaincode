package chaincode

import (
	"github.com/hyperledger-labs/cc-tools/assets"
	tx "github.com/hyperledger-labs/cc-tools/transactions"

	asset "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/assets"
	header "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/header"
	transaction "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/transactions"
)

var (
	// TxList defines all transaction handlers available in the chaincode.
	// Transactions are grouped by functionality:
	//   - Create operations: Initialize new assets (wallets, tokens, escrows)
	//   - Read operations: Query existing assets
	//   - Balance operations: Query and verify account balances
	//   - Token operations: Mint, transfer, and burn digital assets
	//   - Escrow operations: Manage conditional fund transfers
	TxList = []tx.Transaction{
		transaction.DebugTest,
		// Create
		transaction.CreateUserDir,
		transaction.CreateWallet,
		transaction.CreateDigitalAsset,
		transaction.CreateAndLockEscrow,
		// Read
		transaction.ReadUserDir,
		transaction.ReadDigitalAsset,
		transaction.ReadEscrow,
		// misc
		transaction.GetBalance,
		transaction.GetEscrowBalance,
		transaction.GetWalletByOwner,
		transaction.MintTokens,
		transaction.TransferTokens,
		transaction.BurnTokens,
		// Escrow
		transaction.RefundEscrow,
		transaction.VerifyEscrowCondition,
		transaction.ReleaseEscrow,
	}

	// AssetTypeList defines all asset types managed by the chaincode.
	// Each asset type represents a core data structure:
	//   - Wallet: User accounts holding digital assets
	//   - DigitalAssetToken: Fungible tokens (e.g., CBDC)
	//   - UserDirectory: Mapping between public keys and wallets
	//   - Escrow: Conditional payment contracts
	AssetTypeList = []assets.AssetType{
		asset.Wallet,
		asset.DigitalAssetToken,
		asset.UserDirectory,
		asset.Escrow,
	}
)

// SetupCC initializes the chaincode with all necessary components.
// This function configures the chaincode header metadata and registers
// all transaction handlers, asset types, and event definitions with cc-tools.
//
// Returns:
//
//	error: Error if initialization fails, nil on success
func SetupCC() error {
	// Initialize header info
	tx.InitHeader(tx.Header{
		Name:    header.Name,
		Version: header.Version,
		Colors:  header.Colors,
		Title:   header.Title,
	})

	// Initialize transaction and asset lists
	tx.InitTxList(TxList)
	assets.InitAssetList(AssetTypeList)

	return nil
}

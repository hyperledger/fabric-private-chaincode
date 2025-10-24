package chaincode

import (
	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/events"
	tx "github.com/hyperledger-labs/cc-tools/transactions"

	asset "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/assets"
	header "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/header"
	transaction "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/transactions"
)

// Transaction and asset lists
var (
	TxList = []tx.Transaction{
		transaction.DebugTest,
		// Create
		transaction.CreateUserDir,
		transaction.CreateWallet,
		transaction.CreateDigitalAsset,
		transaction.CreateAndLockEscrow,
		// Read
		transaction.ReadUserDir,
		transaction.ReadWallet,
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
	AssetTypeList = []assets.AssetType{
		asset.Wallet,
		asset.DigitalAssetToken,
		asset.UserDirectory,
		asset.Escrow,
	}
	EventTypeList = []events.Event{} // Empty for now
)

// SetupCC initializes the chaincode with assets and transactions
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
	events.InitEventList(EventTypeList)

	return nil
}

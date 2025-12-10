package testutils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// TestFixtures provides common test data for unit tests
type TestFixtures struct {
	// User credentials
	BuyerPubKey     string
	BuyerCertHash   string
	BuyerPubKeyHash string
	BuyerWalletID   string
	BuyerWalletUUID string

	SellerPubKey     string
	SellerCertHash   string
	SellerPubKeyHash string
	SellerWalletID   string
	SellerWalletUUID string

	IssuerCertHash string

	// Asset data
	AssetID     string
	AssetSymbol string
	AssetName   string

	// Escrow data
	EscrowID string
	ParcelID string
	Secret   string
	Amount   float64
}

// NewTestFixtures creates a standard set of test data
func NewTestFixtures() *TestFixtures {
	buyerPubKey := "buyer-public-key-123"
	sellerPubKey := "seller-public-key-456"

	buyerHash := sha256.Sum256([]byte(buyerPubKey))
	buyerPubKeyHash := hex.EncodeToString(buyerHash[:])

	sellerHash := sha256.Sum256([]byte(sellerPubKey))
	sellerPubKeyHash := hex.EncodeToString(sellerHash[:])

	return &TestFixtures{
		BuyerPubKey:      buyerPubKey,
		BuyerCertHash:    "buyer-cert-hash",
		BuyerPubKeyHash:  buyerPubKeyHash,
		BuyerWalletID:    "buyer-wallet-id",
		BuyerWalletUUID:  "buyer-wallet-uuid",
		SellerPubKey:     sellerPubKey,
		SellerCertHash:   "seller-cert-hash",
		SellerPubKeyHash: sellerPubKeyHash,
		SellerWalletID:   "seller-wallet-id",
		SellerWalletUUID: "seller-wallet-uuid",
		IssuerCertHash:   "issuer-cert-hash",
		AssetID:          "test-asset-id",
		AssetSymbol:      "TST",
		AssetName:        "Test Token",
		EscrowID:         "test-escrow-id",
		ParcelID:         "parcel-123",
		Secret:           "secret-key",
		Amount:           100.0,
	}
}

// CreateMockWallet creates a wallet asset in the mock state
// walletID is the user-provided nickname
// walletUUID is the cc-tools generated unique identifier
func (f *TestFixtures) CreateMockWallet(mockStub *MockStub, pubKey, certHash, walletID, walletUUID string, assetID string, balance, escrowBalance float64) error {
	fmt.Printf("\nDEBUG CreateMockWallet called:\n")
	fmt.Printf("DEBUG   pubKey: %q\n", pubKey)
	fmt.Printf("DEBUG   walletID: %q\n", walletID)
	fmt.Printf("DEBUG   walletUUID: %q\n", walletUUID)

	walletMap := map[string]any{
		"@assetType":     "wallet",
		"@key":           "wallet:" + walletUUID,
		"walletId":       walletID,
		"ownerPubKey":    pubKey,
		"ownerCertHash":  certHash,
		"balances":       []any{balance},
		"escrowBalances": []any{escrowBalance},
		"digitalAssetTypes": []any{
			map[string]any{
				"@key": "digitalAsset:" + assetID,
			},
		},
		"createdAt": time.Now(),
	}

	walletJSON, err := json.Marshal(walletMap)
	if err != nil {
		return err
	}

	fmt.Printf("DEBUG   Storing with key: %q\n", "wallet:"+walletUUID)
	return mockStub.PutState("wallet:"+walletUUID, walletJSON)
}

// func (f *TestFixtures) CreateMockWallet(mockStub *MockStub, pubKey, certHash, walletID, walletUUID string, assetID string, balance, escrowBalance float64) error {
// 	walletMap := map[string]any{
// 		"@assetType":     "wallet",
// 		"@key":           "wallet:" + walletUUID, // CC-tools composite key
// 		"walletId":       walletID,               // User-provided nickname
// 		"ownerPubKey":    pubKey,
// 		"ownerCertHash":  certHash,
// 		"balances":       []any{balance},
// 		"escrowBalances": []any{escrowBalance},
// 		"digitalAssetTypes": []any{
// 			map[string]any{
// 				"@key": "digitalAsset:" + assetID,
// 			},
// 		},
// 		"createdAt": time.Now(),
// 	}
//
// 	walletJSON, err := json.Marshal(walletMap)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Store by UUID (the actual ledger key)
// 	return mockStub.PutState("wallet:"+walletUUID, walletJSON)
// }

// CreateMockUserDir creates a user directory entry in the mock state
// The UserDirectory maps publicKeyHash -> walletUUID (NOT walletID)
func (f *TestFixtures) CreateMockUserDir(mockStub *MockStub, pubKeyHash, walletUUID, certHash string) error {
	// Generate UUID for @key
	hash := sha256.Sum256([]byte("userdir" + pubKeyHash))
	uuidStr := fmt.Sprintf("%x-%x-%x-%x-%x",
		hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
	uuidKey := "userdir:" + uuidStr

	userDirMap := map[string]any{
		"@assetType":    "userdir",
		"@key":          uuidKey,
		"publicKeyHash": pubKeyHash,
		"walletUUID":    walletUUID, // References the UUID, not the ID
		"certHash":      certHash,
	}

	userDirJSON, err := json.Marshal(userDirMap)
	if err != nil {
		return err
	}

	// Store with UUID key - PutState will auto-create composite index based on registry
	return mockStub.PutState(uuidKey, userDirJSON)
}

// CreateMockDigitalAsset creates a digital asset in the mock state
// assetID is cc-tools generated UUID
// symbol is user-provided unique identifier
func (f *TestFixtures) CreateMockDigitalAsset(mockStub *MockStub, assetID, symbol, name, issuerHash string, totalSupply float64) error {
	assetMap := map[string]any{
		"@assetType":  "digitalAsset",
		"@key":        "digitalAsset:" + assetID, // CC-tools composite key uses UUID
		"name":        name,
		"symbol":      symbol, // This is the unique key property (IsKey: true)
		"decimals":    2.0,
		"totalSupply": totalSupply,
		"issuerHash":  issuerHash,
		"owner":       "test-owner",
		"issuedAt":    time.Now(),
	}

	assetJSON, err := json.Marshal(assetMap)
	if err != nil {
		return err
	}

	return mockStub.PutState("digitalAsset:"+assetID, assetJSON)
}

// CreateMockEscrow creates an escrow contract in the mock state
// escrowID is user-provided unique identifier (IsKey: true)
func (f *TestFixtures) CreateMockEscrow(
	mockStub *MockStub,
	escrowID string,
	buyerPubKey string,
	sellerPubKey string,
	buyerWalletUUID string,
	sellerWalletUUID string,
	assetID string,
	amount float64,
	parcelID string,
	secret string,
	status string,
	buyerCertHash string,
) error {
	// Compute condition hash: SHA256(secret + parcelId)
	conditionData := secret + parcelID
	conditionHash := sha256.Sum256([]byte(conditionData))
	conditionValue := hex.EncodeToString(conditionHash[:])

	escrowMap := map[string]any{
		"@assetType":       "escrow",
		"@key":             "escrow:" + escrowID,
		"escrowId":         escrowID,
		"buyerPubKey":      buyerPubKey,
		"sellerPubKey":     sellerPubKey,
		"buyerWalletUUID":  buyerWalletUUID,
		"sellerWalletUUID": sellerWalletUUID,
		"amount":           amount,
		"assetType": map[string]any{
			"@key": "digitalAsset:" + assetID,
		},
		"parcelId":       parcelID,
		"conditionValue": conditionValue, // Computed from secret + parcelID
		"status":         status,
		"createdAt":      time.Now(),
		"buyerCertHash":  buyerCertHash,
	}

	escrowJSON, err := json.Marshal(escrowMap)
	if err != nil {
		return err
	}

	return mockStub.PutState("escrow:"+escrowID, escrowJSON)
}

// ComputeConditionHash computes the escrow condition hash
func (f *TestFixtures) ComputeConditionHash(secret, parcelID string) string {
	hash := sha256.Sum256([]byte(secret + parcelID))
	return hex.EncodeToString(hash[:])
}

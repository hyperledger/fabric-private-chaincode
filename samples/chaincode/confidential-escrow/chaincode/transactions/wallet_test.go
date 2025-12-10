package transactions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
	asset "github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/assets"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/testutils"
)

// TestMain runs before all tests to initialize cc-tools
func TestMain(m *testing.M) {
	// Initialize asset types
	assetTypeList := []assets.AssetType{
		asset.Wallet,
		asset.DigitalAssetToken,
		asset.UserDirectory,
		asset.Escrow,
	}
	assets.InitAssetList(assetTypeList)

	// Run tests
	m.Run()
}

func asICCError(err errors.ICCError) errors.ICCError {
	return err
}

// ============================================================================
// CreateWallet Tests
// ============================================================================

func TestCreateWallet_Success(t *testing.T) {
	// Create a fresh mock blockchain stub for this test
	wrapper, mockStub := testutils.NewMockStubWrapper()

	// Define the input arguments for creating a wallet
	args := map[string]any{
		"walletId":      "alice-savings",          // User-friendly nickname
		"ownerPubKey":   "alice-public-key-12345", // Alice's public key
		"ownerCertHash": "alice-cert-hash-xyz",    // Alice's certificate hash
	}

	// Execute the CreateWallet transaction
	response, err := CreateWallet.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, err, "wallet creation should succeed")

	// Parse the response to check wallet properties
	var createdWallet map[string]any
	if err := json.Unmarshal(response, &createdWallet); err != nil {
		t.Fatalf("Failed to parse wallet response: %v", err)
	}

	// Verify the nickname was stored correctly
	testutils.AssertEqual(t, "alice-savings", createdWallet["walletId"], "walletId mismatch")

	// Verify the public key was stored correctly
	testutils.AssertEqual(t, "alice-public-key-12345", createdWallet["ownerPubKey"], "owner Pub key mismatch")

	// Verify the certificate hash was stored correctly
	testutils.AssertEqual(t, "alice-cert-hash-xyz", createdWallet["ownerCertHash"], "cert hash mismatch")

	// Verify that a UUID was generated (in the @key field)
	walletKey, exists := createdWallet["@key"].(string)
	if !exists {
		t.Fatal("Expected wallet to have a @key field with UUID")
	}

	// The key should be in format "wallet:<UUID>"
	if len(walletKey) < 8 || walletKey[:7] != "wallet:" {
		t.Errorf("Expected key format 'wallet:<UUID>', got: %s", walletKey)
	}

	// Extract the UUID portion (everything after "wallet:")
	walletUUID := walletKey[7:]

	// Verify the wallet was actually saved to the mock ledger
	_, exists = mockStub.State[walletKey]
	if !exists {
		t.Errorf("Expected wallet to be saved with key '%s'", walletKey)
	}

	var userDirBytes []byte
	var userDirKey string
	for key, value := range mockStub.State {
		if strings.HasPrefix(key, "userdir:") {
			userDirKey = key
			userDirBytes = value
			break
		}
	}

	if userDirBytes == nil {
		t.Fatalf("Expected UserDirectory entry to exist: %v", userDirKey)
	}

	var userDir map[string]any
	if err := json.Unmarshal(userDirBytes, &userDir); err != nil {
		t.Fatalf("Failed to parse UserDirectory: %v", err)
	}

	// Verify it has the correct publicKeyHash property
	hash := sha256.Sum256([]byte("alice-public-key-12345"))
	expectedPubKeyHash := hex.EncodeToString(hash[:])
	testutils.AssertEqual(t, expectedPubKeyHash, userDir["publicKeyHash"], "publicKeyHash mismatch")

	testutils.AssertEqual(t, walletUUID, userDir["walletUUID"], "UserDir has different walletUUID")

	// Verify UserDirectory contains the correct wallet UUID
	if err := json.Unmarshal(userDirBytes, &userDir); err != nil {
		t.Fatalf("Failed to parse UserDirectory: %v", err)
	}

	testutils.AssertEqual(t, walletUUID, userDir["walletUUID"], "UserDir has different walletUUID")

	// Verify empty balance arrays were initialized
	balances, ok := createdWallet["balances"].([]any)
	if !ok || len(balances) != 0 {
		t.Errorf("Expected empty balances array, got: %v", createdWallet["balances"])
	}

	escrowBalances, ok := createdWallet["escrowBalances"].([]any)
	if !ok || len(escrowBalances) != 0 {
		t.Errorf("Expected empty escrowBalances array, got: %v", createdWallet["escrowBalances"])
	}

	t.Log("✓ Wallet created successfully with all expected properties")
}

// ============================================================================
// GetBalance Tests
// ============================================================================

func TestGetBalance_Success(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup: Create digital asset
	err := fixtures.CreateMockDigitalAsset(
		mockStub,
		fixtures.AssetID,
		fixtures.AssetSymbol,
		fixtures.AssetName,
		fixtures.IssuerCertHash,
		1000.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock digital asset: %v", err)
	}

	// Setup: Create buyer wallet with balance
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		150.0, // balance
		50.0,  // escrow balance
	)
	if err != nil {
		t.Fatalf("Failed to create mock wallet: %v", err)
	}

	// Setup: Create user directory
	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock user directory: %v", err)
	}

	// t.Log("Mock state keys:")
	// for key := range mockStub.State {
	// 	t.Logf("  %q", key)
	// }

	// Execute GetBalance
	args := map[string]any{
		"pubKey":        fixtures.BuyerPubKey,
		"assetSymbol":   fixtures.AssetSymbol,
		"ownerCertHash": fixtures.BuyerCertHash,
	}

	response, txErr := GetBalance.Routine(wrapper.StubWrapper, args)
	if txErr != nil {
		t.Fatalf("GetBalance should succeed: Expected no error, got: %v", txErr)
	}

	// Verify response
	var result map[string]any
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	testutils.AssertEqual(t, fixtures.BuyerWalletUUID, result["walletId"], "walletId mismatch")
	testutils.AssertEqual(t, fixtures.AssetSymbol, result["assetSymbol"], "assetSymbol mismatch")
	testutils.AssertEqual(t, 150.0, result["balance"], "balance mismatch")

	t.Log("✓ GetBalance returned correct balance")
}

func TestGetBalance_WalletNotFound(t *testing.T) {
	wrapper, _ := testutils.NewMockStubWrapper()

	args := map[string]any{
		"pubKey":        "non-existent-pubkey",
		"assetSymbol":   "TST",
		"ownerCertHash": "some-cert-hash",
	}

	_, err := GetBalance.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, err, "should fail when wallet not found")
	testutils.AssertErrorStatus(t, err, 404, "should return 404 status")
}

func TestGetBalance_UnauthorizedAccess(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup wallet and user directory
	err := fixtures.CreateMockDigitalAsset(
		mockStub,
		fixtures.AssetID,
		fixtures.AssetSymbol,
		fixtures.AssetName,
		fixtures.IssuerCertHash,
		1000.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock digital asset: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		150.0,
		50.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock wallet: %v", err)
	}

	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock user directory: %v", err)
	}

	// Try to access with wrong certificate hash
	args := map[string]any{
		"pubKey":        fixtures.BuyerPubKey,
		"assetSymbol":   fixtures.AssetSymbol,
		"ownerCertHash": "wrong-cert-hash", // Wrong certificate
	}

	_, txErr := GetBalance.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong certificate")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

func TestGetBalance_AssetNotInWallet(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup wallet with one asset
	err := fixtures.CreateMockDigitalAsset(
		mockStub,
		fixtures.AssetID,
		fixtures.AssetSymbol,
		fixtures.AssetName,
		fixtures.IssuerCertHash,
		1000.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock digital asset: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		150.0,
		50.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock wallet: %v", err)
	}

	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock user directory: %v", err)
	}

	// Query for a different asset
	args := map[string]any{
		"pubKey":        fixtures.BuyerPubKey,
		"assetSymbol":   "NOTFOUND", // Asset not in wallet
		"ownerCertHash": fixtures.BuyerCertHash,
	}

	_, txErr := GetBalance.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when asset not found in wallet")
	testutils.AssertErrorStatus(t, txErr, 404, "should return 404 status")
}

// ============================================================================
// GetEscrowBalance Tests
// ============================================================================

func TestGetEscrowBalance_Success(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup: Create digital asset
	err := fixtures.CreateMockDigitalAsset(
		mockStub,
		fixtures.AssetID,
		fixtures.AssetSymbol,
		fixtures.AssetName,
		fixtures.IssuerCertHash,
		1000.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock digital asset: %v", err)
	}

	// Setup: Create buyer wallet with escrow balance
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		150.0, // balance
		75.0,  // escrow balance
	)
	if err != nil {
		t.Fatalf("Failed to create mock wallet: %v", err)
	}

	// Setup: Create user directory
	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock user directory: %v", err)
	}

	// Execute GetEscrowBalance
	args := map[string]any{
		"pubKey":        fixtures.BuyerPubKey,
		"assetSymbol":   fixtures.AssetSymbol,
		"ownerCertHash": fixtures.BuyerCertHash,
	}

	response, txErr := GetEscrowBalance.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "GetEscrowBalance should succeed")

	// Verify response
	var result map[string]any
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	testutils.AssertEqual(t, fixtures.BuyerWalletUUID, result["walletId"], "walletId mismatch")
	testutils.AssertEqual(t, fixtures.AssetSymbol, result["assetSymbol"], "assetSymbol mismatch")
	testutils.AssertEqual(t, 75.0, result["escrowBalance"], "escrowBalance mismatch")

	t.Log("✓ GetEscrowBalance returned correct escrow balance")
}

func TestGetEscrowBalance_WalletNotFound(t *testing.T) {
	wrapper, _ := testutils.NewMockStubWrapper()

	args := map[string]any{
		"pubKey":        "non-existent-pubkey",
		"assetSymbol":   "TST",
		"ownerCertHash": "some-cert-hash",
	}

	_, err := GetEscrowBalance.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, err, "should fail when wallet not found")
	testutils.AssertErrorStatus(t, err, 404, "should return 404 status")
}

func TestGetEscrowBalance_UnauthorizedAccess(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup
	err := fixtures.CreateMockDigitalAsset(
		mockStub,
		fixtures.AssetID,
		fixtures.AssetSymbol,
		fixtures.AssetName,
		fixtures.IssuerCertHash,
		1000.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock digital asset: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		150.0,
		75.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock wallet: %v", err)
	}

	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock user directory: %v", err)
	}

	// Try to access with wrong certificate hash
	args := map[string]any{
		"pubKey":        fixtures.BuyerPubKey,
		"assetSymbol":   fixtures.AssetSymbol,
		"ownerCertHash": "wrong-cert-hash",
	}

	_, txErr := GetEscrowBalance.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong certificate")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

func TestGetEscrowBalance_AssetNotInWallet(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup
	err := fixtures.CreateMockDigitalAsset(
		mockStub,
		fixtures.AssetID,
		fixtures.AssetSymbol,
		fixtures.AssetName,
		fixtures.IssuerCertHash,
		1000.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock digital asset: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		150.0,
		75.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock wallet: %v", err)
	}

	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock user directory: %v", err)
	}

	// Query for non-existent asset
	args := map[string]any{
		"pubKey":        fixtures.BuyerPubKey,
		"assetSymbol":   "NOTFOUND",
		"ownerCertHash": fixtures.BuyerCertHash,
	}

	_, txErr := GetEscrowBalance.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when asset not found")
	testutils.AssertErrorStatus(t, txErr, 404, "should return 404 status")
}

// ============================================================================
// GetWalletByOwner Tests
// ============================================================================

func TestGetWalletByOwner_Success(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup: Create digital asset
	err := fixtures.CreateMockDigitalAsset(
		mockStub,
		fixtures.AssetID,
		fixtures.AssetSymbol,
		fixtures.AssetName,
		fixtures.IssuerCertHash,
		1000.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock digital asset: %v", err)
	}

	// Setup: Create buyer wallet
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		150.0,
		50.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock wallet: %v", err)
	}

	// Setup: Create user directory
	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock user directory: %v", err)
	}

	// Execute GetWalletByOwner
	args := map[string]any{
		"pubKey":        fixtures.BuyerPubKey,
		"ownerCertHash": fixtures.BuyerCertHash,
	}

	response, txErr := GetWalletByOwner.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "GetWalletByOwner should succeed")

	// Verify response contains complete wallet
	var wallet map[string]any
	if err := json.Unmarshal(response, &wallet); err != nil {
		t.Fatalf("Failed to parse wallet: %v", err)
	}

	testutils.AssertEqual(t, fixtures.BuyerWalletID, wallet["walletId"], "walletId mismatch")
	testutils.AssertEqual(t, fixtures.BuyerPubKey, wallet["ownerPubKey"], "ownerPubKey mismatch")
	testutils.AssertEqual(t, fixtures.BuyerCertHash, wallet["ownerCertHash"], "ownerCertHash mismatch")

	// Verify balances exist
	balances, ok := wallet["balances"].([]any)
	if !ok {
		t.Fatal("Expected balances array")
	}
	if len(balances) != 1 {
		t.Errorf("Expected 1 balance entry, got %d", len(balances))
	}

	escrowBalances, ok := wallet["escrowBalances"].([]any)
	if !ok {
		t.Fatal("Expected escrowBalances array")
	}
	if len(escrowBalances) != 1 {
		t.Errorf("Expected 1 escrow balance entry, got %d", len(escrowBalances))
	}

	t.Log("✓ GetWalletByOwner returned complete wallet")
}

func TestGetWalletByOwner_WalletNotFound(t *testing.T) {
	wrapper, _ := testutils.NewMockStubWrapper()

	args := map[string]any{
		"pubKey":        "non-existent-pubkey",
		"ownerCertHash": "some-cert-hash",
	}

	_, err := GetWalletByOwner.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, err, "should fail when wallet not found")
	testutils.AssertErrorStatus(t, err, 404, "should return 404 status")
}

func TestGetWalletByOwner_UnauthorizedAccess(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup
	err := fixtures.CreateMockDigitalAsset(
		mockStub,
		fixtures.AssetID,
		fixtures.AssetSymbol,
		fixtures.AssetName,
		fixtures.IssuerCertHash,
		1000.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock digital asset: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		150.0,
		50.0,
	)
	if err != nil {
		t.Fatalf("Failed to create mock wallet: %v", err)
	}

	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock user directory: %v", err)
	}

	// Try to access with wrong certificate
	args := map[string]any{
		"pubKey":        fixtures.BuyerPubKey,
		"ownerCertHash": "wrong-cert-hash",
	}

	_, txErr := GetWalletByOwner.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong certificate")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

package transactions

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/testutils"
)

// ============================================================================
// CreateDigitalAsset Tests
// ============================================================================

func TestCreateDigitalAsset_Success(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	args := map[string]any{
		"name":        fixtures.AssetName,
		"symbol":      fixtures.AssetSymbol,
		"decimals":    2.0,
		"totalSupply": 1000000.0,
		"owner":       "test-owner",
		"issuerHash":  fixtures.IssuerCertHash,
	}

	response, txErr := CreateDigitalAsset.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "CreateDigitalAsset should succeed")

	// Parse response
	var createdAsset map[string]any
	if err := json.Unmarshal(response, &createdAsset); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify properties
	testutils.AssertEqual(t, fixtures.AssetName, createdAsset["name"], "name mismatch")
	testutils.AssertEqual(t, fixtures.AssetSymbol, createdAsset["symbol"], "symbol mismatch")
	testutils.AssertEqual(t, 2.0, createdAsset["decimals"], "decimals mismatch")
	testutils.AssertEqual(t, 1000000.0, createdAsset["totalSupply"], "totalSupply mismatch")
	testutils.AssertEqual(t, "test-owner", createdAsset["owner"], "owner mismatch")
	testutils.AssertEqual(t, fixtures.IssuerCertHash, createdAsset["issuerHash"], "issuerHash mismatch")

	// Verify asset was saved to ledger
	assetKey, exists := createdAsset["@key"].(string)
	if !exists {
		t.Fatal("Expected asset to have @key field")
	}

	_, exists = mockStub.State[assetKey]
	if !exists {
		t.Errorf("Expected asset to be saved with key '%s'", assetKey)
	}

	t.Log("✓ Digital asset created successfully")
}

// ============================================================================
// ReadDigitalAsset Tests
// ============================================================================

func TestReadDigitalAsset_Success(t *testing.T) {
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

	// Execute ReadDigitalAsset
	args := map[string]any{
		"uuid": fixtures.AssetID,
	}

	response, txErr := ReadDigitalAsset.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "ReadDigitalAsset should succeed")

	// Verify response
	var asset map[string]any
	if err := json.Unmarshal(response, &asset); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	testutils.AssertEqual(t, fixtures.AssetSymbol, asset["symbol"], "symbol mismatch")
	testutils.AssertEqual(t, fixtures.AssetName, asset["name"], "name mismatch")

	t.Log("✓ Digital asset read successfully")
}

func TestReadDigitalAsset_NotFound(t *testing.T) {
	wrapper, _ := testutils.NewMockStubWrapper()

	args := map[string]any{
		"uuid": "non-existent-asset-id",
	}

	_, txErr := ReadDigitalAsset.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when asset not found")
	testutils.AssertErrorStatus(t, txErr, 404, "should return 404 status")
}

// ============================================================================
// MintTokens Tests
// ============================================================================

func TestMintTokens_Success(t *testing.T) {
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
		100.0,
		0.0,
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

	// DEBUG: Dump all state keys
	fmt.Printf("\nDEBUG Current MockStub State:\n")
	for k := range mockStub.State {
		fmt.Printf("DEBUG   Key: %q\n", k)
	}
	fmt.Printf("\n")

	// Execute MintTokens
	args := map[string]any{
		"assetId":        fixtures.AssetID,
		"pubKey":         fixtures.BuyerPubKey,
		"amount":         50.0,
		"issuerCertHash": fixtures.IssuerCertHash,
	}

	response, txErr := MintTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "MintTokens should succeed")

	// Verify response
	var result map[string]any
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	testutils.AssertEqual(t, "Tokens minted successfully", result["message"], "message mismatch")
	testutils.AssertEqual(t, fixtures.AssetID, result["assetId"], "assetId mismatch")
	testutils.AssertEqual(t, 50.0, result["amount"], "amount mismatch")
	testutils.AssertEqual(t, 1050.0, result["totalSupply"], "totalSupply should increase")

	// Verify wallet balance updated
	walletBytes := mockStub.State["wallet:"+fixtures.BuyerWalletUUID]
	var wallet map[string]any
	json.Unmarshal(walletBytes, &wallet)
	balances := wallet["balances"].([]any)
	testutils.AssertEqual(t, 150.0, balances[0].(float64), "wallet balance should be 100 + 50")

	t.Log("✓ Tokens minted successfully")
}

func TestMintTokens_NewAssetInWallet(t *testing.T) {
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

	// Setup: Create wallet WITHOUT this asset
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		"other-asset-id", // Different asset
		100.0,
		0.0,
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

	// Execute MintTokens for new asset
	args := map[string]any{
		"assetId":        fixtures.AssetID,
		"pubKey":         fixtures.BuyerPubKey,
		"amount":         25.0,
		"issuerCertHash": fixtures.IssuerCertHash,
	}

	_, txErr := MintTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "MintTokens should succeed for new asset")

	// Verify wallet now has 2 assets
	walletBytes := mockStub.State["wallet:"+fixtures.BuyerWalletUUID]
	var wallet map[string]any
	json.Unmarshal(walletBytes, &wallet)
	digitalAssetTypes := wallet["digitalAssetTypes"].([]any)
	testutils.AssertEqual(t, 2, len(digitalAssetTypes), "wallet should have 2 assets")

	t.Log("✓ Tokens minted for new asset in wallet")
}

func TestMintTokens_UnauthorizedIssuer(t *testing.T) {
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
		100.0,
		0.0,
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

	// Try to mint with wrong issuer hash
	args := map[string]any{
		"assetId":        fixtures.AssetID,
		"pubKey":         fixtures.BuyerPubKey,
		"amount":         50.0,
		"issuerCertHash": "wrong-issuer-hash",
	}

	_, txErr := MintTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong issuer")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

func TestMintTokens_WalletNotFound(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup: Only create asset, no wallet
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

	args := map[string]any{
		"assetId":        fixtures.AssetID,
		"pubKey":         "non-existent-pubkey",
		"amount":         50.0,
		"issuerCertHash": fixtures.IssuerCertHash,
	}

	_, txErr := MintTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when wallet not found")
	testutils.AssertErrorStatus(t, txErr, 404, "should return 404 status")
}

// ============================================================================
// TransferTokens Tests
// ============================================================================

func TestTransferTokens_UnauthorizedSender(t *testing.T) {
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
		200.0,
		0.0,
	)
	if err != nil {
		t.Fatalf("Failed to create buyer wallet: %v", err)
	}

	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create buyer user directory: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.SellerPubKey,
		fixtures.SellerCertHash,
		fixtures.SellerWalletID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		50.0,
		0.0,
	)
	if err != nil {
		t.Fatalf("Failed to create seller wallet: %v", err)
	}

	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.SellerPubKeyHash,
		fixtures.SellerWalletUUID,
		fixtures.SellerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create seller user directory: %v", err)
	}

	// Try to transfer with wrong sender cert
	args := map[string]any{
		"fromPubKey":     fixtures.BuyerPubKey,
		"toPubKey":       fixtures.SellerPubKey,
		"assetId":        fixtures.AssetID,
		"amount":         50.0,
		"senderCertHash": "wrong-cert-hash",
	}

	_, txErr := TransferTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong sender cert")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

// func TestTransferTokens_ToNewAssetHolder(t *testing.T) {
// 	wrapper, mockStub := testutils.NewMockStubWrapper()
// 	fixtures := testutils.NewTestFixtures()
//
// 	// Setup
// 	err := fixtures.CreateMockDigitalAsset(
// 		mockStub,
// 		fixtures.AssetID,
// 		fixtures.AssetSymbol,
// 		fixtures.AssetName,
// 		fixtures.IssuerCertHash,
// 		1000.0,
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create mock digital asset: %v", err)
// 	}
//
// 	err = fixtures.CreateMockWallet(
// 		mockStub,
// 		fixtures.BuyerPubKey,
// 		fixtures.BuyerCertHash,
// 		fixtures.BuyerWalletID,
// 		fixtures.BuyerWalletUUID,
// 		fixtures.AssetID,
// 		200.0,
// 		0.0,
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create buyer wallet: %v", err)
// 	}
//
// 	err = fixtures.CreateMockUserDir(
// 		mockStub,
// 		fixtures.BuyerPubKeyHash,
// 		fixtures.BuyerWalletUUID,
// 		fixtures.BuyerCertHash,
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create buyer user directory: %v", err)
// 	}
//
// 	t.Logf("Created UserDir with key: userdir:%s", fixtures.BuyerPubKeyHash)
// 	allKeys := []string{}
// 	for key := range mockStub.State {
// 		if strings.HasPrefix(key, "userdir:") {
// 			allKeys = append(allKeys, key)
// 		}
// 	}
// 	t.Logf("All UserDir keys in state: %v", allKeys)
//
// 	// Now compute what TransferTokens will look for:
// 	hash := sha256.Sum256([]byte(fixtures.BuyerPubKey))
// 	expectedKey := "userdir:" + hex.EncodeToString(hash[:])
// 	t.Logf("TransferTokens will look for: %s", expectedKey)
// 	t.Logf("Keys match: %v", expectedKey == "userdir:"+fixtures.BuyerPubKeyHash)
//
// 	// Seller wallet WITHOUT this asset
// 	err = fixtures.CreateMockWallet(
// 		mockStub,
// 		fixtures.SellerPubKey,
// 		fixtures.SellerCertHash,
// 		fixtures.SellerWalletID,
// 		fixtures.SellerWalletUUID,
// 		"other-asset-id",
// 		100.0,
// 		0.0,
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create seller wallet: %v", err)
// 	}
//
// 	err = fixtures.CreateMockUserDir(
// 		mockStub,
// 		fixtures.SellerPubKeyHash,
// 		fixtures.SellerWalletUUID,
// 		fixtures.SellerCertHash,
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create seller user directory: %v", err)
// 	}
//
// 	// Transfer to seller who doesn't have this asset yet
// 	args := map[string]any{
// 		"fromPubKey":     fixtures.BuyerPubKey,
// 		"toPubKey":       fixtures.SellerPubKey,
// 		"assetId":        fixtures.AssetID,
// 		"amount":         30.0,
// 		"senderCertHash": fixtures.BuyerCertHash,
// 	}
//
// 	// Debug: Verify the hash computation
// 	hash = sha256.Sum256([]byte(fixtures.BuyerPubKey))
// 	computedHash := hex.EncodeToString(hash[:])
// 	t.Logf("BuyerPubKey: %s", fixtures.BuyerPubKey)
// 	t.Logf("Computed hash: %s", computedHash)
// 	t.Logf("Fixture hash: %s", fixtures.BuyerPubKeyHash)
//
// 	compositeKey := "userdir" + string('\x00') + fixtures.BuyerPubKeyHash
// 	t.Logf("Looking for composite key: %q", compositeKey)
//
// 	t.Logf("All keys in state:")
// 	for key := range mockStub.State {
// 		t.Logf("  Key: %q", key)
// 	}
//
// 	// Check if composite key exists
// 	value := mockStub.State[compositeKey]
// 	t.Logf("Composite key exists: %v", value != nil)
//
// 	// Check what's actually in the UserDirectory
// 	userDirBytes := mockStub.State["userdir:"+fixtures.BuyerPubKeyHash]
// 	t.Logf("UserDir key exists: %v", userDirBytes != nil)
//
// 	// Try with computed hash
// 	userDirBytes2 := mockStub.State["userdir:"+computedHash]
// 	t.Logf("UserDir with computed hash exists: %v", userDirBytes2 != nil)
//
// 	// ==========================
//
// 	_, txErr := TransferTokens.Routine(wrapper.StubWrapper, args)
// 	testutils.AssertNoError(t, txErr, "TransferTokens should succeed")
//
// 	// ✅ ADD THIS DEBUG BLOCK
// 	t.Log("=== POST-TRANSFER STATE DEBUG ===")
// 	t.Logf("Checking seller wallet updates...")
//
// 	// Read seller wallet from BOTH possible keys
// 	sellerWalletUUID := "wallet:" + fixtures.SellerWalletUUID
// 	sellerCompositeKey := "wallet" + string('\x00') + fixtures.SellerPubKey
//
// 	sellerByUUID := mockStub.State[sellerWalletUUID]
// 	sellerByComposite := mockStub.State[sellerCompositeKey]
//
// 	t.Logf("Seller wallet by UUID key (%s): exists=%v", sellerWalletUUID, sellerByUUID != nil)
// 	t.Logf("Seller wallet by composite key: exists=%v", sellerByComposite != nil)
//
// 	if sellerByUUID != nil {
// 		var wallet map[string]any
// 		json.Unmarshal(sellerByUUID, &wallet)
// 		t.Logf("Seller wallet (UUID key) assets: %+v", wallet["digitalAssetTypes"])
// 	}
//
// 	if sellerByComposite != nil {
// 		var wallet map[string]any
// 		json.Unmarshal(sellerByComposite, &wallet)
// 		t.Logf("Seller wallet (composite key) assets: %+v", wallet["digitalAssetTypes"])
// 	}
//
// 	// Also check what PutState calls were made
// 	t.Logf("PutState invocations: %v", mockStub.Invocations)
//
// 	// Verify seller now has the asset
// 	sellerWalletKey := "wallet:" + fixtures.SellerWalletUUID
// 	sellerWalletBytes := mockStub.State[sellerWalletKey]
//
// 	if sellerWalletBytes == nil {
// 		t.Fatalf("Seller wallet not found in state at key: %s", sellerWalletKey)
// 	}
//
// 	var sellerWallet map[string]any
// 	err = json.Unmarshal(sellerWalletBytes, &sellerWallet)
// 	if err != nil {
// 		t.Fatalf("Failed to unmarshal seller wallet: %v", err)
// 	}
//
// 	balances, ok := sellerWallet["balances"].([]any)
// 	if !ok {
// 		t.Fatalf("Balances not found or wrong type in seller wallet")
// 	}
//
// 	digitalAssetTypes, ok := sellerWallet["digitalAssetTypes"].([]any)
// 	if !ok {
// 		t.Fatalf("DigitalAssetTypes not found or wrong type in seller wallet")
// 	}
//
// 	// Debug: print what we have
// 	t.Logf("Seller wallet digitalAssetTypes: %+v", digitalAssetTypes)
// 	t.Logf("Seller wallet balances: %+v", balances)
//
// 	// Find the index of the transferred asset
// 	foundIndex := -1
// 	for i, assetRef := range digitalAssetTypes {
// 		var refAssetId string
// 		switch ref := assetRef.(type) {
// 		case map[string]any:
// 			if keyVal, exists := ref["@key"]; exists {
// 				refAssetId = strings.Split(keyVal.(string), ":")[1]
// 			}
// 		case string:
// 			refAssetId = ref
// 		}
//
// 		t.Logf("Checking asset at index %d: %s (looking for %s)", i, refAssetId, fixtures.AssetID)
//
// 		if refAssetId == fixtures.AssetID {
// 			foundIndex = i
// 			break
// 		}
// 	}
//
// 	if foundIndex == -1 {
// 		t.Fatalf("Transferred asset %s not found in seller wallet. Available assets: %+v",
// 			fixtures.AssetID, digitalAssetTypes)
// 	}
//
// 	testutils.AssertEqual(t, 30.0, balances[foundIndex].(float64), "new asset balance should be 30")
//
// 	t.Log("✓ Tokens transferred to new asset holder")
// }

// ============================================================================
// BurnTokens Tests
// ============================================================================

func TestBurnTokens_Success(t *testing.T) {
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
		200.0,
		0.0,
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

	// Execute BurnTokens
	args := map[string]any{
		"assetId":        fixtures.AssetID,
		"pubKey":         fixtures.BuyerPubKey,
		"amount":         50.0,
		"issuerCertHash": fixtures.IssuerCertHash,
	}

	response, txErr := BurnTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "BurnTokens should succeed")

	// Verify response
	var result map[string]any
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	testutils.AssertEqual(t, "Tokens burned successfully", result["message"], "message mismatch")
	testutils.AssertEqual(t, 50.0, result["amount"], "amount mismatch")
	testutils.AssertEqual(t, 950.0, result["totalSupply"], "totalSupply should decrease")

	// Verify wallet balance decreased
	walletBytes := mockStub.State["wallet:"+fixtures.BuyerWalletUUID]
	var wallet map[string]any
	json.Unmarshal(walletBytes, &wallet)
	balances := wallet["balances"].([]any)
	testutils.AssertEqual(t, 150.0, balances[0].(float64), "wallet balance should be 200 - 50")

	t.Log("✓ Tokens burned successfully")
}

func TestBurnTokens_InsufficientBalance(t *testing.T) {
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
		30.0, // Only 30 tokens
		0.0,
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

	// Try to burn more than available
	args := map[string]any{
		"assetId":        fixtures.AssetID,
		"pubKey":         fixtures.BuyerPubKey,
		"amount":         50.0,
		"issuerCertHash": fixtures.IssuerCertHash,
	}

	_, txErr := BurnTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with insufficient balance")
	testutils.AssertErrorStatus(t, txErr, 400, "should return 400 status")
}

func TestBurnTokens_UnauthorizedIssuer(t *testing.T) {
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
		200.0,
		0.0,
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

	// Try to burn with wrong issuer
	args := map[string]any{
		"assetId":        fixtures.AssetID,
		"pubKey":         fixtures.BuyerPubKey,
		"amount":         50.0,
		"issuerCertHash": "wrong-issuer-hash",
	}

	_, txErr := BurnTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong issuer")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

func TestBurnTokens_AssetNotInWallet(t *testing.T) {
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

	// Wallet with different asset
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		"other-asset-id",
		200.0,
		0.0,
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

	// Try to burn asset not in wallet
	args := map[string]any{
		"assetId":        fixtures.AssetID,
		"pubKey":         fixtures.BuyerPubKey,
		"amount":         50.0,
		"issuerCertHash": fixtures.IssuerCertHash,
	}

	_, txErr := BurnTokens.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when asset not in wallet")
	testutils.AssertErrorStatus(t, txErr, 404, "should return 404 status")
}

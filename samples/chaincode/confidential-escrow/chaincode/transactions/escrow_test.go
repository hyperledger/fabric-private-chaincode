package transactions

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow/chaincode/testutils"
)

// ============================================================================
// CreateAndLockEscrow Tests
// ============================================================================

func TestCreateAndLockEscrow_Success(t *testing.T) {
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

	// Setup: Create buyer wallet with sufficient balance
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		200.0, // available balance
		0.0,   // escrow balance
	)
	if err != nil {
		t.Fatalf("Failed to create buyer wallet: %v", err)
	}

	// Setup: Create buyer user directory
	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create buyer user directory: %v", err)
	}

	// Setup: Create seller wallet
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.SellerPubKey,
		fixtures.SellerCertHash,
		fixtures.SellerWalletID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		0.0, // seller starts with 0 balance
		0.0,
	)
	if err != nil {
		t.Fatalf("Failed to create seller wallet: %v", err)
	}

	// Setup: Create seller user directory
	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.SellerPubKeyHash,
		fixtures.SellerWalletUUID,
		fixtures.SellerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create seller user directory: %v", err)
	}

	// Execute CreateAndLockEscrow
	escrowAmount := 100.0
	args := map[string]any{
		"escrowId":      fixtures.EscrowID,
		"buyerPubKey":   fixtures.BuyerPubKey,
		"sellerPubKey":  fixtures.SellerPubKey,
		"amount":        escrowAmount,
		"assetType":     assets.Key{"@key": "digitalAsset:" + fixtures.AssetID},
		"parcelId":      fixtures.ParcelID,
		"secret":        fixtures.Secret,
		"buyerCertHash": fixtures.BuyerCertHash,
	}

	response, txErr := CreateAndLockEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "CreateAndLockEscrow should succeed")

	// Parse escrow response
	var createdEscrow map[string]any
	if err := json.Unmarshal(response, &createdEscrow); err != nil {
		t.Fatalf("Failed to parse escrow response: %v", err)
	}

	// Verify escrow properties
	testutils.AssertEqual(t, fixtures.EscrowID, createdEscrow["escrowId"], "escrowId mismatch")
	testutils.AssertEqual(t, fixtures.BuyerPubKey, createdEscrow["buyerPubKey"], "buyerPubKey mismatch")
	testutils.AssertEqual(t, fixtures.SellerPubKey, createdEscrow["sellerPubKey"], "sellerPubKey mismatch")
	testutils.AssertEqual(t, escrowAmount, createdEscrow["amount"], "amount mismatch")
	testutils.AssertEqual(t, "Active", createdEscrow["status"], "status should be Active")
	testutils.AssertEqual(t, fixtures.ParcelID, createdEscrow["parcelId"], "parcelId mismatch")

	// Verify condition hash was computed correctly
	expectedHash := sha256.Sum256([]byte(fixtures.Secret + fixtures.ParcelID))
	expectedCondition := hex.EncodeToString(expectedHash[:])
	testutils.AssertEqual(t, expectedCondition, createdEscrow["conditionValue"], "conditionValue mismatch")

	// Verify buyer wallet balances were updated
	buyerWalletKey := "wallet:" + fixtures.BuyerWalletUUID
	buyerWalletBytes, exists := mockStub.State[buyerWalletKey]
	if !exists {
		t.Fatal("Buyer wallet should exist in state")
	}

	var buyerWallet map[string]any
	if err := json.Unmarshal(buyerWalletBytes, &buyerWallet); err != nil {
		t.Fatalf("Failed to parse buyer wallet: %v", err)
	}

	balances := buyerWallet["balances"].([]any)
	escrowBalances := buyerWallet["escrowBalances"].([]any)

	testutils.AssertEqual(t, 100.0, balances[0].(float64), "buyer available balance should be reduced by escrow amount")
	testutils.AssertEqual(t, 100.0, escrowBalances[0].(float64), "buyer escrow balance should increase by escrow amount")

	t.Log("✓ Escrow created and funds locked successfully")
}

func TestCreateAndLockEscrow_InsufficientBalance(t *testing.T) {
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

	// Setup: Create buyer wallet with insufficient balance
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.BuyerPubKey,
		fixtures.BuyerCertHash,
		fixtures.BuyerWalletID,
		fixtures.BuyerWalletUUID,
		fixtures.AssetID,
		50.0, // only 50 available
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

	// Setup: Create seller wallet and directory
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.SellerPubKey,
		fixtures.SellerCertHash,
		fixtures.SellerWalletID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		0.0,
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

	// Try to lock more than available
	args := map[string]any{
		"escrowId":      fixtures.EscrowID,
		"buyerPubKey":   fixtures.BuyerPubKey,
		"sellerPubKey":  fixtures.SellerPubKey,
		"amount":        100.0, // trying to lock 100 but only have 50
		"assetType":     assets.Key{"@key": "digitalAsset:" + fixtures.AssetID},
		"parcelId":      fixtures.ParcelID,
		"secret":        fixtures.Secret,
		"buyerCertHash": fixtures.BuyerCertHash,
	}

	_, txErr := CreateAndLockEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with insufficient balance")
	testutils.AssertErrorStatus(t, txErr, 400, "should return 400 status")
}

func TestCreateAndLockEscrow_BuyerWalletNotFound(t *testing.T) {
	wrapper, _ := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	args := map[string]any{
		"escrowId":      fixtures.EscrowID,
		"buyerPubKey":   "non-existent-buyer",
		"sellerPubKey":  fixtures.SellerPubKey,
		"amount":        100.0,
		"assetType":     assets.Key{"@key": "digitalAsset:" + fixtures.AssetID},
		"parcelId":      fixtures.ParcelID,
		"secret":        fixtures.Secret,
		"buyerCertHash": fixtures.BuyerCertHash,
	}

	_, txErr := CreateAndLockEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when buyer wallet not found")
	testutils.AssertErrorStatus(t, txErr, 404, "should return 404 status")
}

func TestCreateAndLockEscrow_UnauthorizedBuyer(t *testing.T) {
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
		0.0,
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

	// Try with wrong certificate hash
	args := map[string]any{
		"escrowId":      fixtures.EscrowID,
		"buyerPubKey":   fixtures.BuyerPubKey,
		"sellerPubKey":  fixtures.SellerPubKey,
		"amount":        100.0,
		"assetType":     assets.Key{"@key": "digitalAsset:" + fixtures.AssetID},
		"parcelId":      fixtures.ParcelID,
		"secret":        fixtures.Secret,
		"buyerCertHash": "wrong-cert-hash",
	}

	_, txErr := CreateAndLockEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong certificate")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

// func TestCreateAndLockEscrow_SellerWalletNotFound(t *testing.T) {
// 	wrapper, mockStub := testutils.NewMockStubWrapper()
// 	fixtures := testutils.NewTestFixtures()
//
// 	// Setup only buyer, not seller
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
// 	args := map[string]any{
// 		"escrowId":      fixtures.EscrowID,
// 		"buyerPubKey":   fixtures.BuyerPubKey,
// 		"sellerPubKey":  "non-existent-seller",
// 		"amount":        100.0,
// 		"assetType":     assets.Key{"@key": "digitalAsset:" + fixtures.AssetID},
// 		"parcelId":      fixtures.ParcelID,
// 		"secret":        fixtures.Secret,
// 		"buyerCertHash": fixtures.BuyerCertHash,
// 	}
//
// 	_, txErr := CreateAndLockEscrow.Routine(wrapper.StubWrapper, args)
// 	testutils.AssertError(t, txErr, "should fail when seller wallet not found")
// 	testutils.AssertErrorStatus(t, txErr, 404, "should return 404 status")
// }

// ============================================================================
// VerifyEscrowCondition Tests
// ============================================================================

// func TestVerifyEscrowCondition_Success(t *testing.T) {
// 	wrapper, mockStub := testutils.NewMockStubWrapper()
// 	fixtures := testutils.NewTestFixtures()
//
// 	// Setup: Create an active escrow
// 	err := fixtures.CreateMockEscrow(
// 		mockStub,
// 		fixtures.EscrowID,
// 		fixtures.BuyerPubKey,
// 		fixtures.SellerPubKey,
// 		fixtures.BuyerWalletUUID,
// 		fixtures.SellerWalletUUID,
// 		fixtures.AssetID,
// 		100.0,
// 		fixtures.ParcelID,
// 		fixtures.Secret,
// 		"Active",
// 		fixtures.BuyerCertHash,
// 	)
// 	if err != nil {
// 		t.Fatalf("Failed to create mock escrow: %v", err)
// 	}
//
// 	// Execute VerifyEscrowCondition
// 	args := map[string]any{
// 		"escrowId": fixtures.EscrowID,
// 		"secret":   fixtures.Secret,
// 		"parcelId": fixtures.ParcelID,
// 	}
//
// 	response, txErr := VerifyEscrowCondition.Routine(wrapper.StubWrapper, args)
// 	testutils.AssertNoError(t, txErr, "VerifyEscrowCondition should succeed")
//
// 	// Parse response
// 	var result map[string]any
// 	if err := json.Unmarshal(response, &result); err != nil {
// 		t.Fatalf("Failed to parse response: %v", err)
// 	}
//
// 	testutils.AssertEqual(t, "Condition verified successfully", result["message"], "message mismatch")
// 	testutils.AssertEqual(t, fixtures.EscrowID, result["escrowId"], "escrowId mismatch")
// 	testutils.AssertEqual(t, "ReadyForRelease", result["status"], "status should be ReadyForRelease")
// 	testutils.AssertEqual(t, fixtures.ParcelID, result["parcelId"], "parcelId mismatch")
//
// 	// Verify computed hash matches expected
// 	expectedHash := sha256.Sum256([]byte(fixtures.Secret + fixtures.ParcelID))
// 	expectedCondition := hex.EncodeToString(expectedHash[:])
// 	testutils.AssertEqual(t, expectedCondition, result["computedHash"], "computedHash mismatch")
//
// 	// Verify escrow status was updated in state
// 	escrowKey := "escrow:" + fixtures.EscrowID
// 	escrowBytes, exists := mockStub.State[escrowKey]
// 	if !exists {
// 		t.Fatal("Escrow should exist in state")
// 	}
//
// 	var escrow map[string]any
// 	if err := json.Unmarshal(escrowBytes, &escrow); err != nil {
// 		t.Fatalf("Failed to parse escrow: %v", err)
// 	}
//
// 	testutils.AssertEqual(t, "ReadyForRelease", escrow["status"], "escrow status should be updated to ReadyForRelease")
//
// 	t.Log("✓ Escrow condition verified successfully")
// }

func TestVerifyEscrowCondition_WrongSecret(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup: Create an active escrow
	err := fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Active",
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	// Try to verify with wrong secret
	args := map[string]any{
		"escrowId": fixtures.EscrowID,
		"secret":   "wrong-secret",
		"parcelId": fixtures.ParcelID,
	}

	_, txErr := VerifyEscrowCondition.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong secret")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 status")
}

func TestVerifyEscrowCondition_EscrowNotActive(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup: Create a released escrow
	err := fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Released", // Already released
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	args := map[string]any{
		"escrowId": fixtures.EscrowID,
		"secret":   fixtures.Secret,
		"parcelId": fixtures.ParcelID,
	}

	_, txErr := VerifyEscrowCondition.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when escrow not active")
	testutils.AssertErrorStatus(t, txErr, 400, "should return 400 status")
}

func TestVerifyEscrowCondition_EscrowNotFound(t *testing.T) {
	wrapper, _ := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	args := map[string]any{
		"escrowId": "non-existent-escrow",
		"secret":   fixtures.Secret,
		"parcelId": fixtures.ParcelID,
	}

	_, txErr := VerifyEscrowCondition.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when escrow not found")
}

// ============================================================================
// ReleaseEscrow Tests
// ============================================================================

func TestReleaseEscrow_Success(t *testing.T) {
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
		100.0, // available balance
		100.0, // escrow balance
	)
	if err != nil {
		t.Fatalf("Failed to create buyer wallet: %v", err)
	}

	// Setup: Create seller wallet
	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.SellerPubKey,
		fixtures.SellerCertHash,
		fixtures.SellerWalletID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		0.0, // seller starts with 0
		0.0,
	)
	if err != nil {
		t.Fatalf("Failed to create seller wallet: %v", err)
	}

	// Setup: Create active escrow
	err = fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Active",
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	// Execute ReleaseEscrow
	args := map[string]any{
		"escrowUUID":     fixtures.EscrowID,
		"secret":         fixtures.Secret,
		"parcelId":       fixtures.ParcelID,
		"sellerCertHash": fixtures.SellerCertHash,
	}

	response, txErr := ReleaseEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "ReleaseEscrow should succeed")

	// Parse response
	var result map[string]any
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	testutils.AssertEqual(t, "Escrow released successfully", result["message"], "message mismatch")
	testutils.AssertEqual(t, fixtures.EscrowID, result["escrowId"], "escrowId mismatch")
	testutils.AssertEqual(t, 100.0, result["amount"], "amount mismatch")

	// Verify buyer wallet: escrow balance reduced
	buyerWalletKey := "wallet:" + fixtures.BuyerWalletUUID
	buyerWalletBytes := mockStub.State[buyerWalletKey]
	var buyerWallet map[string]any
	json.Unmarshal(buyerWalletBytes, &buyerWallet)

	buyerEscrowBalances := buyerWallet["escrowBalances"].([]any)
	testutils.AssertEqual(t, 0.0, buyerEscrowBalances[0].(float64), "buyer escrow balance should be 0")

	// Verify seller wallet: balance increased
	sellerWalletKey := "wallet:" + fixtures.SellerWalletUUID
	sellerWalletBytes := mockStub.State[sellerWalletKey]
	var sellerWallet map[string]any
	json.Unmarshal(sellerWalletBytes, &sellerWallet)

	sellerBalances := sellerWallet["balances"].([]any)
	testutils.AssertEqual(t, 100.0, sellerBalances[0].(float64), "seller balance should be 100")

	// Verify escrow status updated to Released
	escrowKey := "escrow:" + fixtures.EscrowID
	escrowBytes := mockStub.State[escrowKey]
	var escrow map[string]any
	json.Unmarshal(escrowBytes, &escrow)

	testutils.AssertEqual(t, "Released", escrow["status"], "escrow status should be Released")

	t.Log("✓ Escrow released and funds transferred successfully")
}

func TestReleaseEscrow_WrongSecret(t *testing.T) {
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
		100.0,
	)
	if err != nil {
		t.Fatalf("Failed to create buyer wallet: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.SellerPubKey,
		fixtures.SellerCertHash,
		fixtures.SellerWalletID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		0.0,
		0.0,
	)
	if err != nil {
		t.Fatalf("Failed to create seller wallet: %v", err)
	}

	err = fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Active",
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	// Try with wrong secret
	args := map[string]any{
		"escrowUUID":     fixtures.EscrowID,
		"secret":         "wrong-secret",
		"parcelId":       fixtures.ParcelID,
		"sellerCertHash": fixtures.SellerCertHash,
	}

	_, txErr := ReleaseEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong secret")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 status")
}

func TestReleaseEscrow_WrongParcelId(t *testing.T) {
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
		100.0,
	)
	if err != nil {
		t.Fatalf("Failed to create buyer wallet: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.SellerPubKey,
		fixtures.SellerCertHash,
		fixtures.SellerWalletID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		0.0,
		0.0,
	)
	if err != nil {
		t.Fatalf("Failed to create seller wallet: %v", err)
	}

	err = fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Active",
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	// Try with wrong parcelId
	args := map[string]any{
		"escrowUUID":     fixtures.EscrowID,
		"secret":         fixtures.Secret,
		"parcelId":       "wrong-parcel-id",
		"sellerCertHash": fixtures.SellerCertHash,
	}

	_, txErr := ReleaseEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong parcelId")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 status")
}

func TestReleaseEscrow_UnauthorizedSeller(t *testing.T) {
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
		100.0,
	)
	if err != nil {
		t.Fatalf("Failed to create buyer wallet: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.SellerPubKey,
		fixtures.SellerCertHash,
		fixtures.SellerWalletID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		0.0,
		0.0,
	)
	if err != nil {
		t.Fatalf("Failed to create seller wallet: %v", err)
	}

	err = fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Active",
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	// Try with wrong seller certificate
	args := map[string]any{
		"escrowUUID":     fixtures.EscrowID,
		"secret":         fixtures.Secret,
		"parcelId":       fixtures.ParcelID,
		"sellerCertHash": "wrong-seller-cert",
	}

	_, txErr := ReleaseEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong seller certificate")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

func TestReleaseEscrow_EscrowNotActive(t *testing.T) {
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
		100.0,
	)
	if err != nil {
		t.Fatalf("Failed to create buyer wallet: %v", err)
	}

	err = fixtures.CreateMockWallet(
		mockStub,
		fixtures.SellerPubKey,
		fixtures.SellerCertHash,
		fixtures.SellerWalletID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		0.0,
		0.0,
	)
	if err != nil {
		t.Fatalf("Failed to create seller wallet: %v", err)
	}

	// Create already released escrow
	err = fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Released", // Already released
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	args := map[string]any{
		"escrowUUID":     fixtures.EscrowID,
		"secret":         fixtures.Secret,
		"parcelId":       fixtures.ParcelID,
		"sellerCertHash": fixtures.SellerCertHash,
	}

	_, txErr := ReleaseEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when escrow not active")
	testutils.AssertErrorStatus(t, txErr, 400, "should return 400 status")
}

// ============================================================================
// RefundEscrow Tests
// ============================================================================

func TestRefundEscrow_Success(t *testing.T) {
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
		100.0, // available balance
		100.0, // escrow balance
	)
	if err != nil {
		t.Fatalf("Failed to create buyer wallet: %v", err)
	}

	// Setup: Create buyer user directory
	err = fixtures.CreateMockUserDir(
		mockStub,
		fixtures.BuyerPubKeyHash,
		fixtures.BuyerWalletUUID,
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create buyer user directory: %v", err)
	}

	// Setup: Create active escrow
	err = fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Active",
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	// Execute RefundEscrow
	args := map[string]any{
		"escrowUUID":    fixtures.EscrowID,
		"buyerPubKey":   fixtures.BuyerPubKey,
		"buyerCertHash": fixtures.BuyerCertHash,
	}

	response, txErr := RefundEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "RefundEscrow should succeed")

	// Parse response
	var result map[string]any
	if err := json.Unmarshal(response, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	testutils.AssertEqual(t, "Escrow refunded successfully", result["message"], "message mismatch")
	testutils.AssertEqual(t, fixtures.EscrowID, result["escrowUUID"], "escrowUUID mismatch")
	testutils.AssertEqual(t, 100.0, result["amount"], "amount mismatch")

	// Verify buyer wallet: funds moved from escrow back to available
	buyerWalletKey := "wallet:" + fixtures.BuyerWalletUUID
	buyerWalletBytes := mockStub.State[buyerWalletKey]
	var buyerWallet map[string]any
	json.Unmarshal(buyerWalletBytes, &buyerWallet)

	buyerBalances := buyerWallet["balances"].([]any)
	buyerEscrowBalances := buyerWallet["escrowBalances"].([]any)

	testutils.AssertEqual(t, 200.0, buyerBalances[0].(float64), "buyer balance should increase by 100")
	testutils.AssertEqual(t, 0.0, buyerEscrowBalances[0].(float64), "buyer escrow balance should be 0")

	// Verify escrow status updated to Refunded
	escrowKey := "escrow:" + fixtures.EscrowID
	escrowBytes := mockStub.State[escrowKey]
	var escrow map[string]any
	json.Unmarshal(escrowBytes, &escrow)

	testutils.AssertEqual(t, "Refunded", escrow["status"], "escrow status should be Refunded")

	t.Log("✓ Escrow refunded successfully")
}

func TestRefundEscrow_UnauthorizedBuyer(t *testing.T) {
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
		100.0,
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

	err = fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Active",
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	// Try with wrong buyer certificate
	args := map[string]any{
		"escrowUUID":    fixtures.EscrowID,
		"buyerPubKey":   fixtures.BuyerPubKey,
		"buyerCertHash": "wrong-buyer-cert",
	}

	_, txErr := RefundEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail with wrong buyer certificate")
	testutils.AssertErrorStatus(t, txErr, 403, "should return 403 Unauthorized")
}

func TestRefundEscrow_EscrowNotActive(t *testing.T) {
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
		100.0,
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

	// Create already refunded escrow
	err = fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Refunded", // Already refunded
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	args := map[string]any{
		"escrowUUID":    fixtures.EscrowID,
		"buyerPubKey":   fixtures.BuyerPubKey,
		"buyerCertHash": fixtures.BuyerCertHash,
	}

	_, txErr := RefundEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when escrow not active")
	testutils.AssertErrorStatus(t, txErr, 400, "should return 400 status")
}

func TestRefundEscrow_EscrowNotFound(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup buyer wallet and directory
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
		100.0,
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

	args := map[string]any{
		"escrowUUID":    "non-existent-escrow",
		"buyerPubKey":   fixtures.BuyerPubKey,
		"buyerCertHash": fixtures.BuyerCertHash,
	}

	_, txErr := RefundEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when escrow not found")
	testutils.AssertErrorStatus(t, txErr, 404, "should return 404 status")
}

// ============================================================================
// ReadEscrow Tests
// ============================================================================

func TestReadEscrow_Success(t *testing.T) {
	wrapper, mockStub := testutils.NewMockStubWrapper()
	fixtures := testutils.NewTestFixtures()

	// Setup: Create an escrow
	err := fixtures.CreateMockEscrow(
		mockStub,
		fixtures.EscrowID,
		fixtures.BuyerPubKey,
		fixtures.SellerPubKey,
		fixtures.BuyerWalletUUID,
		fixtures.SellerWalletUUID,
		fixtures.AssetID,
		100.0,
		fixtures.ParcelID,
		fixtures.Secret,
		"Active",
		fixtures.BuyerCertHash,
	)
	if err != nil {
		t.Fatalf("Failed to create mock escrow: %v", err)
	}

	// Execute ReadEscrow
	args := map[string]any{
		"uuid": fixtures.EscrowID,
	}

	response, txErr := ReadEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertNoError(t, txErr, "ReadEscrow should succeed")

	// Parse response
	var escrow map[string]any
	if err := json.Unmarshal(response, &escrow); err != nil {
		t.Fatalf("Failed to parse escrow: %v", err)
	}

	// Verify escrow properties
	testutils.AssertEqual(t, fixtures.EscrowID, escrow["escrowId"], "escrowId mismatch")
	testutils.AssertEqual(t, fixtures.BuyerPubKey, escrow["buyerPubKey"], "buyerPubKey mismatch")
	testutils.AssertEqual(t, fixtures.SellerPubKey, escrow["sellerPubKey"], "sellerPubKey mismatch")
	testutils.AssertEqual(t, 100.0, escrow["amount"], "amount mismatch")
	testutils.AssertEqual(t, "Active", escrow["status"], "status mismatch")
	testutils.AssertEqual(t, fixtures.ParcelID, escrow["parcelId"], "parcelId mismatch")

	t.Log("✓ Escrow read successfully")
}

func TestReadEscrow_NotFound(t *testing.T) {
	wrapper, _ := testutils.NewMockStubWrapper()

	args := map[string]any{
		"uuid": "non-existent-escrow",
	}

	_, txErr := ReadEscrow.Routine(wrapper.StubWrapper, args)
	testutils.AssertError(t, txErr, "should fail when escrow not found")
}

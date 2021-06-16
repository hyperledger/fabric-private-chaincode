// +build WITH_PDO_CRYPTO

package crypto

func init() {

	// add another test case using the PDO crypto lib
	allTestCases = append(allTestCases,
		testCases{"PDO Crypto", NewPdoCrypto()},
	)
}

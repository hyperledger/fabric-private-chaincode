package header

var (
	// Name defines the chaincode identifier used in registration and logging.
	Name = "Confidential Escrow"
	// Version specifies the current chaincode version following semantic versioning.
	Version = "1.0.0"
	// Colors defines UI color schemes for different contexts.
	// The @default scheme uses a blue/gray palette suitable for financial applications.
	Colors = map[string][]string{
		"@default": {"#4267B2", "#34495E", "#ECF0F1"},
	}
	// Title provides human-readable descriptions for the chaincode in different contexts.
	Title = map[string]string{
		"@default": "Confidential Digital Assets & Programmable Escrow",
	}
)

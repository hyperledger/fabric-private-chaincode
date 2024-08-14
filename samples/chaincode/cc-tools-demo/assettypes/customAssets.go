package assettypes

import (
	"github.com/hyperledger-labs/cc-tools/assets"
)

// CustomAssets contains all assets inserted via GoFabric's Template mode.
// For local development, this can be empty or could contain assets that
// are supposed to be defined via the Template mode.
var CustomAssets = []assets.AssetType{}

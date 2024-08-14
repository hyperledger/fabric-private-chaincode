package datatypes

import (
	"fmt"
	"strconv"

	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
)

// Example of a custom data type using enum-like structure (iota)
// This allows the use of verification by const values instead of float64, improving readability
// Example:
// 		if assetMap["bookType"].(float64) == (float64)(BookTypeHardcover)
// 			...

type BookType float64

const (
	BookTypeHardcover BookType = iota
	BookTypePaperback
	BookTypeEbook
)

// CheckType checks if the given value is defined as valid BookType consts
func (b BookType) CheckType() errors.ICCError {
	switch b {
	case BookTypeHardcover:
		return nil
	case BookTypePaperback:
		return nil
	case BookTypeEbook:
		return nil
	default:
		return errors.NewCCError("invalid type", 400)
	}

}

var bookType = assets.DataType{
	AcceptedFormats: []string{"number"},
	DropDownValues: map[string]interface{}{
		"Hardcover": BookTypeHardcover,
		"Paperback": BookTypePaperback,
		"Ebook":     BookTypeEbook,
	},
	Description: ``,

	Parse: func(data interface{}) (string, interface{}, errors.ICCError) {
		var dataVal float64
		switch v := data.(type) {
		case float64:
			dataVal = v
		case int:
			dataVal = (float64)(v)
		case BookType:
			dataVal = (float64)(v)
		case string:
			var err error
			dataVal, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return "", nil, errors.WrapErrorWithStatus(err, "asset property must be an integer, is %t", 400)
			}
		default:
			return "", nil, errors.NewCCError("asset property must be an integer, is %t", 400)
		}

		retVal := (BookType)(dataVal)
		err := retVal.CheckType()
		return fmt.Sprint(retVal), retVal, err
	},
}

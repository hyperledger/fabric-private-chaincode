package datatypes

import (
	"strings"

	"github.com/hyperledger-labs/cc-tools/assets"
	"github.com/hyperledger-labs/cc-tools/errors"
)

var cpf = assets.DataType{
	AcceptedFormats: []string{"string"},
	Parse: func(data interface{}) (string, interface{}, errors.ICCError) {
		cpf, ok := data.(string)
		if !ok {
			return "", nil, errors.NewCCError("property must be a string", 400)
		}

		cpf = strings.ReplaceAll(cpf, ".", "")
		cpf = strings.ReplaceAll(cpf, "-", "")

		if len(cpf) != 11 {
			return "", nil, errors.NewCCError("CPF must have 11 digits", 400)
		}

		var vd0 int
		for i, d := range cpf {
			if i >= 9 {
				break
			}
			dnum := int(d) - '0'
			vd0 += (10 - i) * dnum
		}
		vd0 = 11 - vd0%11
		if vd0 > 9 {
			vd0 = 0
		}
		if int(cpf[9])-'0' != vd0 {
			return "", nil, errors.NewCCError("Invalid CPF", 400)
		}

		var vd1 int
		for i, d := range cpf {
			if i >= 10 {
				break
			}
			dnum := int(d) - '0'
			vd1 += (11 - i) * dnum
		}
		vd1 = 11 - vd1%11
		if vd1 > 9 {
			vd1 = 0
		}
		if int(cpf[10])-'0' != vd1 {
			return "", nil, errors.NewCCError("Invalid CPF", 400)
		}

		return cpf, cpf, nil
	},
}

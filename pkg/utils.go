package pkg

import (
	"errors"
	"math/big"

	"github.com/shopspring/decimal"
)

var (
	BigZero      = big.NewInt(0)
	Big9         = big.NewInt(9)
	Big10        = big.NewInt(10)
	Big18        = big.NewInt(18)
	BigGwei      = new(big.Int).Exp(Big10, Big9, nil)
	BigEther     = new(big.Int).Exp(Big10, Big18, nil)
	DecimalEther = decimal.New(1, 18)
)

func BigToDecimal(input *big.Int) decimal.Decimal {
	return decimal.RequireFromString(input.String())
}

func EtherToWei(input *big.Int) *big.Int {
	return new(big.Int).Mul(input, BigEther)
}

func AmountToWei(input string) (*big.Int, error) {
	value, err := decimal.NewFromString(input)
	if err != nil {
		return nil, err
	}

	wei := value.Mul(DecimalEther)
	result, done := new(big.Int).SetString(wei.String(), 10)
	if !done {
		return nil, errors.New("Error parsing amount")
	}

	return result, nil
}

func Units(wei bool) string {
	if wei {
		return "wei"
	}

	return "ETH"
}

// Implying input is wei
func WeiOrEther(input *big.Int, wei bool) decimal.Decimal {
	result := BigToDecimal(input)
	if !wei {
		result = result.Div(DecimalEther)
	}

	return result
}

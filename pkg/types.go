package pkg

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type TxData struct {
	ID      uint32
	Address common.Address
	Balance *big.Int // Real balance
	Value   *big.Int // Transferred value (balance - fees)
}

type Result struct {
	Data    []TxData
	GasCost *big.Int
	Target  *big.Int // Target balance
	Total   *big.Int // Total available balance
}

func (res Result) PrintConfirmation(destination common.Address, amount *big.Int, wei bool) {
	units := Units(wei)
	total := WeiOrEther(res.Total, wei)
	target := WeiOrEther(res.Target, wei)
	printAmount := WeiOrEther(amount, wei)

	fmt.Println()
	fmt.Printf("Available balance: %s %s\n", total.String(), units)
	fmt.Printf("Amount to transfer: %s of %s %s\n", target.String(), printAmount.String(), units)
	for _, data := range res.Data {
		value := WeiOrEther(data.Value, wei)
		fmt.Printf("- Will send %s %s from %s\n", value, units, data.Address.String())
	}

	fmt.Printf("Destination: %s\n", destination.String())
	fmt.Println()
	fmt.Printf("Do you wish to proceed? [yes/no]: ")
}

func (res Result) PrintSummary(wei bool) {
	units := Units(wei)
	total := WeiOrEther(res.Total, wei)

	fmt.Println()
	fmt.Printf("Total available balance: %s %s\n", total.String(), units)
	for _, data := range res.Data {
		balance := WeiOrEther(data.Balance, wei)
		fmt.Printf("- Address №%d (%s) has %s %s\n", data.ID, data.Address.String(), balance.String(), units)
	}
}

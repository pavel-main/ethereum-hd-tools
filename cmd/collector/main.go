package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pavel-main/ethereum-hd-tools/pkg"
	"github.com/urfave/cli"
)

func main() {
	// Init CLI
	app := cli.NewApp()
	app.Name = "collector"
	app.Usage = "collects funds from multiple derived addresses"
	app.Version = "1.0.1"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "wei",
			Usage: "output values in wei",
		},
		cli.StringFlag{
			Name:  "rpc",
			Usage: "Ethereum node RPC URL",
		},
		cli.Uint64Flag{
			Name:  "chain",
			Usage: "Ethereum chain ID",
			Value: 1,
		},
		cli.Uint64Flag{
			Name:  "fee",
			Usage: "custom gas price (in gwei)",
			Value: 0,
		},
		cli.StringFlag{
			Name:  "xprv",
			Usage: "source account extended private key",
		},
		cli.UintFlag{
			Name:  "from",
			Usage: "start account number",
			Value: 0,
		},
		cli.UintFlag{
			Name:  "until",
			Usage: "final account number",
			Value: 0,
		},
		cli.StringFlag{
			Name:  "amount",
			Usage: "desired amount (in ETH)",
		},
		cli.StringFlag{
			Name:  "destination",
			Usage: "destination address",
		},
	}

	app.Action = func(ctx *cli.Context) error {
		// Parse CLI flags
		rpc, xprv, from, until, dest, amount, err := parseFlags(ctx)
		if err != nil {
			return err
		}

		// Init manager
		chain := ctx.Uint64("chain")
		fee := ctx.Uint64("fee")
		wei := ctx.Bool("wei")
		manager, err := pkg.NewManager(rpc, chain, fee, wei)
		if err != nil {
			return err
		}

		// Set gas price
		if err := manager.SetGasPrice(); err != nil {
			return err
		}

		// Init keychain
		keychain, err := pkg.New(xprv)
		if err != nil {
			return err
		}

		// Get balance
		result, err := manager.GetBalancesUntil(keychain, amount, from, until)
		if err != nil {
			return err
		}

		if len(result.Data) == 0 {
			return errors.New("No funds available (for selected accounts)")
		}

		// Confirmation window
		destination := common.HexToAddress(dest)
		result.PrintConfirmation(destination, amount, wei)

		// Scan for input
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if scanner.Text() == "yes" {
			fmt.Println()

			total, err := manager.Collect(keychain, result, destination)
			units := pkg.Units(wei)
			sent := pkg.WeiOrEther(total, wei)
			fmt.Printf("Total sent: %s %s\n", sent.String(), units)
			return err
		}

		fmt.Printf("Operation aborted\n")
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}

func parseFlags(ctx *cli.Context) (string, string, uint, uint, string, *big.Int, error) {
	rpc := ctx.String("rpc")
	if len(rpc) == 0 {
		return "", "", 0, 0, "", nil, errors.New("Please provide RPC URL using --rpc flag")
	}

	xprv := ctx.String("xprv")
	if len(xprv) == 0 {
		return "", "", 0, 0, "", nil, errors.New("Please provide account extended private key using --xprv flag")
	}

	from := ctx.Uint("from")
	until := ctx.Uint("until")
	if until == 0 {
		return "", "", 0, 0, "", nil, errors.New("Please provide account scan limit with --until flag")
	}

	if from > until {
		return "", "", 0, 0, "", nil, errors.New("From should be greater than until")
	}

	dest := ctx.String("destination")
	if len(dest) == 0 {
		return "", "", 0, 0, "", nil, errors.New("Please provide destination address using --destination flag")
	}

	if !common.IsHexAddress(dest) {
		return "", "", 0, 0, "", nil, errors.New("Please provide valid destination address using --destination flag")
	}

	raw := ctx.String("amount")
	if len(raw) == 0 {
		return "", "", 0, 0, "", nil, errors.New("Please provide amount using --amount flag")
	}

	amount, err := pkg.AmountToWei(raw)
	if err != nil {
		return "", "", 0, 0, "", nil, err
	}

	if amount.Cmp(pkg.BigZero) <= 0 { // amount <= 0
		return "", "", 0, 0, "", nil, errors.New("Amount should be greater than zero")
	}

	return rpc, xprv, from, until, dest, amount, nil
}

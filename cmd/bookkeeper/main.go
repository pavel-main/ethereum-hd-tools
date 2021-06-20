package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/pavel-main/ethereum-hd-tools/pkg"
	"github.com/urfave/cli"
)

func main() {
	// Init CLI
	app := cli.NewApp()
	app.Name = "bookkeeper"
	app.Usage = "shows balances from multiple derived addresses"
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
		cli.StringFlag{
			Name:  "xpub",
			Usage: "destination account extended public key",
		},
		cli.UintFlag{
			Name:  "from",
			Usage: "start account number",
			Value: 0,
		},
		cli.UintFlag{
			Name:  "until",
			Usage: "final account number",
		},
	}

	app.Action = func(ctx *cli.Context) error {
		rpc, xpub, from, until, err := parseFlags(ctx)
		if err != nil {
			return err
		}

		// Init manager
		chain := ctx.Uint64("chain")
		wei := ctx.Bool("wei")
		manager, err := pkg.NewManager(rpc, chain, 0, wei)
		if err != nil {
			return err
		}

		// Set gas price
		if err := manager.SetGasPrice(); err != nil {
			return err
		}

		// Init keychain
		keychain, err := pkg.New(xpub)
		if err != nil {
			return err
		}

		// Get balances
		result, err := manager.GetBalances(keychain, from, until)
		if err != nil {
			return err
		}

		if len(result.Data) == 0 {
			return errors.New("No funds available (for selected accounts)")
		}

		result.PrintSummary(wei)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}

func parseFlags(ctx *cli.Context) (string, string, uint, uint, error) {
	// Parse CLI flags
	rpc := ctx.String("rpc")
	if len(rpc) == 0 {
		return "", "", 0, 0, errors.New("Please provide RPC URL using --rpc flag")
	}

	xpub := ctx.String("xpub")
	if len(xpub) == 0 {
		return "", "", 0, 0, errors.New("Please provide account extended public key using --xpub flag")
	}

	from := ctx.Uint("from")
	until := ctx.Uint("until")
	if until == 0 {
		return "", "", 0, 0, errors.New("Please provide account scan limit with --until flag")
	}

	if from > until {
		return "", "", 0, 0, errors.New("From should be greater than until")
	}

	return rpc, xpub, from, until, nil
}

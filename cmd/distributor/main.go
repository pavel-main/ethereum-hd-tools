package main

import (
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/pavel-main/ethereum-hd-tools/pkg"
	"github.com/urfave/cli"
)

func main() {
	// Init CLI
	app := cli.NewApp()
	app.Name = "distributor"
	app.Usage = "distributes funds to multiple addresses derived from BIP-44 HD wallet"
	app.Version = "1.0.1"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "wei",
			Usage: "output values in wei",
		},
		cli.BoolFlag{
			Name:  "random",
			Usage: "randomize values a bit",
		},
		cli.StringFlag{
			Name:  "rpc",
			Usage: "Ethereum node RPC URL",
			Value: "http://localhost:8545",
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
			Name:  "xpub",
			Usage: "destination account extended public key",
		},
		cli.StringFlag{
			Name:  "prv",
			Usage: "source account private key",
		},
		cli.UintFlag{
			Name:  "from",
			Usage: "start account number",
			Value: 0,
		},
		cli.UintFlag{
			Name:  "until",
			Usage: "final account number",
			Value: 1,
		},
		cli.UintFlag{
			Name:  "step",
			Usage: "step size",
			Value: 1,
		},
		cli.StringFlag{
			Name:  "amount",
			Usage: "amount to transfer to each account (in ETH)",
		},
	}

	app.Action = func(ctx *cli.Context) error {
		// Parse CLI flags
		rpc, prv, xpub, from, until, step, amount, err := parseFlags(ctx)
		if err != nil {
			return err
		}

		// Prepare accounts
		accounts := []uint32{}
		for i := from; i <= until; i += step {
			accounts = append(accounts, uint32(i))
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
		keychain, err := pkg.New(xpub)
		if err != nil {
			return err
		}

		// Derive keys
		keys, err := keychain.DeriveMultiPublic(accounts)
		if err != nil {
			return err
		}

		// Parse private key
		key, err := manager.GetPrivateKey(prv)
		if err != nil {
			return err
		}

		// Distribute
		random := ctx.Bool("random")
		total, err := manager.Distribute(key, keys, amount, random)
		fmt.Printf("Sent %d transactions of %d\n", total, len(accounts))
		return err
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %s", err.Error())
		os.Exit(1)
	}
}

func parseFlags(ctx *cli.Context) (string, string, string, uint, uint, uint, *big.Int, error) {
	rpc := ctx.String("rpc")
	if len(rpc) == 0 {
		return "", "", "", 0, 0, 0, nil, errors.New("Please provide RPC URL using --rpc flag")
	}

	prv := ctx.String("prv")
	if len(prv) == 0 {
		return "", "", "", 0, 0, 0, nil, errors.New("Please provide private key using --prv flag")
	}

	xpub := ctx.String("xpub")
	if len(xpub) == 0 {
		return "", "", "", 0, 0, 0, nil, errors.New("Please provie account extended public key using --xpub flag")
	}

	from := ctx.Uint("from")
	until := ctx.Uint("until")
	step := ctx.Uint("step")
	if until == 0 {
		return "", "", "", 0, 0, 0, nil, errors.New("Please provide account scan limit with --until flag")
	}

	if step == 0 {
		return "", "", "", 0, 0, 0, nil, errors.New("Please provide valid step size with --step flag")
	}

	if from > until {
		return "", "", "", 0, 0, 0, nil, errors.New("From should be greater than until")
	}

	raw := ctx.String("amount")
	if len(raw) == 0 {
		return "", "", "", 0, 0, 0, nil, errors.New("Please provide amount using --amount flag")
	}

	amount, err := pkg.AmountToWei(raw)
	if err != nil {
		return "", "", "", 0, 0, 0, nil, err
	}

	if amount.Cmp(pkg.BigZero) <= 0 { // amount <= 0
		return "", "", "", 0, 0, 0, nil, errors.New("Amount should be greater than zero")
	}

	return rpc, prv, xpub, from, until, step, amount, nil
}

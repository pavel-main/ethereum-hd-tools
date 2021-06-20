# Distributor

Usage:

```
NAME:
   distributor - distributes funds to multiple addresses derived from BIP-44 HD wallet

USAGE:
   distributor [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --rpc value     Ethereum node RPC URL (default: "http://localhost:8545")
   --chain value   Ethereum chain ID (default: 1)
   --fee value     custom gas price (in gwei) (default: 0)
   --xpub value    destination account extended public key
   --prv value     source account private key
   --from value    start account number (default: 0)
   --until value   final account number (default: 1)
   --step value    step size (default: 1)
   --amount value  amount to transfer to each account (in ETH) (default: 0)
   --help, -h      show help
   --version, -v   print the version
```

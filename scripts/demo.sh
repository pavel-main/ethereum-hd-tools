#!/usr/bin/env bash
set -euo pipefail

# Set parameters
SLEEP=30
CHAIN=4
RPC=https://rinkeby.infura.io

EACH=0.05
TOTAL=0.6

DEST=0x8888460E435D2DDff7108c12c50cc363c6057b8B
PRV=0c36641c01f626a63b0583be839605779e0ac08348b3fa68319d21afbfcfb17d

XPUB=xpub6C2LnDVzUb4BdyDfTgSMg9beCUbYj898Za3ZeoENfdUqsmVd3dWeUEeLAXnm4zQa9PSozSTDScY52paafnxj3kb2SSXNn3wd7Y6uF4v52Si
XPRV=xprv9y2zNhy6eDVtRV9CMeuMK1eueSm4KfRHCM7xrQpm7HwrzyAUW6CPvSKrKEzunJxoYaDieB2MUDtT5hwFYeKb19wyMtf9THAh5kAaug7rTwy

# Distribute and wait
./distributor --chain=$CHAIN --rpc=$RPC --xpub=$XPUB --prv=$PRV --from=0 --until=9 --amount=$EACH --random
echo "Waiting 30s for blocks to be mined..."
sleep $SLEEP

# Collect
./collector --chain=$CHAIN --rpc=$RPC --from=0 --until=20 --xprv=$XPRV --destination=$DEST --amount=$TOTAL
echo "Done"

package pkg

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Manager struct {
	Wei      bool
	ChainID  *big.Int
	GasPrice *big.Int
	GasLimit *big.Int
	GasCost  *big.Int
	Context  context.Context
	Client   *ethclient.Client
}

func NewManager(url string, chainID, gasPrice uint64, wei bool) (*Manager, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		return nil, err
	}

	m := new(Manager)
	m.Wei = wei
	m.Context = context.Background()
	m.Client = client
	m.ChainID = new(big.Int).SetUint64(chainID)

	gasPriceBig := new(big.Int).SetUint64(gasPrice)
	m.GasPrice = new(big.Int).Mul(gasPriceBig, BigGwei)
	m.GasLimit = big.NewInt(21000)
	m.GasCost = new(big.Int).Mul(m.GasPrice, m.GasLimit)
	return m, nil
}

func (m *Manager) GetPrivateKey(input string) (*ecdsa.PrivateKey, error) {
	key, err := crypto.HexToECDSA(input)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (m *Manager) SetGasPrice() error {
	// Get gas price (if necessary)
	if m.GasPrice.Cmp(BigZero) != 1 {
		fmt.Printf("Fetching suggested gas price...\n")
		gasPrice, err := m.Client.SuggestGasPrice(m.Context)
		if err != nil {
			return err
		}

		if gasPrice.Cmp(BigZero) == 1 {
			m.GasPrice = gasPrice
			m.GasCost = new(big.Int).Mul(m.GasPrice, m.GasLimit)
		} else {
			return errors.New("Network returned invalid gas price")
		}
	}

	units := Units(m.Wei)
	gasCost := WeiOrEther(m.GasCost, m.Wei)
	fmt.Printf("Gas cost per tx: %s %s\n", gasCost.String(), units)
	return nil
}

func (m *Manager) Distribute(prv *ecdsa.PrivateKey, keys []*btcec.PublicKey, amount *big.Int, random bool) (int, error) {
	// Get nonce
	from := crypto.PubkeyToAddress(prv.PublicKey)
	nonce, err := m.Client.PendingNonceAt(m.Context, from)
	if err != nil {
		return 0, err
	}

	units := Units(m.Wei)
	total := 0

	fmt.Printf("From address: %s\n", from.String())

	entropy := rand.NewSource(time.Now().UnixNano())
	limit := new(big.Int).SetInt64(1000000000)

	for _, key := range keys {
		value := amount
		if random {
			epsilon := new(big.Int).Rand(rand.New(entropy), limit)
			value = value.Add(value, epsilon)
		}

		// Print destination & value
		to := crypto.PubkeyToAddress(*key.ToECDSA())
		printVal := WeiOrEther(value, m.Wei)
		fmt.Printf("Sending %s %s to %s\n", printVal, units, to.String())

		// Sign tx
		rawTx := types.NewTransaction(nonce, to, value, m.GasLimit.Uint64(), m.GasPrice, nil)
		tx, err := types.SignTx(rawTx, types.NewEIP155Signer(m.ChainID), prv)
		if err != nil {
			return total, err
		}

		// Send tx
		fmt.Printf("Sending transaction %s\n", tx.Hash().String())
		if err := m.Client.SendTransaction(m.Context, tx); err != nil {
			return total, err
		}

		// Increase counters
		nonce++
		total++
	}

	return total, nil
}

func (m *Manager) GetBalances(keychain *Keychain, from, until uint) (*Result, error) {
	data := []TxData{}
	total := BigZero

	for i := from; i <= until; i = i + 1 {
		accountID := uint32(i)
		key, err := keychain.DerivePublic(accountID)
		if err != nil {
			return nil, err
		}

		address := crypto.PubkeyToAddress(*key.ToECDSA())
		fmt.Printf("Fetching balance for account %d (%s)\r", i, address.String())
		balance, err := m.Client.BalanceAt(m.Context, address, nil)
		if err != nil {
			return nil, err
		}

		// Update results
		if balance.Uint64() > 0 {
			total = new(big.Int).Add(total, balance)
			data = append(data, TxData{
				ID:      accountID,
				Address: address,
				Balance: balance,
			})
		}
	}

	fmt.Println()
	result := &Result{Total: total, Data: data, GasCost: m.GasCost}
	return result, nil
}

func (m *Manager) GetBalancesUntil(keychain *Keychain, amount *big.Int, from, until uint) (*Result, error) {
	data := []TxData{}
	total := BigZero
	target := BigZero

	for i := from; i <= until; i = i + 1 {
		accountID := uint32(i)
		key, err := keychain.DerivePublic(accountID)
		if err != nil {
			return nil, err
		}

		address := crypto.PubkeyToAddress(*key.ToECDSA())
		fmt.Printf("Fetching balance for account %d (%s)\r", i, address.String())
		balance, err := m.Client.BalanceAt(m.Context, address, nil)
		if err != nil {
			return nil, err
		}

		// Subtract fees (if not used by bookkeeper), skip if no funds
		delta := new(big.Int)
		if amount != nil {
			delta := delta.Sub(balance, m.GasCost)
			//fmt.Printf("Delta (balance - gas costs) == %s\n", delta.String())
			if delta.Cmp(BigZero) <= 0 { // delta <= 0
				//fmt.Printf("Delta <= 0, skipping\n")
				continue
			}
		}

		// Compare with target amount
		value := delta
		sum := new(big.Int).Add(target, delta)
		if amount != nil && sum.Cmp(amount) >= 0 { // (total + account balance - fees) >= target amount
			//fmt.Printf("Corner case, amount: %d, total: %d\n", amount.Uint64(), total.Uint64())
			value = new(big.Int).Sub(amount, target)
		}

		// Update results
		if value.Cmp(BigZero) > 0 {
			total = new(big.Int).Add(total, balance)
			target = new(big.Int).Add(target, value)
			data = append(data, TxData{
				ID:      accountID,
				Address: address,
				Balance: balance,
				Value:   value,
			})
		}
	}

	fmt.Println()
	result := &Result{Total: total, Target: target, Data: data, GasCost: m.GasCost}
	return result, nil
}

func (m *Manager) Collect(keychain *Keychain, result *Result, to common.Address) (*big.Int, error) {
	total := BigZero
	units := Units(m.Wei)

	for _, data := range result.Data {
		key, err := keychain.DerivePrivate(data.ID)
		if err != nil {
			return total, err
		}

		prv, err := crypto.ToECDSA(key.Serialize())
		if err != nil {
			return total, err
		}

		nonce, err := m.Client.PendingNonceAt(m.Context, data.Address)
		if err != nil {
			return total, err
		}

		printValue := WeiOrEther(data.Value, m.Wei)
		fmt.Printf("Sending %s %s from %s\n", printValue.String(), units, data.Address.String())

		rawTx := types.NewTransaction(nonce, to, data.Value, m.GasLimit.Uint64(), m.GasPrice, nil)
		tx, err := types.SignTx(rawTx, types.NewEIP155Signer(m.ChainID), prv)
		if err != nil {
			return total, err
		}

		fmt.Printf("Sending transaction %s\n", tx.Hash().String())
		if err := m.Client.SendTransaction(m.Context, tx); err != nil {
			return total, err
		}

		total = total.Add(total, data.Value)
	}

	return total, nil
}

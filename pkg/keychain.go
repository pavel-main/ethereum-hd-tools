package pkg

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/hdkeychain"
)

type Keychain struct {
	AccountKey  *hdkeychain.ExtendedKey
	ExtChainKey *hdkeychain.ExtendedKey
}

func New(xpub string) (*Keychain, error) {
	account, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return nil, err
	}

	external, err := account.Child(0)
	if err != nil {
		return nil, err
	}

	k := new(Keychain)
	k.AccountKey = account
	k.ExtChainKey = external
	return k, nil
}

func (k *Keychain) DerivePublic(index uint32) (*btcec.PublicKey, error) {
	child, err := k.ExtChainKey.Child(index)
	if err != nil {
		return nil, err
	}

	return child.ECPubKey()
}

func (k *Keychain) DerivePrivate(index uint32) (*btcec.PrivateKey, error) {
	child, err := k.ExtChainKey.Child(index)
	if err != nil {
		return nil, err
	}

	return child.ECPrivKey()
}

func (k *Keychain) DeriveMultiPublic(accounts []uint32) ([]*btcec.PublicKey, error) {
	keys := []*btcec.PublicKey{}
	for _, account := range accounts {
		key, err := k.DerivePublic(account)
		if err != nil {
			return keys, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (k *Keychain) DeriveMultiPrivate(accounts []uint32) ([]*btcec.PrivateKey, error) {
	keys := []*btcec.PrivateKey{}
	for _, account := range accounts {
		key, err := k.DerivePrivate(account)
		if err != nil {
			return keys, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

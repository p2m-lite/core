package web3

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/p2m-lite/core/daemon/internal/contract"
)

func SendLog(cAddr string, key *ecdsa.PrivateKey, ph, turbidity int) error {
	client, cErr := ethclient.Dial("https://opbnb-testnet-rpc.bnbchain.org")
	if cErr != nil {
		return cErr
	}
	defer client.Close()

	chainID, nErr := client.NetworkID(context.Background())
	if nErr != nil {
		return nErr
	}

	address := common.HexToAddress(cAddr)
	instance, iErr := contract.NewP2MContract(address, client)
	if iErr != nil {
		return iErr
	}
	auth, tErr := bind.NewKeyedTransactorWithChainID(key, chainID)
	if tErr != nil {
		return tErr
	}

	_, err := instance.StoreLog(auth, big.NewInt(int64(ph)), big.NewInt(int64(turbidity)))
	return err
}

package web3

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/p2m-lite/core/daemon/internal/contract"
)

func SendLog(cAddr string, key *ecdsa.PrivateKey, ph, turbidity int) error {
	pk := key.Public()
	pubAddress := common.HexToAddress(crypto.PubkeyToAddress(*pk.(*ecdsa.PublicKey)).Hex())
	fmt.Println("Sending log to blockchain via pub-address:", pubAddress.Hex())
	
	client, cErr := ethclient.Dial(os.Getenv("BLOCKCHAIN_URL"))
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

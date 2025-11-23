package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from system env")
	}

	rpcURL := os.Getenv("OPBNB_RPC_URL")
	privKeyHex := os.Getenv("PRIVATE_KEY")
	binFilePath := os.Getenv("BIN_FILE_PATH")

	if rpcURL == "" || privKeyHex == "" || binFilePath == "" {
		log.Fatal("Please set OPBNB_RPC_URL, PRIVATE_KEY, and BIN_FILE_PATH in .env")
	}

	// 2. Connect to the Eth client (opBNB Testnet)
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	fmt.Println("Connected to RPC:", rpcURL)

	// 3. Load Private Key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privKeyHex, "0x"))
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Println("Deploying from address:", fromAddress.Hex())

	// 4. Read the compiled bytecode (.bin file)
	bytecodeBytes, err := os.ReadFile(binFilePath)
	if err != nil {
		log.Fatalf("Failed to read bin file: %v", err)
	}

	// Clean up the string (remove whitespace/newlines) and decode hex
	bytecodeStr := strings.TrimSpace(string(bytecodeBytes))
	contractBytecode, err := hex.DecodeString(bytecodeStr)
	if err != nil {
		log.Fatalf("Failed to decode bytecode hex: %v", err)
	}

	// 5. Get Network Chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Chain ID: %v\n", chainID)

	// 6. Get Nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	// 7. Estimate Gas Price and Limit
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	head, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	gasFeeCap := new(big.Int).Add(head.BaseFee, gasTipCap)

	// Construct a call message to estimate gas limit
	msg := ethereum.CallMsg{
		From:      fromAddress,
		To:        nil, // nil = contract creation
		Data:      contractBytecode,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
	}

	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Fatalf("Failed to estimate gas: %v", err)
	}

	// Add a small buffer to gas limit for safety
	gasLimit = gasLimit + (gasLimit / 10)
	fmt.Printf("Estimated Gas: %v\n", gasLimit)

	// 8. Create Transaction (EIP-1559 Dynamic Fee)
	txData := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        nil, // Contract creation
		Value:     big.NewInt(0),
		Data:      contractBytecode,
	}

	tx := types.NewTx(txData)

	// 9. Sign Transaction
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	// 10. Broadcast Transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n--- Transaction Sent! ---\n")
	fmt.Printf("Tx Hash: %s\n", signedTx.Hash().Hex())
	fmt.Println("Waiting for mining...")

	// 11. Wait for Receipt
	receipt, err := bindWaitMined(context.Background(), client, signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n--- Contract Deployed! ---\n")
	fmt.Printf("Contract Address: %s\n", receipt.ContractAddress.Hex())
	fmt.Printf("Block Number: %v\n", receipt.BlockNumber)
	fmt.Printf("Gas Used: %v\n", receipt.GasUsed)
}

// Helper to wait for mining (simplified version of bind.WaitMined)
func bindWaitMined(ctx context.Context, client *ethclient.Client, tx *types.Transaction) (*types.Receipt, error) {
	queryTicker := time.NewTicker(time.Second)
	defer queryTicker.Stop()

	for {
		receipt, err := client.TransactionReceipt(ctx, tx.Hash())
		if err == nil {
			return receipt, nil
		}

		// Wait for the next round
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
		}
	}
}

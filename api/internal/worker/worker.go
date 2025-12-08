package worker

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"time"

	"p2m-lite/config"
	"p2m-lite/internal/contract"
	"p2m-lite/internal/database"
	"p2m-lite/vals"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func StartListener(cfg *config.Config) {
	go func() {
		client, err := ethclient.Dial(cfg.BlockchainURL)
		if err != nil {
			log.Printf("Listener: Failed to connect to blockchain: %v", err)
			return
		}

		contractAddr := common.HexToAddress(cfg.ContractAddress)
		p2m, err := contract.NewP2MContract(contractAddr, client)
		if err != nil {
			log.Printf("Listener: Failed to instantiate contract: %v", err)
			return
		}

		logs := make(chan *contract.P2MContractLogStored)
		// Subscribe to ALL logs (nil filter)
		sub, err := p2m.WatchLogStored(nil, logs, nil)
		if err != nil {
			log.Printf("Listener: Failed to subscribe to logs: %v", err)
			return
		}
		defer sub.Unsubscribe()

		log.Println("Listener: Started listening for blockchain events...")

		for {
			select {
			case err := <-sub.Err():
				log.Printf("Listener: Subscription error: %v", err)
				// Reconnect logic could go here
				return
			case vLog := <-logs:
				recorderAddr := vLog.Recorder.Hex() // Keep original casing or normalize?
				// Let's rely on DB COLLATE NOCASE, but passing Hex() is standard.

				// Fetch location
				lat, lon := database.GetRecorderLocation(recorderAddr)

				// Insert into DB
				newLog := database.Log{
					Recorder:  recorderAddr,
					Ph:        int(vLog.PhValue.Int64()),
					Turbidity: int(vLog.Turbidity.Int64()),
					Timestamp: vLog.Timestamp.Int64(),
					Lat:       lat,
					Lon:       lon,
				}
				if err := database.DB.Create(&newLog).Error; err != nil {
					log.Printf("Listener: Failed to insert log: %v", err)
				} else {
					log.Printf("Listener: Log stored for %s", vLog.Recorder.Hex())
				}
			}
		}
	}()
}

func StartAnalyzer(cfg *config.Config) {
	ticker := time.NewTicker(vals.AnalysisIntervalHours)
	go func() {
		for range ticker.C {
			log.Println("Analyzer: Starting analysis cycle...")
			analyze(cfg)
		}
	}()
}

func analyze(cfg *config.Config) {
	// 1. Get all unique recorders from logs in the last M days
	cutoff := time.Now().AddDate(0, 0, -vals.LookbackPeriodDays).Unix()

	type Result struct {
		Recorder     string
		AvgPH        float64
		AvgTurbidity float64
	}

	var results []Result
	err := database.DB.Model(&database.Log{}).
		Select("recorder, AVG(ph) as avg_ph, AVG(turbidity) as avg_turbidity").
		Where("timestamp > ?", cutoff).
		Group("recorder").
		Scan(&results).Error

	if err != nil {
		log.Printf("Analyzer: Failed to query logs: %v", err)
		return
	}

	for _, res := range results {
		recorder := res.Recorder
		avgPH := res.AvgPH
		avgTurbidity := res.AvgTurbidity

		// 2. Check if processed recently (Cooldown N days)
		if isProcessed(recorder) {
			log.Printf("Analyzer: Skipping %s (Cooldown)", recorder)
			continue
		}

		// 3. Evaluate Quality
		isLowQuality := avgPH < vals.MinPH || avgPH > vals.MaxPH || avgTurbidity > vals.MaxTurbidity

		var actionErr error
		if isLowQuality {
			log.Printf("Analyzer: Low quality detected for %s. Sending email...", recorder)
			lat, lon := database.GetRecorderLocation(recorder)
			actionErr = sendEmail(cfg, recorder, avgPH, avgTurbidity, lat, lon)
		} else {
			log.Printf("Analyzer: Good quality for %s. Sending reward...", recorder)
			actionErr = sendReward(cfg, recorder)
		}

		// 4. Mark as processed if action successful
		if actionErr == nil {
			markProcessed(recorder)
		} else {
			log.Printf("Analyzer: Action failed for %s: %v", recorder, actionErr)
		}
	}
}

func isProcessed(recorder string) bool {
	var count int64
	database.DB.Model(&database.ProcessedRecorder{}).
		Where("recorder = ? AND expires_at > ?", recorder, time.Now().Unix()).
		Count(&count)
	return count > 0
}

func markProcessed(recorder string) {
	expiresAt := time.Now().AddDate(0, 0, vals.CooldownPeriodDays).Unix()
	processed := database.ProcessedRecorder{
		Recorder:    recorder,
		ProcessedAt: time.Now().Unix(),
		ExpiresAt:   expiresAt,
	}
	// Upsert
	if err := database.DB.Save(&processed).Error; err != nil {
		log.Printf("Analyzer: Failed to mark processed: %v", err)
	}
}

func sendReward(cfg *config.Config, toAddress string) error {
	client, err := ethclient.Dial(cfg.BlockchainURL)
	if err != nil {
		return err
	}
	defer client.Close()

	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return log.Output(1, "error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}

	value := big.NewInt(vals.RewardAmount) // in wei
	gasLimit := uint64(21000)              // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return err
	}

	to := common.HexToAddress(toAddress)
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return err
	}

	return client.SendTransaction(context.Background(), signedTx)
}

package ws

import (
	"fmt"
	"log"
	"net/http"
	"p2m-lite/config"
	"p2m-lite/internal/contract"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

func HandleLogs(c *gin.Context, cfg *config.Config) {
	recorderAddr := c.Param("Recorder")
	var filter []common.Address

	if recorderAddr != "" {
		if !common.IsHexAddress(recorderAddr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recorder address"})
			return
		}
		filter = []common.Address{common.HexToAddress(recorderAddr)}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}
	defer conn.Close()

	if strings.HasPrefix(cfg.BlockchainURL, "http") {
		log.Printf("Error: BlockchainURL must be a WebSocket URL (ws:// or wss://) for subscriptions, got: %s", cfg.BlockchainURL)
		conn.WriteJSON(gin.H{"error": "BlockchainURL must be ws:// or wss:// for real-time logs"})
		return
	}

	client, err := ethclient.Dial(cfg.BlockchainURL)
	if err != nil {
		log.Printf("Failed to connect to blockchain: %v", err)
		conn.WriteJSON(gin.H{"error": "Failed to connect to blockchain"})
		return
	}
	defer client.Close()

	contractAddr := common.HexToAddress(cfg.ContractAddress)
	p2m, err := contract.NewP2MContract(contractAddr, client)
	if err != nil {
		log.Printf("Failed to instantiate contract: %v", err)
		conn.WriteJSON(gin.H{"error": "Failed to instantiate contract"})
		return
	}

	logs := make(chan *contract.P2MContractLogStored)
	sub, err := p2m.WatchLogStored(nil, logs, filter)
	if err != nil {
		log.Printf("Failed to subscribe to logs: %v", err)
		conn.WriteJSON(gin.H{"error": "Failed to subscribe to logs"})
		return
	}
	defer sub.Unsubscribe()
	fmt.Println("Subscribed to log events for recorder:", recorderAddr)

	// Handle client disconnect
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				sub.Unsubscribe()
				return
			}
		}
	}()

	for {
		select {
		case err := <-sub.Err():
			log.Printf("Subscription error: %v", err)
			conn.WriteJSON(gin.H{"error": "Subscription error"})
			return
		case vLog := <-logs:
			err := conn.WriteJSON(gin.H{
				"recorder":  vLog.Recorder.Hex(),
				"phValue":   vLog.PhValue.String(),
				"turbidity": vLog.Turbidity.String(),
				"timestamp": vLog.Timestamp.String(),
			})
			if err != nil {
				log.Printf("Failed to write to websocket: %v", err)
				return
			}
		}
	}
}

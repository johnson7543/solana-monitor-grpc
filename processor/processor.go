package processor

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rpcpool/yellowstone-grpc/examples/golang/model"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

// HandleTransaction processes a transaction response
func HandleTransaction(resp *pb.SubscribeUpdate, ownerSet map[string]struct{}) {
	now := time.Now().UTC()
	loc, _ := time.LoadLocation("Asia/Shanghai")
	formattedTime := now.In(loc).Format("2006-01-02_15-04-05.05")

	// Check if the UpdateOneof field is of type SubscribeUpdate_Transaction
	if update, ok := resp.UpdateOneof.(*pb.SubscribeUpdate_Transaction); ok {
		log.Printf("%+v Received Transaction: %+v\n", formattedTime, update.Transaction)

		if update.Transaction.Transaction.Meta.PreTokenBalances != nil {
			go writeToFile(formattedTime, resp)

			startTime := time.Now()
			transactions := AnalyzeBalances(update.Transaction.Transaction.Meta.PreTokenBalances, update.Transaction.Transaction.Meta.PostTokenBalances, ownerSet)
			endTime := time.Now()

			log.Printf("Time taken for analysis: %.6f seconds\n", endTime.Sub(startTime).Seconds())
			log.Printf("%+v Transaction Detail: %+v\n", endTime.In(loc).Format("2006-01-02_15-04-05.05"), transactions)

		}
	} else {
		// log.Printf("%+v Received non-transaction response: %+v\n", formattedTime, resp)
	}
}

// AnalyzeBalances analyzes the balances before and after a transaction
func AnalyzeBalances(preBalances, postBalances []*pb.TokenBalance, ownerSet map[string]struct{}) []model.Transaction {
	if len(preBalances) == 0 || len(postBalances) == 0 {
		log.Printf("Empty pre or post balances\n")
		return nil
	}

	ownerBalances := make(map[string]map[string]float64)
	transactions := []model.Transaction{}

	// Process preBalances
	for _, balance := range preBalances {
		if _, exists := ownerSet[balance.Owner]; exists {
			if ownerBalances[balance.Owner] == nil {
				ownerBalances[balance.Owner] = make(map[string]float64)
			}
			ownerBalances[balance.Owner][balance.Mint] = balance.UiTokenAmount.UiAmount
		}
	}

	// Process postBalances and compare
	for _, balance := range postBalances {
		if _, exists := ownerSet[balance.Owner]; exists {
			if ownerBalances[balance.Owner] == nil {
				ownerBalances[balance.Owner] = make(map[string]float64)
			}

			preAmount := ownerBalances[balance.Owner][balance.Mint]
			postAmount := balance.UiTokenAmount.UiAmount
			ownerBalances[balance.Owner][balance.Mint] = postAmount

			change := postAmount - preAmount
			if change != 0 {
				action := "buy"
				if change < 0 {
					action = "sell"
					change = -change
				}

				transactions = append(transactions, model.Transaction{
					TokenAddress: balance.Mint,
					Action:       action,
					Amount:       change,
				})
			}
		}
	}

	return transactions
}

func writeToFile(formattedTime string, resp *pb.SubscribeUpdate) {
	filename := fmt.Sprintf("response_%s.json", formattedTime)

	jsonData, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response to JSON: %v", err)
		return
	}

	f, _ := os.Create(filename)
	defer f.Close()

	_, _ = f.Write(jsonData)
}

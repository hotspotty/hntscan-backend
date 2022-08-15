package handlers

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hntscan/db"
	"log"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/labstack/echo/v4"
)

func GetTransactions(c echo.Context) error {

	var transactions []Transaction

	page := c.QueryParam("page")

	if page == "" {
		page = "0"
	}

	offset, err := strconv.Atoi(page)
	if err != nil {
		fmt.Println("GetTransactions", err)
	}

	limit := 25
	offset = offset * limit

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("transaction-%v", offset)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query("SELECT block, hash, type, time, fields FROM transactions ORDER BY block DESC LIMIT $1 OFFSET $2", limit, offset)
			if err != nil {
				fmt.Println("GetTransactions-sql", err)
			}

			var block, time sql.NullInt64
			var hash, tx_type, fields sql.NullString

			defer rows.Close()

			for rows.Next() {

				err = rows.Scan(&block, &hash, &tx_type, &time, &fields)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				transactions = append(transactions, Transaction{
					"transaction",
					block.Int64,
					hash.String,
					tx_type.String,
					time.Int64,
					fields.String,
				})

			}

			if err := enc.Encode(transactions); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&transactions); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return c.JSON(200, transactions)
}

func GetSingleTransaction(c echo.Context) error {

	input := c.Param("tx")

	tx := getTransactionData(input)

	return c.JSON(200, tx)
}

func GetRewardTxPagination(c echo.Context) error {

	input := c.Param("tx")

	page := c.QueryParam("page")

	if page == "" {
		page = "0"
	}
	pageInt, _ := strconv.Atoi(page)

	limit := c.QueryParam("limit")
	if limit == "" {
		limit = "25"
	}
	limitInt, _ := strconv.Atoi(limit)

	tx := getTransactionDataPagination(input, pageInt, limitInt)

	return c.JSON(200, tx.Rewards)

}

func getTransactionData(hash string) []Transaction {

	tx := make([]Transaction, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("single-tx-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow("SELECT block, hash, type, fields, time FROM transactions WHERE hash = $1", hash)

			var block, timestamp sql.NullInt64
			var hash, txType, fields sql.NullString

			err = row.Scan(&block, &hash, &txType, &fields, &timestamp)

			if err != nil {
				log.Printf("[ERROR GetSingleTransaction] %v", err)
			}

			// Parse spacific tx types here
			if txType.String == "rewards_v3" || txType.String == "rewards_v2" || txType.String == "rewards_v1" {

				importantDate := parseRewardV2(fields.String)

				count := make(map[string]int, 0)

				for _, v := range importantDate.Rewards {

					// key amount
					keyName := fmt.Sprintf("%v_amount", v.Type)
					if _, ok := count[v.Type]; ok {
						count[keyName] += v.Amount
					} else {
						count[keyName] = v.Amount
					}

					// increment if key exists
					if _, ok := count[v.Type]; ok {
						count[v.Type]++
					} else {
						count[v.Type] = 1
					}
				}

				// Add start and end epoch
				count[`start_epoch`] = int(importantDate.StartEpoch)
				count[`end_epoch`] = int(importantDate.EndEpoch)

				jsonString, _ := json.Marshal(count)

				// parse fields
				tx = append(tx, Transaction{"transaction", block.Int64, hash.String, txType.String, timestamp.Int64, string(jsonString)})

			} else {
				tx = append(tx, Transaction{"transaction", block.Int64, hash.String, txType.String, timestamp.Int64, fields.String})
			}

			if err := enc.Encode(tx); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&tx); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return tx
}

func getTransactionDataPagination(hash string, page int, limit int) RewardV2 {

	tx := RewardV2{}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("single-tx-%v-%v-%v", hash, page, limit)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow("SELECT block, hash, type, fields, time FROM transactions WHERE hash = $1", hash)

			var block, timestamp sql.NullInt64
			var hash, txType, fields sql.NullString

			err = row.Scan(&block, &hash, &txType, &fields, &timestamp)

			if err != nil {
				log.Printf("[ERROR GetSingleTransaction] %v", err)
			}

			// Parse spacific tx types here
			if txType.String == "rewards_v3" || txType.String == "rewards_v2" || txType.String == "rewards_v1" {

				importantDate := parseRewardV2(fields.String)

				tx.Hash = hash.String
				tx.Type = txType.String
				tx.EndEpoch = importantDate.EndEpoch
				tx.StartEpoch = importantDate.StartEpoch

				rewards := make([]SingleReward, 0)

				startPoint := page * limit
				currentPoint := 0
				for _, v := range importantDate.Rewards {
					if currentPoint >= startPoint && currentPoint < (startPoint+limit) {
						rewards = append(rewards, SingleReward{v.Type, v.Amount, v.Account, v.Gateway})
					}
					currentPoint++
				}

				tx.Rewards = rewards

			} else {
				return tx
			}
			if err := enc.Encode(tx); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&tx); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return tx
}

func parseRewardV2(fields string) RewardV2 {

	var reward RewardV2

	err := json.Unmarshal([]byte(fields), &reward)
	if err != nil {
		log.Println("[ERROR]", err)
	}

	return reward
}

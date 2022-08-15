package handlers

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"hntscan/db"
	"log"
	"math"
	"sort"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/labstack/echo/v4"
)

func GetOraclePrices(c echo.Context) error {

	var payload OraclePrices

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "oracle-prices"
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			// Get the first block
			startPoint := time.Now().AddDate(0, 0, -30).Unix()

			rows, err := db.DB.Query(`SELECT
										o.price,
										b.time
									FROM
										oracle_prices o
										INNER JOIN blocks b ON b.height = o.block
									WHERE
										b.time > $1`, startPoint)
			if err != nil {
				log.Printf("[ERROR GetOraclePrices] %v", err)
			}

			var price, time int64
			prices := make(map[int64]float64, 0)
			unsortedPrices := make(map[int64]float64, 0)

			max, min := 0.0, 9999.0
			prevPrice := 0.0

			defer rows.Close()

			for rows.Next() {
				err := rows.Scan(&price, &time)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				price := math.Floor((float64(price)/100000000)*100) / 100

				if price > max {
					max = price
				}

				if price < min {
					min = price
				}

				unsortedPrices[time] = price

			}
			rows.Close()

			// Sort by timestamp
			keys := make([]int, 0, len(unsortedPrices))

			for k := range unsortedPrices {
				keys = append(keys, int(k))
			}

			sort.Ints(keys)

			// Save sorted and remove consecutive duplicates
			for _, k := range keys {

				if prevPrice != unsortedPrices[int64(k)] {
					prevPrice = unsortedPrices[int64(k)]
					prices[int64(k)] = unsortedPrices[int64(k)]
				}
			}

			payload = OraclePrices{min, max, prices}

			if err := enc.Encode(payload); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&payload); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return c.JSON(200, payload)
}

func sortRewardsPerDay(rewards map[int64]int64) map[int64]int64 {

	combinedTimestamp := make(map[int64]int64, 0)

	for timestamp, rewards := range rewards {

		// Convert the timestamp to a date
		date := time.Unix(timestamp, 0)
		newDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Unix()

		// Check if key exists
		if _, ok := combinedTimestamp[newDate]; ok {
			combinedTimestamp[newDate] += rewards
		} else {
			combinedTimestamp[newDate] = rewards
		}

	}

	return combinedTimestamp
}

func getLast24HWalletRewards(hash string) map[int64]int64 {

	rewards := make(map[int64]int64, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallet-reward-24h-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			// Beginning of the day timestamp
			startTimestamp := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.UTC).Unix()

			// Get the first block
			rows, err := db.DB.Query(`SELECT amount, time FROM rewards WHERE account = $1 AND time > $2`, hash, startTimestamp)
			if err != nil {
				log.Printf("[ERROR getLast24HWalletRewards] %v", err)
			}

			defer rows.Close()

			var amount, time sql.NullInt64

			for rows.Next() {
				err := rows.Scan(&amount, &time)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				rewards[time.Int64] = amount.Int64
			}

			if err := enc.Encode(rewards); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})
		}
	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&rewards); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return rewards
}

func getLast24HHotspotRewards(hash string) map[int64]int64 {

	rewards := make(map[int64]int64, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-reward-24h-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			// Beginning of the day timestamp
			startTimestamp := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.UTC).Unix()

			// Get the first block
			rows, err := db.DB.Query(`SELECT amount, time FROM rewards WHERE gateway = $1 AND time > $2`, hash, startTimestamp)
			if err != nil {
				log.Printf("[ERROR getLast24HWalletRewards] %v", err)
			}

			defer rows.Close()

			var amount, time sql.NullInt64

			for rows.Next() {
				err := rows.Scan(&amount, &time)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				rewards[time.Int64] = amount.Int64
			}

			if err := enc.Encode(rewards); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})

		}
	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&rewards); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return rewards
}

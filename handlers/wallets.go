package handlers

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"hntscan/db"
	"log"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/labstack/echo/v4"
)

func GetWallets(c echo.Context) error {

	var wallets []WalletList

	page := c.QueryParam("page")

	if page == "" {
		page = "0"
	}

	offset, err := strconv.Atoi(page)
	if err != nil {
		fmt.Println(err)
	}

	limit := 25
	offset = offset * limit

	// memcached
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallets-%v", offset)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query(`SELECT
										address,
										dc_balance,
										security_balance,
										balance,
										staked_balance,
										mobile_balance,
										iot_balance,
										block
									FROM
										accounts
									ORDER BY
										block DESC
									LIMIT $1 OFFSET $2`, limit, offset)
			if err != nil {
				log.Println(err)
			}

			var dc_balance, security_balane, balance, staked_balance, mobile_balance, iot_balance, block sql.NullInt64
			var address sql.NullString

			defer rows.Close()

			for rows.Next() {

				err = rows.Scan(&address, &dc_balance, &security_balane, &balance, &staked_balance, &mobile_balance, &iot_balance, &block)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				walletHotspots := getWalletHotspotCount(address.String)

				walletValidators := getWalletValidatorCount(address.String)

				wallets = append(wallets, WalletList{
					DataType:       "wallet",
					Address:        address.String,
					HotspotCount:   walletHotspots,
					ValidatorCount: walletValidators,
					Balance: WalletBalance{
						DC:     dc_balance.Int64,
						HST:    security_balane.Int64,
						HNT:    balance.Int64,
						STAKE:  staked_balance.Int64,
						MOBILE: mobile_balance.Int64,
						IOT:    iot_balance.Int64,
					},
					LastBlock: block.Int64,
				})

			}

			if err := enc.Encode(wallets); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 300})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&wallets); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return c.JSON(200, wallets)
}

func GetSingleWallets(c echo.Context) error {

	wallet := c.Param("hash")

	walletHotspotsCount := getWalletHotspotCount(wallet)
	walletBalance := getWalletBalance(wallet)
	walletRewards := getWalletRewards(wallet, 30)
	walletRewards24H := getLast24HWalletRewards(wallet)
	walletValidatorsCount := getWalletValidatorCount(wallet)
	walletLastBlock := getWalletLastBlock(wallet)

	return c.JSON(200, Wallet{
		"wallet",
		wallet,
		walletHotspotsCount,
		walletBalance,
		walletRewards,
		walletRewards24H,
		walletValidatorsCount,
		walletLastBlock,
	})

}

func GetSingleWalletHotspots(c echo.Context) error {

	var walletHotspots = make([]Hotspot, 0)

	wallet := c.Param("hash")

	if wallet != "" {
		walletHotspots = getWalletHotspots(wallet)
	}

	return c.JSON(200, walletHotspots)
}

func GetSingleWalletValidators(c echo.Context) error {
	var walletValidators = make([]Validator, 0)

	wallet := c.Param("hash")

	if wallet != "" {
		walletValidators = getWalletValidators(wallet)
	}

	return c.JSON(200, walletValidators)
}

func getWalletData(wallet string) []Wallet {

	wallets := make([]Wallet, 0)

	wallethotspotsCount := getWalletHotspotCount(wallet)
	walletBalance := getWalletBalance(wallet)
	walletRewards := getWalletRewards(wallet, 30)
	walletRewards24H := getLast24HWalletRewards(wallet)
	walletValidatorsCount := getWalletValidatorCount(wallet)
	walletLastBlock := getWalletLastBlock(wallet)

	wallets = append(wallets, Wallet{
		"wallet",
		wallet,
		wallethotspotsCount,
		walletBalance,
		walletRewards,
		walletRewards24H,
		walletValidatorsCount,
		walletLastBlock,
	})

	return wallets
}

func getWalletHotspots(hash string) []Hotspot {

	hotspots := make([]Hotspot, 0)

	// memcached
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallet-hotspots-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query(`SELECT
										h.address,
										h.name,
										h.owner,
										h.last_poc_challenge,
										h.first_block,
										h.last_block,
										h.first_timestamp,
										h.nonce,
										h.reward_scale,
										h.elevation,
										h.gain,
										h.location,
										h.payer 
									FROM
										gateway_inventory h
									WHERE
										h.owner = $1`, hash)
			if err != nil {
				log.Printf("[ERROR getWalletHotspots A] %v", err)
			}

			defer rows.Close()

			var lastPocChallenge, firstBlock, lastBlock, nonce, elevation, gain sql.NullInt64
			var rewardScale sql.NullFloat64
			var address, name, owner, location, firstTimestamp, payer sql.NullString

			for rows.Next() {

				err = rows.Scan(&address, &name, &owner, &lastPocChallenge, &firstBlock, &lastBlock, &firstTimestamp, &nonce, &rewardScale, &elevation, &gain, &location, &payer)
				if err != nil {
					log.Printf("[ERROR getWalletHotspots B] %v", err)
				}

				firstTimestampInt := timestamptzConverter(firstTimestamp.String)
				maker, payer := getMaker(payer.String)
				geolocation := getGeolocationData(location.String)

				active := Active{false, 0, ""}
				hotspots = append(hotspots, Hotspot{
					"hotspot",
					address.String,
					name.String,
					owner.String,
					Location{
						location.String,
						geolocation.LongCountry,
						geolocation.ShortCountry,
						geolocation.LongCity,
						geolocation.LongStreet,
					},
					lastPocChallenge.Int64,
					firstBlock.Int64,
					lastBlock.Int64,
					firstTimestampInt,
					nonce.Int64,
					rewardScale.Float64,
					elevation.Int64,
					gain.Int64,
					maker,
					payer,
					active.Active,
					active.Timestamp,
					active.TX,
				})

			}
			rows.Close()

			if err := enc.Encode(hotspots); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&hotspots); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return hotspots
}

func getWalletHotspotCount(hash string) int {

	count := 0

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallet-hotspots-count-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow("SELECT COUNT(*) FROM gateway_inventory WHERE owner = $1", hash)

			var countData sql.NullInt64

			err := row.Scan(&countData)
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}

			count = int(countData.Int64)

			if err := enc.Encode(count); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&count); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return count
}

func getWalletValidators(hash string) []Validator {

	validators := make([]Validator, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallet-validators-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query(`SELECT
										i.address,
										i.name,
										s.online,
										i.version_heartbeat,
										i.status,
										i.last_heartbeat,
										i.penalty
									FROM
										validator_inventory i
										INNER JOIN validator_status s
										ON i.address = s.address
									WHERE
										i.owner = $1`, hash)
			if err != nil {
				log.Printf("[ERROR getWalletHotspots C] %v", err)
			}

			defer rows.Close()

			var address, name, online, staked sql.NullString
			var version_heartbeat, last_heartbeat sql.NullInt64
			var penalty sql.NullFloat64

			for rows.Next() {

				err = rows.Scan(&address, &name, &online, &version_heartbeat, &staked, &last_heartbeat, &penalty)
				if err != nil {
					log.Printf("[ERROR getWalletHotspots D] %v", err)
				}

				validators = append(validators, Validator{
					DataType:         "validator",
					Address:          address.String,
					Name:             name.String,
					Online:           online.String,
					VersionHeartbeat: version_heartbeat.Int64,
					LastHeartbeat:    last_heartbeat.Int64,
					Staked:           staked.String,
					PenaltyScore:     penalty.Float64,
				})

			}
			rows.Close()

			if err := enc.Encode(validators); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&validators); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return validators
}

func getWalletValidatorCount(hash string) int {

	count := 0

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallet-validator-count-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow("SELECT COUNT(*) FROM validator_inventory WHERE owner = $1", hash)

			var countData sql.NullInt64

			err := row.Scan(&countData)
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}

			count = int(countData.Int64)

			if err := enc.Encode(count); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&count); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return count
}

func getWalletBalance(hash string) WalletBalance {

	var walletBalance WalletBalance

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallet-balance-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			countQuery := db.DB.QueryRow(`SELECT count(*) FROM accounts WHERE address = $1`, hash)
			var count int
			countQuery.Scan(&count)
			if count > 0 {

				query, err := db.DB.Query(`SELECT block, dc_balance, security_balance, balance, staked_balance, mobile_balance, iot_balance FROM accounts WHERE address = $1`, hash)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				defer query.Close()

				var block, dc_balance, security_balance, balance, staked_balance, mobile_balance, iot_balance sql.NullInt64

				var currentBlock int64 = 0
				for query.Next() {
					err = query.Scan(&block, &dc_balance, &security_balance, &balance, &staked_balance, &mobile_balance, &iot_balance)
					if err != nil {
						log.Printf("[ERROR] %v", err)
					}

					// Get the latest block
					if block.Int64 > currentBlock {
						walletBalance = WalletBalance{security_balance.Int64, dc_balance.Int64, balance.Int64, staked_balance.Int64, mobile_balance.Int64, iot_balance.Int64}
					}
				}

			} else {
				walletBalance = WalletBalance{-1, -1, -1, -1, -1, -1}
			}

			if err := enc.Encode(walletBalance); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&walletBalance); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return walletBalance
}

func getWalletRewards(hash string, days int) map[int64]int64 {

	dayList := make(map[int64]int64, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallet-rewards-%v-%v", hash, days)
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			today := time.Now()
			startTimestamp := today.AddDate(0, 0, -days).Unix()

			query, err := db.DB.Query("SELECT block, time, amount, type FROM rewards WHERE account = $1 AND time >= $2", hash, startTimestamp)
			if err != nil {
				log.Println("[Error getSingleHotspotRewards] ", err)
			}

			var block, timestamp, amount sql.NullInt64
			var txType sql.NullString

			rewards := make(map[int64]int64, 0)

			defer query.Close()
			for query.Next() {

				query.Scan(&block, &timestamp, &amount, &txType)

				if amount.Valid {
					rewards[timestamp.Int64] = amount.Int64
				}
			}

			// Get first time of slice
			var firstTime int64 = 4121533591
			for d, _ := range rewards {
				if d < firstTime {
					firstTime = d
				}
			}

			// Generate empty array for dates
			year, month, day := time.Now().Date()
			todayDate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
			startDate := time.Unix(firstTime, 0)
			totalDays := int(todayDate.Sub(startDate).Hours() / 24)

			// Create empty array
			for i := 0; i < totalDays; i++ {
				day := todayDate.AddDate(0, 0, -i).Unix()
				dayList[day] = 0
			}

			// Fill dayList array with rewards
			combinedRewardsSort := sortRewardsPerDay(rewards)

			for k, _ := range dayList {
				dayList[k] = combinedRewardsSort[k]
			}

			if err := enc.Encode(dayList); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 3600})

		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&dayList); err != nil {
			log.Println("Error decode: ", err)
		}
	}
	return dayList
}

func getWalletLastBlock(hash string) int64 {

	var blockHeight int64 = 0

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("wallet-block-%v", hash)
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			query, err := db.DB.Query("SELECT block FROM accounts WHERE address = $1 ORDER BY block DESC", hash)
			if err != nil {
				log.Println("[Error getWalletLastBlock] ", err)
			}

			var block sql.NullInt64

			defer query.Close()
			for query.Next() {

				query.Scan(&block)

				if block.Valid {
					blockHeight = block.Int64
				}
			}

			if err := enc.Encode(blockHeight); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 3600})

		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&blockHeight); err != nil {
			log.Println("Error decode: ", err)
		}
	}
	return blockHeight
}

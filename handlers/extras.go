package handlers

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hntscan/db"
	"log"
	"net/http"
	"strconv"
	"time"
	"unicode"

	"github.com/bradfitz/gomemcache/memcache"
)

func GetStatsInventory() StatsInventory {

	var response StatsInventory

	// memcached
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "stats-inventory"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query("SELECT name, value FROM stats_inventory")
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}

			var value sql.NullInt64
			var name sql.NullString

			defer rows.Close()
			for rows.Next() {

				err := rows.Scan(&name, &value)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				switch name.String {
				case "blocks":
					response.Blocks = int(value.Int64)
				case "challenges":
					response.Challenges = int(value.Int64)
				case "cities":
					response.Cities = int(value.Int64)
				case "coingecko_price_eur":
					response.CoingeckoPriceEUR = int(value.Int64)
				case "coingecko_price_gbp":
					response.CoingeckoPriceGBP = int(value.Int64)
				case "coingecko_price_usd":
					response.CoingeckoPriceUSD = int(value.Int64)
				case "consensus_groups":
					response.ConsensusGroups = int(value.Int64)
				case "countries":
					response.Countries = int(value.Int64)
				case "hotspots":
					response.Hotspots = int(value.Int64)
				case "hotspots_dataonly":
					response.HotspotsDataOnly = int(value.Int64)
				case "hotspots_online":
					response.HotspotsOnline = int(value.Int64)
				case "ouis":
					response.OUIs = int(value.Int64)
				case "transactions":
					response.Transactions = int(value.Int64)
				case "validators":
					response.Validators = int(value.Int64)
				}

			}
			rows.Close()

			if err := enc.Encode(response); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})

		}

	} else {

		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&response); err != nil {
			log.Println("Error decode: ", err)
		}

	}

	return response
}

func GetVarsInventory() VarsInventory {

	var response VarsInventory

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "vars-inventory"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query("SELECT name, value FROM vars_inventory")
			if err != nil {
				log.Println("[ERROR] %v", err)
			}

			var name sql.NullString
			var value sql.NullString

			defer rows.Close()

			for rows.Next() {

				err := rows.Scan(&name, &value)
				if err != nil {
					log.Printf("[ERROR GetVarsInventory] %v", err)
				}

				switch name.String {

				case "block_time":
					blockTime, _ := strconv.Atoi(value.String)
					response.BlockTime = int64(blockTime)
				case "dc_payload_size":
					blockTime, _ := strconv.Atoi(value.String)
					response.DCPayloadSize = int64(blockTime)

				case "monthly_rewards":
					monthlyRewards, _ := strconv.Atoi(value.String)
					response.MonthlyRewards = int64(monthlyRewards)
				case "num_consensus_members":
					consensusNumber, _ := strconv.Atoi(value.String)
					response.ConsensusNumber = int64(consensusNumber)

				case "stake_withdrawal_cooldown":
					stakeWithdrawalCooldown, _ := strconv.Atoi(value.String)
					response.StakeWithdrawalCooldown = int64(stakeWithdrawalCooldown)

				case "validator_minimum_stake":
					validatorMinimumStake, _ := strconv.Atoi(value.String)
					response.ValidatorMinimumStake = int64(validatorMinimumStake)
				}
			}
			rows.Close()

			if err := enc.Encode(response); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})

		}
	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&response); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return response
}

func CoingeckoAPI() Coingecko {

	var cg Coingecko

	// memcached
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "coingecko-api"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			url := "https://api.coingecko.com/api/v3/coins/helium"

			// Build the request
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Println("NewRequest: ", err)
			}

			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				log.Println("Do: ", err)
			}

			defer resp.Body.Close()

			if err := json.NewDecoder(resp.Body).Decode(&cg); err != nil {
				log.Println(err)
			}

			if err := enc.Encode(cg); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 3600})

		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&cg); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return cg
}

func DcSpent(days int) int64 {

	var dcSpent int64

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "dc-spent"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			today := time.Now()
			previous := today.AddDate(0, 0, -days).Unix()

			row := db.DB.QueryRow("SELECT SUM(amount) FROM dc_burns WHERE time >= $1", previous)

			err := row.Scan(&dcSpent)
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}

			if err := enc.Encode(dcSpent); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 3600})

		}
	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&dcSpent); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return dcSpent
}

func GetMakersData() LastMaker {

	var response LastMaker

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "makers-data"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			var name, address sql.NullString

			rows, err := db.DB.Query("SELECT name, address FROM makers ORDER BY id ASC")
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}

			totalMakers := 0
			lastMakerName := ""
			lastMakerAddress := ""

			for rows.Next() {

				err = rows.Scan(&name, &address)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				totalMakers++
				lastMakerName = name.String
				lastMakerAddress = address.String

			}

			response = LastMaker{totalMakers, lastMakerName, lastMakerAddress}

			if err := enc.Encode(response); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 3600})

		}
	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&response); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return response
}

func GetLastOraclePrice() int {

	var response int

	// memcached
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "oracle-price"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow("SELECT price FROM oracle_prices ORDER BY block DESC LIMIT 1")

			err := row.Scan(&response)
			if err != nil {
				log.Printf("[ERROR] %v", err)
			}

			if err := enc.Encode(response); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 600})

		}
	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&response); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return response

}

// Support functions // --------------------------------------------------------

func timestamptzConverter(input string) int64 {

	t, err := time.Parse(time.RFC3339, input)
	if err != nil {
		return 0
	}
	return t.Unix()
}

func convertTimezoneToDay(startDate string) time.Time {

	t, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		log.Printf("Error parsing start date: %v", err)
	}

	year, month, day := t.Date()
	t2 := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	return t2
}

func convertTimezoneToDayString(startDate string) string {

	t, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		log.Printf("Error parsing start date: %v", err)
	}

	y, m, d := t.Date()
	month := fmt.Sprintf("%02d", int(m))
	day := fmt.Sprintf("%02d", int(d))

	return fmt.Sprintf("%v-%v-%v", y, month, day)
}

func isLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func isNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func uniqueSlice(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

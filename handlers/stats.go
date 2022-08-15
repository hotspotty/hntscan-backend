package handlers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"hntscan/db"
	"log"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/labstack/echo/v4"
)

// Homepage Stats: Overview
func GetStatsOverview(c echo.Context) error {

	start := time.Now()

	var response Stats

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "homepage-stats"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			statsInventory := GetStatsInventory()
			varsInventory := GetVarsInventory()
			hotspotTrend := HotspotTrend30Days()
			coingeckoData := CoingeckoAPI()
			dcSpend := DcSpent(30)
			lastHotspot := GetLastHotspot()
			makersData := GetMakersData()
			oraclePrice := GetLastOraclePrice()
			validatorData := GetValidatorList()

			// Parse validators
			onlineValidators := 0
			stakedValidators := 0
			validatorVersions := make(map[int64]int64, 0)

			for _, v := range validatorData {

				if v.Online == "online" && v.Staked == "staked" {
					onlineValidators++
				}

				if v.Staked == "staked" {
					stakedValidators++
				}

				if _, ok := validatorVersions[v.VersionHeartbeat]; ok {
					validatorVersions[v.VersionHeartbeat]++
				} else {
					validatorVersions[v.VersionHeartbeat] = 1
				}

			}

			validatorAPR := calculateValidatorAPR(onlineValidators)

			response = Stats{
				Hotspots:          Hotspots{int64(statsInventory.Hotspots), int64(statsInventory.HotspotsOnline), hotspotTrend},
				HNTPrice:          HNTPrice{coingeckoData.MarketData.CurrentPrice.Usd, coingeckoData.MarketData.PriceChangePercentage24H, oraclePrice},
				Height:            BlockHeight{statsInventory.Blocks, 0},
				DCSpent:           dcSpend,
				ValidatorCount:    ValidatorStats{int64(stakedValidators), varsInventory.ConsensusNumber, varsInventory.StakeWithdrawalCooldown, varsInventory.ValidatorMinimumStake, int64(onlineValidators), validatorVersions, validatorAPR},
				Challenges:        statsInventory.Challenges,
				OUICount:          statsInventory.OUIs,
				Countries:         statsInventory.Countries,
				Cities:            statsInventory.Cities,
				CirculatingSupply: coingeckoData.MarketData.CirculatingSupply,
				MarketCap:         coingeckoData.MarketData.MarketCap.Usd,
				MarketCapRank:     coingeckoData.MarketCapRank,
				LastHotspot:       lastHotspot,
				LastMaker:         makersData,
			}

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

	end := time.Now()
	fmt.Printf("/stats/overview/ - %v", end.Sub(start))

	return c.JSON(200, response)
}

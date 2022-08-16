package handlers

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hntscan/db"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/labstack/echo/v4"
	"github.com/uber/h3-go/v3"
)

func GetHotspots(c echo.Context) error {

	var hotspots []Hotspot

	page := c.QueryParam("page")

	if page == "" {
		page = "0"
	}

	offset, err := strconv.Atoi(page)
	if err != nil {
		log.Println(err)
	}

	limit := 25
	offset = offset * limit

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspots-%v", offset)
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
										l.location,
										l.long_country,
										l.long_city,
										l.long_street,
										l.short_country
									FROM
										gateway_inventory h
										INNER JOIN locations l ON l.location = h.location
									ORDER BY
										h.first_block DESC
									LIMIT $1 OFFSET $2`, limit, offset)
			if err != nil {
				log.Println(err)
			}

			var lastPocChallenge, firstBlock, lastBlock, nonce, elevation, gain sql.NullInt64
			var rewardScale sql.NullFloat64
			var address, name, owner, location, country, city, street, firstTimestamp, short_country sql.NullString

			defer rows.Close()

			for rows.Next() {

				err = rows.Scan(&address, &name, &owner, &lastPocChallenge, &firstBlock, &lastBlock, &firstTimestamp, &nonce, &rewardScale, &elevation, &gain, &location, &country, &city, &street, &short_country)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				firstTimestampInt := timestamptzConverter(firstTimestamp.String)

				maker, payer := getHotspotMaker(address.String)

				active := getHotspotStatus(address.String)

				hotspots = append(hotspots, Hotspot{
					"hotspot",
					address.String,
					name.String,
					owner.String,
					Location{
						location.String,
						country.String,
						short_country.String,
						city.String,
						street.String,
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

			if err := enc.Encode(hotspots); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&hotspots); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return c.JSON(200, hotspots)
}

func GetSingleHotspot(c echo.Context) error {

	hotspots := make([]SingleHotspot, 0)

	hash := c.Param("hash")

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-%v", hash)
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
										h.location
									FROM
										gateway_inventory h
									WHERE h.address = $1`, hash)
			if err != nil {
				log.Println(err)
			}

			var lastPocChallenge, firstBlock, lastBlock, nonce, elevation, gain sql.NullInt64
			var rewardScale sql.NullFloat64
			var address, name, owner, location, firstTimestamp sql.NullString

			defer rows.Close()

			for rows.Next() {

				err = rows.Scan(&address, &name, &owner, &lastPocChallenge, &firstBlock, &lastBlock, &firstTimestamp, &nonce, &rewardScale, &elevation, &gain, &location)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				firstTimestampInt := timestamptzConverter(firstTimestamp.String)
				maker, payer := getHotspotMaker(address.String)
				witnessCount := getHotspotWitnessesCount(address.String)
				geolocation := getGeolocationData(location.String)
				active := getHotspotStatus(address.String)

				hotspots = append(hotspots, SingleHotspot{
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
					witnessCount,
					active.Active,
					active.Timestamp,
					active.TX,
				})
			}

			if err := enc.Encode(hotspots); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&hotspots); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return c.JSON(200, hotspots)
}

func GetSingleHotspotActivities(c echo.Context) error {

	hash := c.Param("hash")

	page := c.QueryParam("page")

	if page == "" {
		page = "0"
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		log.Println(err)
	}

	limit := 5
	offset := pageInt * limit

	activities := getSingleHotspotActivities(hash, limit, offset)

	return c.JSON(200, ActivityResponsePayload{limit, pageInt, activities})
}

func GetSingleHotspotRewards(c echo.Context) error {

	hash := c.Param("hash")
	days := c.Param("days")

	if days == "" || hash == "" {
		return c.JSON(400, "Bad request")
	}

	daysInt, err := strconv.Atoi(days)
	if err != nil {
		log.Println(err)
	}

	rewards := getSingleHotspotRewards(hash, daysInt)
	rewards24h := getLast24HHotspotRewards(hash)

	intDay, err := strconv.Atoi(days)
	if err != nil {
		intDay = 30
	}

	payload := Reward{intDay, rewards, rewards24h}

	return c.JSON(200, payload)
}

func GetSingleHotspot24Rewards(hash string) map[int64]int64 {

	rewards := make(map[int64]int64, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-reward-24h-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			// Beginning of the day timestamp
			startTimestamp := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 0, 0, 0, 0, time.UTC).Unix()

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

func GetSingleHotspotStatus(c echo.Context) error {

	hash := c.Param("hash")

	status := getHotspotStatus(hash)

	return c.JSON(200, status)
}

func GetMultipleHotspotStatus(c echo.Context) error {

	type payload struct {
		HotspotIDs []string `json:"hotspots"`
	}

	type response struct {
		HotspotID string `json:"hotspot_id"`
		Active    Active `json:"active"`
	}

	var responsePayload []response

	res := new(payload)
	if err := c.Bind(res); err != nil {
		log.Printf("%v", err)
		return err
	}

	for _, hotspot := range res.HotspotIDs {
		status := getHotspotStatus(hotspot)
		responsePayload = append(responsePayload, response{hotspot, status})
	}

	return c.JSON(200, responsePayload)

}

func GetSingleHotspotAvgBeacons(c echo.Context) error {

	hash := c.Param("hash")

	sevenDayBeacon := getSingleHotspotBeacons(hash)

	return c.JSON(200, SevenDayAvgBeacons{sevenDayBeacon})
}

func HotspotTrend30Days() HotspotTrend {

	stats := GetStatsInventory()
	var response HotspotTrend

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "hotspot-30-day-trend"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			last30Days := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)

			rows, err := db.DB.Query("SELECT address, first_timestamp FROM gateway_inventory WHERE first_timestamp >= $1", last30Days)
			if err != nil {
				log.Printf("[ERROR]1 %v", err)
			}

			var address, firstTimestamp sql.NullString

			defer rows.Close()

			hotspotCount := make(map[string]int, 0)
			totalHotspots := stats.Hotspots
			totalCount := 0

			for rows.Next() {

				err := rows.Scan(&address, &firstTimestamp)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				totalCount++

				day := convertTimezoneToDayString(firstTimestamp.String)

				if _, ok := hotspotCount[day]; ok {
					hotspotCount[day]++
				} else {
					hotspotCount[day] = 1
				}
			}
			rows.Close()

			// Sort
			keys := make([]string, 0, len(hotspotCount))
			for k := range hotspotCount {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			hotspotCountUpdated := make(map[string]int, 0)
			min := totalHotspots - totalCount

			value := min

			for _, k := range keys {

				value = value + hotspotCount[k]
				hotspotCountUpdated[k] = value

			}

			response = HotspotTrend{hotspotCountUpdated, min, totalHotspots}

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

func GetLastHotspot() LastHotspot {

	var response LastHotspot

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := "last-hotspot"
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow(`SELECT
										h.address,
										h.name,
										h.location,
										l.long_country,
										l.short_country
									FROM
										gateway_inventory h
										INNER JOIN locations l ON h.location = l.location
									WHERE
										h.nonce > 0
									ORDER BY
										h.first_block DESC
									LIMIT 1`)

			var address, name, location, long_country, short_country sql.NullString

			err := row.Scan(&address, &name, &location, &long_country, &short_country)
			if err != nil {
				log.Printf("[ERROR]3 %v", err)
			}

			response = LastHotspot{
				address.String,
				name.String,
				location.String,
				long_country.String,
				short_country.String,
			}

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

func getSingleHotspotActivities(address string, limit int, offset int) ActivityResponsePayloadData {

	var response ActivityResponsePayloadData

	witnessData := make([]WitnessParsed, 0)       // received a beacon
	challengerData := make([]ChallengerParsed, 0) // generated a challenge
	rewardData := make([]RewardParsed, 0)         // rewards
	challengeeData := make([]ChallengeeParsed, 0) // submit beacon
	dataPacketData := make([]DataPacketParsed, 0) // data packets
	gatewayData := make([]GatewayParsed, 0)       // gateway data

	// Activity limit
	limitStr := fmt.Sprintf("%v", limit)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-activity-%v-%v", address, offset)
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query(`SELECT ta.actor_role, t.type, t.hash, t.time, t.block, t.fields
			FROM transaction_actors AS ta
			INNER JOIN transactions AS t
			ON ta.transaction_hash = t.hash
			WHERE actor = $1 ORDER BY block DESC LIMIT $2 OFFSET $3`, address, limitStr, offset*limit)

			if err != nil {
				log.Printf("[ERROR]4 %v", err)
			}

			defer rows.Close()

			var actor_role, txType, hash, fields sql.NullString
			var timeField, block sql.NullInt64

			for rows.Next() {

				err := rows.Scan(&actor_role, &txType, &hash, &timeField, &block, &fields)
				if err != nil {
					log.Printf("[ERROR]5 %v", err)
				}

				// Data
				if actor_role.String == "packet_receiver" {
					dataPacket := new(DataPacket)
					json.Unmarshal([]byte(fields.String), &dataPacket)

					for _, summary := range dataPacket.StateChannel.Summaries {
						if summary.Client == address {

							dataPacketData = append(dataPacketData, DataPacketParsed{
								hash.String,
								timeField.Int64,
								block.Int64,
								summary.NumDcs,
								summary.Location,
								summary.NumPackets,
							})

						}
					}
				}

				// Rewards
				if actor_role.String == "reward_gateway" {
					rewardList := new(RewardStruct)
					json.Unmarshal([]byte(fields.String), &rewardList)

					for _, reward := range rewardList.Rewards {
						if reward.Gateway == address {
							rewardData = append(rewardData, RewardParsed{hash.String, timeField.Int64, block.Int64, reward.Amount})
						}
					}
				}

				// Challengee
				if actor_role.String == "challengee" {

					allWitnesses := make([]WitnessData, 0)
					challengeeList := new(ChallengeeStruct)
					json.Unmarshal([]byte(fields.String), &challengeeList)

					validBeaconCount := 0
					invalidBeaconCount := 0

					beaconer := challengeeList.Path[0].Receipt.Gateway
					beaconerDataTemp := getHotspotData(beaconer)

					var beaconerData HotspotStruct
					if len(beaconerDataTemp) == 1 {
						beaconerData = beaconerDataTemp[0]
					}

					challengerDataTemp := getHotspotData(challengeeList.Challenger)

					var challengerData HotspotStruct
					if len(challengerDataTemp) == 1 {
						challengerData = challengerDataTemp[0]
					}

					for _, witness := range challengeeList.Path[0].Witnesses {

						invalidReason := ""
						if witness.IsValid {
							validBeaconCount++
						} else {
							invalidBeaconCount++
							invalidReason = witness.InvalidReason
						}

						distance := h3.PointDistM(h3.ToGeo(h3.FromString(beaconerData.Location)), h3.ToGeo(h3.FromString(witness.Location)))
						singleWitnessDataTemp := getHotspotData(witness.Gateway)

						var singleWitnessData HotspotStruct
						if len(singleWitnessDataTemp) == 1 {
							singleWitnessData = singleWitnessDataTemp[0]
						}

						// Append to the list of hotspot witnesses
						allWitnesses = append(allWitnesses, WitnessData{
							witness.Gateway,
							int(distance),
							witness.Datarate,
							witness.Signal,
							witness.Snr,
							witness.Frequency,
							witness.IsValid,
							witness.Timestamp,
							witness.Channel,
							singleWitnessData.Place,
							invalidReason,
						})

					}

					challengeeData = append(challengeeData, ChallengeeParsed{
						hash.String,
						timeField.Int64,
						block.Int64,
						challengeeList.Challenger,
						challengerData.Place,
						beaconer,
						beaconerData.Place,
						validBeaconCount,
						invalidBeaconCount,
						allWitnesses,
					})
				}

				// Challenger
				if actor_role.String == "challenger" {

					allWitnesses := make([]WitnessData, 0)
					challengerList := new(ChallengerStruct)

					json.Unmarshal([]byte(fields.String), &challengerList)

					if len(challengerList.Path) > 0 {

						validWitnessCount := 0
						invalidWitnessCount := 0

						beaconer := challengerList.Path[0].Receipt.Gateway

						challengerLocationTemp := getHotspotData(challengerList.Challenger)
						var challengerLocation HotspotStruct
						if len(challengerLocationTemp) == 1 {
							challengerLocation = challengerLocationTemp[0]
						}

						beaconerLocationTemp := getHotspotData(beaconer)
						var beaconerLocation HotspotStruct
						if len(beaconerLocationTemp) == 1 {
							beaconerLocation = beaconerLocationTemp[0]
						}

						for _, witness := range challengerList.Path[0].Witnesses {

							// Add counts
							invalidReason := ""
							if witness.IsValid {
								validWitnessCount++
							} else {
								invalidWitnessCount++
								invalidReason = witness.InvalidReason
							}

							distance := h3.PointDistM(h3.ToGeo(h3.FromString(beaconerLocation.Location)), h3.ToGeo(h3.FromString(witness.Location)))

							singleWitnessDataTemp := getHotspotData(witness.Gateway)
							var singleWitnessData HotspotStruct
							if len(singleWitnessDataTemp) == 1 {
								singleWitnessData = singleWitnessDataTemp[0]
							}

							// Append to the list of hotspot witnesses
							allWitnesses = append(allWitnesses, WitnessData{
								witness.Gateway,
								int(distance),
								witness.Datarate,
								witness.Signal,
								witness.Snr,
								witness.Frequency,
								witness.IsValid,
								witness.Timestamp,
								witness.Channel,
								singleWitnessData.Place,
								invalidReason,
							})
						}

						challengerData = append(challengerData, ChallengerParsed{
							hash.String,
							timeField.Int64,
							block.Int64,
							challengerList.Challenger,
							challengerLocation.Place,
							beaconer,
							beaconerLocation.Place,
							validWitnessCount,
							invalidWitnessCount,
							allWitnesses,
						})

					}
				}

				// Witness Data Parsing (This hotspot receied a beacon from someone else)
				if actor_role.String == "witness" {

					allWitnesses := make([]WitnessData, 0)
					witnessList := new(WitnessStruct)
					json.Unmarshal([]byte(fields.String), &witnessList)

					// Count the number of valid and invalid witnesses
					validWitnessCount := 0
					invalidWitnessCount := 0
					isValidWitness := false
					thisDistance := 0
					var challengerData HotspotStruct
					var beaconerData HotspotStruct

					beaconer := witnessList.Path[0].Challengee

					for _, witness := range witnessList.Path[0].Witnesses {

						// Add counts
						invalidReason := ""
						if witness.IsValid {
							validWitnessCount++
						} else {
							invalidWitnessCount++
							invalidReason = witness.InvalidReason
						}

						// Get information for this single hotspot
						if witness.Gateway == address {
							isValidWitness = witness.IsValid

							// challengerDataTemp := getHotspotData(witnessList.Challenger)

							// var challengerData HotspotStruct
							// if len(challengerDataTemp) == 1 {
							// 	challengerData = challengerDataTemp[0]
							// }

							beaconerDataTemp := getHotspotData(beaconer)
							var beaconerData HotspotStruct
							if len(beaconerDataTemp) == 1 {
								beaconerData = beaconerDataTemp[0]
							}

							thisDistance = int(h3.PointDistM(h3.ToGeo(h3.FromString(beaconerData.Location)), h3.ToGeo(h3.FromString(witness.Location))))

						}

						distance := h3.PointDistM(h3.ToGeo(h3.FromString(witnessList.Path[0].ChallengeeLocation)), h3.ToGeo(h3.FromString(witness.Location)))

						singleWitnessDataTemp := getHotspotData(witness.Gateway)
						var singleWitnessData HotspotStruct
						if len(singleWitnessDataTemp) == 1 {
							singleWitnessData = singleWitnessDataTemp[0]
						}

						// Append to the list of hotspot witnesses
						allWitnesses = append(allWitnesses, WitnessData{
							witness.Gateway,
							int(distance),
							witness.Datarate,
							witness.Signal,
							witness.Snr,
							witness.Frequency,
							witness.IsValid,
							witness.Timestamp,
							witness.Channel,
							singleWitnessData.Place,
							invalidReason,
						})

					}

					witnessData = append(witnessData, WitnessParsed{
						witnessList.Hash,
						timeField.Int64,
						block.Int64,
						thisDistance,
						witnessList.Challenger,
						challengerData.Place,
						beaconer,
						beaconerData.Place,
						isValidWitness,
						validWitnessCount,
						invalidWitnessCount,
						allWitnesses,
					})
				}

				// Assertion
				if actor_role.String == "gateway" {

					gatewayDataSingle := new(GatewayParsed)
					json.Unmarshal([]byte(fields.String), &gatewayDataSingle)

					gatewayDataSingle.Timestamp = timeField.Int64

					gatewayData = append(gatewayData, *gatewayDataSingle)

				}

			}

			response = ActivityResponsePayloadData{witnessData, challengerData, challengeeData, rewardData, dataPacketData, gatewayData}

			if err := enc.Encode(response); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 7200})

		}
	} else {

		var tempResponse ActivityResponsePayloadData

		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&tempResponse); err != nil {
			log.Println("Error decode: ", err)
		}

		tempWitnesses := []WitnessParsed{}
		if len(tempResponse.Witnesses) != 0 {
			tempWitnesses = tempResponse.Witnesses
		}

		tempChallengers := []ChallengerParsed{}
		if len(tempResponse.Challengers) != 0 {
			tempChallengers = tempResponse.Challengers
		}

		tempChallengees := []ChallengeeParsed{}
		if len(tempResponse.Challengees) != 0 {
			tempChallengees = tempResponse.Challengees
		}

		tempRewards := []RewardParsed{}
		if len(tempResponse.Rewards) != 0 {
			tempRewards = tempResponse.Rewards
		}

		tempDataPackets := []DataPacketParsed{}
		if len(tempResponse.DataPackets) != 0 {
			tempDataPackets = tempResponse.DataPackets
		}

		tempGatewayPackets := []GatewayParsed{}
		if len(tempResponse.GatewayData) != 0 {
			tempGatewayPackets = tempResponse.GatewayData
		}

		response = ActivityResponsePayloadData{tempWitnesses, tempChallengers, tempChallengees, tempRewards, tempDataPackets, tempGatewayPackets}

	}

	return response
}

func getSingleHotspotBeacons(address string) int {

	responseData := 0
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-activity-7days-%v", address)
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			sevenDaysAgo := time.Now().AddDate(0, 0, -7)

			rows, err := db.DB.Query(`SELECT
										ta.actor_role,
										t.type,
										t.hash,
										t.time,
										t.block,
										t.fields
									FROM
										transaction_actors AS ta
										INNER JOIN transactions AS t ON ta.transaction_hash = t.hash
									WHERE
										actor = $1
										AND t.time > $2
									ORDER BY
										block`, address, sevenDaysAgo.Unix())

			if err != nil {
				log.Printf("[ERROR]7 %v", err)
			}

			defer rows.Close()

			var actor_role, txType, hash, fields sql.NullString
			var timeField, block sql.NullInt64

			// witnessCount := 0
			beaconCount := 0

			for rows.Next() {

				err := rows.Scan(&actor_role, &txType, &hash, &timeField, &block, &fields)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				// Challengee
				if actor_role.String == "challengee" {
					beaconCount++
				}

				// Challenger
				if actor_role.String == "challenger" {
					beaconCount++
				}

				// Witness Data Parsing (This hotspot receied a beacon from someone else)
				if actor_role.String == "witness" {
					beaconCount++
				}

			}

			responseData = beaconCount

			if err := enc.Encode(responseData); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60 * 60 * 24})

		}
	} else {

		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&responseData); err != nil {
			log.Println("Error decode: ", err)
		}

	}

	return responseData
}

// getHotspotData returns a single hotspot data (cached)
func getHotspotData(hash string) []HotspotStruct {

	returnStruct := make([]HotspotStruct, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("get-hotspot-data-%v", hash)
	hotspotDataCache, err := db.MC.Get(cacheName)
	if err != nil {
		if err == memcache.ErrCacheMiss {

			var address, owner, location, name, first_timestamp, peer_timestamp, online, listen_addrs, updated_at, payer, mode sql.NullString
			var last_block, first_block, last_poc_challenge, nonce, elevation, gain, last_assertion sql.NullInt64
			var reward_scale sql.NullFloat64

			row := db.DB.QueryRow(`SELECT
									gi.address,
									gi.owner,
									gi.name,
									gi.location,
									gs.block,
									gi.first_block,
									gi.last_poc_challenge,
									gi.nonce,
									gi.reward_scale,
									gi.first_timestamp,
									gs.peer_timestamp,
									gs.online,
									gs.listen_addrs,
									gi.elevation, 
									gi.gain,
									gs.updated_at,
									gab.timestamp,
									gi.payer,
									gi.mode
								FROM
									gateway_inventory gi
									INNER JOIN gateway_status gs ON gi.address = gs.address
									INNER JOIN gateway_assertion_blocks gab ON gi.address = gab.address
								WHERE
									gi.address = $1;`, hash)

			err := row.Scan(&address, &owner, &name, &location, &last_block, &first_block, &last_poc_challenge, &nonce, &reward_scale, &first_timestamp, &peer_timestamp, &online, &listen_addrs, &elevation, &gain, &updated_at, &last_assertion, &payer, &mode)

			if err != nil {
				if err != sql.ErrNoRows {
					log.Printf("[ERROR]10: %v ", err)
				}
			}

			place := calculatePlace(location.String)

			lastUpdate := timestamptzConverter(updated_at.String)
			timestampAdded, _ := dateparse.ParseLocal(first_timestamp.String)
			lastAssetion := int(last_assertion.Int64)
			maker, _ := getHotspotMaker(hash)

			payerAddress := ""
			if payer.Valid {
				payerAddress = payer.String
			}

			active := getHotspotStatus(address.String)

			returnStruct = append(returnStruct, HotspotStruct{
				"hotspot",
				address.String,
				int(last_block.Int64),
				int(first_block.Int64),
				place,
				int(last_poc_challenge.Int64),
				location.String,
				name.String,
				int(nonce.Int64),
				owner.String,
				reward_scale.Float64,
				int(timestampAdded.Unix()),
				int(elevation.Int64),
				int(gain.Int64),
				int(lastUpdate),
				"hotspot",
				lastAssetion,
				payerAddress,
				mode.String,
				maker,
				active.Active,
				active.Timestamp,
				active.TX,
			})

			if err := enc.Encode(returnStruct); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 3600})

		}
	} else {

		bufDecode := bytes.NewBuffer(hotspotDataCache.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&returnStruct); err != nil {
			log.Println("Error decode: ", err)
		}

	}

	return returnStruct
}

func calculatePlace(location string) string {

	g := getGeolocationData(location)

	if g.LongCity == "" && g.ShortState == "" && g.LongCountry == "" {
		return "Unknown location"
	}

	var locationTerms []string
	if g.LongCity != "" {
		locationTerms = append(locationTerms, g.LongCity)
	}

	if g.ShortState != "" {
		locationTerms = append(locationTerms, g.ShortState)
	}

	if g.LongCountry != "" {
		locationTerms = append(locationTerms, g.LongCountry)
	}

	finalString := strings.Join(locationTerms, ", ")

	if finalString == "" {
		return "Unknown location"
	} else {
		return finalString
	}
}

func getGeolocationData(location string) GeoCode {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("geolocation-data-%v", location)
	cacheData, err := db.MC.Get(cacheName)

	var newGeoCode GeoCode

	if err != nil {
		if err == memcache.ErrCacheMiss {

			var long_street, short_street, long_city, short_city, long_state, short_state, long_country, short_country, city_id sql.NullString

			row := db.DB.QueryRow("SELECT  long_street, short_street, long_city, short_city, long_state, short_state, long_country, short_country, city_id FROM locations WHERE location = $1", location)
			err := row.Scan(&long_street, &short_street, &long_city, &short_city, &long_state, &short_state, &long_country, &short_country, &city_id)

			if err != nil {
				if err == sql.ErrNoRows {
					// Return empty struct
					return GeoCode{"", "", "", "", "", "", "", "", ""}

				} else {
					log.Printf("[ERROR]11 error: %v ", err)
				}
			}

			newGeoCode = GeoCode{
				short_street.String,
				short_state.String,
				short_country.String,
				short_city.String,
				long_street.String,
				long_state.String,
				long_country.String,
				long_city.String,
				city_id.String,
			}

			if err := enc.Encode(newGeoCode); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 3600})
		}
	} else {

		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&newGeoCode); err != nil {
			log.Println("Error decode: ", err)
		}
	}
	return newGeoCode
}

func getSingleHotspotRewards(hash string, days int) map[int64]int64 {

	rewards := make(map[int64]int64, 0)
	rewardsSorted := make(map[int64]int64, 0)
	dayList := make(map[int64]int64, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-rewards-%v-%v", hash, days)
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			today := time.Now()
			startTimestamp := today.AddDate(0, 0, -days).Unix()

			query, err := db.DB.Query("SELECT block, time, amount, type FROM rewards WHERE gateway = $1 AND time >= $2", hash, startTimestamp)
			if err != nil {
				log.Println("[Error getSingleHotspotRewards] ", err)
			}

			var block, timestamp, amount sql.NullInt64
			var txType sql.NullString

			defer query.Close()
			for query.Next() {

				query.Scan(&block, &timestamp, &amount, &txType)

				if amount.Valid {
					rewards[timestamp.Int64] = amount.Int64
				}
			}

			// Sort by timestamp
			keys := make([]int, 0, len(rewards))

			for k := range rewards {
				keys = append(keys, int(k))
			}

			sort.Ints(keys)

			// Save sorted
			for _, k := range keys {
				rewardsSorted[int64(k)] = rewards[int64(k)]
			}

			// Get first time of slice
			var firstTime int64 = 4121533591
			for d, _ := range rewardsSorted {
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
			rewardsSortedSort := sortRewardsPerDay(rewardsSorted)

			for k, _ := range dayList {
				dayList[k] = rewardsSortedSort[k]
			}

			if err := enc.Encode(dayList); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
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

func getHotspotMaker(hash string) (string, string) {

	makerName := Maker{}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-maker-%v", hash)
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow(`SELECT
									m.name,
									m.address
								FROM
									makers m
									INNER JOIN gateway_inventory g ON g.payer = m.address
								WHERE
									g.address = $1`, hash)

			var maker, payer sql.NullString

			row.Scan(&maker, &payer)

			if maker.Valid && maker.String != "" {
				makerName = Maker{maker.String, payer.String}
			} else {
				makerName = Maker{"", ""}
			}

			if err := enc.Encode(makerName); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 24 * 60 * 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&makerName); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return makerName.Name, makerName.Payer
}

func getMaker(payer string) (string, string) {

	makerName := Maker{}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-maker-payer-%v", payer)
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow(`SELECT
									m.name,
									m.address
								FROM
									makers m	
								WHERE
									m.address = $1`, payer)

			var maker, payer sql.NullString

			row.Scan(&maker, &payer)

			if maker.Valid && maker.String != "" {
				makerName = Maker{maker.String, payer.String}
			} else {
				makerName = Maker{"", ""}
			}

			if err := enc.Encode(makerName); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 24 * 60 * 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&makerName); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return makerName.Name, makerName.Payer
}

func getHotspotWitnessesCount(hotspotID string) int {

	type combinedHotspots struct {
		Valid   []HotspotsValid   `json:"valid"`
		Invalid []HotspotsInvalid `json:"invalid"`
	}

	type hotspotStruct struct {
		HotspotID string           `json:"id"`
		Beacon    combinedHotspots `json:"witnessedBy"`
		Witness   combinedHotspots `json:"witnessedOthers"`
	}

	type returnStruct struct {
		Object string        `json:"object"`
		Url    string        `json:"url"`
		Data   hotspotStruct `json:"data"`
	}

	type receipts struct {
		SNR       float64 `json:"snr"`
		Data      string  `json:"data"`
		Origin    string  `json:"origin"`
		Signal    float64 `json:"signal"`
		Channel   int     `json:"channel"`
		Gateway   string  `json:"gateway"`
		DataRate  string  `json:"datarate"`
		Frequency float64 `json:"frequency"`
		Timestamp int64   `json:"timestamp"`
	}

	type witness struct {
		SNR           float64 `json:"snr"`
		Owner         string  `json:"owner"`
		Signal        int     `json:"signal"`
		Channel       int     `json:"channel"`
		Gateway       string  `json:"gateway"`
		DataRate      string  `json:"datarate"`
		IsValid       bool    `json:"is_valid"`
		Location      string  `json:"location"`
		Frequency     float64 `json:"frequency"`
		Timestamp     int64   `json:"timestamp"`
		PacketHash    string  `json:"packet_hash"`
		InvalidReason string  `json:"invalid_reason"`
	}

	type path struct {
		Receipt            []receipts `json:"receipt"`
		Witness            []witness  `json:"witnesses"`
		Challengee         string     `json:"challengee"`
		ChallengeeOwner    string     `json:"challengee_owner"`
		ChallengeeLocation string     `json:"challengee_location"`
	}

	type field struct {
		Fee                int    `json:"fee"`
		Hash               string `json:"hash"`
		Path               []path `json:"path"`
		Type               string `json:"type"`
		Secret             string `json:"secret"`
		Challenger         string `json:"challenger"`
		OnionKeyHash       string `json:"onion_key_hash"`
		ChallengerOwner    string `json:"challenger_owner"`
		RequestBlockHash   string `json:"request_block_hash"`
		ChallengerLocation string `json:"challenger_location"`
	}

	sevenDays := time.Now().AddDate(0, 0, -7).Unix()

	var block, timestamp sql.NullInt64
	var fields, actor_role sql.NullString

	rows, err := db.DB.Query(`SELECT
								a.block,
								time,
								fields,
								a.actor_role
							FROM
								transaction_actors a
								INNER JOIN transactions t ON a.transaction_hash = t.hash
							WHERE
								a.actor = $1
								AND(a.actor_role = 'witness'
									OR a.actor_role = 'challengee')
								AND t. time >= $2
							ORDER BY
								a.block ASC;`, hotspotID, sevenDays)

	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	defer rows.Close()

	witnessValidList := make([]ValidStruct, 0)

	i := 0
	for rows.Next() {
		i++
		err := rows.Scan(&block, &timestamp, &fields, &actor_role)
		if err != nil {
			log.Printf("[ERROR]13 %v", err)
		}

		var parsed field
		json.Unmarshal([]byte(fields.String), &parsed)

		if actor_role.String == "witness" {
			for _, paths := range parsed.Path {

				for _, witness := range paths.Witness {

					// Only calculate for this hotspot
					if witness.Gateway == hotspotID {

						// Check if valid or invalid
						if witness.IsValid {
							// Add this TX as a valid transaction
							witnessValidList = append(witnessValidList, ValidStruct{paths.Challengee})
						}
					}
				}
			}
		}

	}
	rows.Close()

	parsedValidWitness := parseValid(witnessValidList)
	witnessOthersList := []string{}

	for _, wOther := range parsedValidWitness {
		witnessOthersList = append(witnessOthersList, wOther.Address)
	}

	return len(uniqueSlice(witnessOthersList))
}

func parseValid(validList []ValidStruct) []HotspotsValid {
	validHotspots := make([]HotspotsValid, 0)
	temporaryList := make(map[string]int, 0)

	// Sort values
	for _, hotspot := range validList {
		if _, ok := temporaryList[hotspot.Address]; ok {
			temporaryList[hotspot.Address] += 1
		} else {
			temporaryList[hotspot.Address] = 1
		}
	}

	// Build return list
	for key, value := range temporaryList {
		validHotspots = append(validHotspots, HotspotsValid{key, value})
	}

	return validHotspots
}

func parseInvalid(invalidList []InvalidStruct) []HotspotsInvalid {

	temporaryInvalidList := make(map[string][]string, 0)

	invalidHotspotsResponse := make([]HotspotsInvalid, 0)

	// Build array
	for _, hotspot := range invalidList {
		temporaryInvalidList[hotspot.Address] = append(temporaryInvalidList[hotspot.Address], hotspot.Reason)
	}

	// Count
	for id, reason := range temporaryInvalidList {

		totalInvalids := 0
		temporaryList := make(map[string]int, 0)

		for _, r := range reason {
			totalInvalids++
			if _, ok := temporaryList[r]; ok {
				temporaryList[r] += 1
			} else {
				temporaryList[r] = 1
			}
		}

		temporaryInvalidReasons := make([]InvalidReasons, 0)

		for reas, count := range temporaryList {
			temporaryInvalidReasons = append(temporaryInvalidReasons, InvalidReasons{reas, count})
		}

		invalidHotspotsResponse = append(invalidHotspotsResponse, HotspotsInvalid{id, totalInvalids, temporaryInvalidReasons})
	}

	return invalidHotspotsResponse
}

func getHotspotDataByName(query string) []HotspotSearch {

	hotspots := make([]HotspotSearch, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("search-hotspots-%v", query)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			searchName := strings.ToLower(query)
			searchName = strings.Replace(searchName, " ", "%", 3)
			searchName = strings.Replace(searchName, "-", "%", 3)
			searchName = strings.Replace(searchName, "_", "%", 3)

			searchNameArray := strings.Split(searchName, "%")

			var queryNameArray []string
			for _, n := range searchNameArray {
				queryNameArray = append(queryNameArray, "'%"+n+"%'")
			}

			queryString := `SELECT
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
				l.location,
				l.long_country,
				l.long_city,
				l.long_street,
				l.short_country
				FROM
				gateway_inventory h
			INNER JOIN locations l ON l.location = h.location
			WHERE
		h.name LIKE ALL(array[`

			totalWords := len(queryNameArray)
			k := 1
			for _, val := range queryNameArray {

				if k < totalWords {
					queryString += val + ","
					k++
				} else {
					queryString += val
				}
			}

			queryString += `]);`

			rows, err := db.DB.Query(queryString)
			if err != nil {
				log.Printf("[ERROR]a %v", err)
			}

			defer rows.Close()

			var lastPocChallenge, firstBlock, lastBlock, nonce, elevation, gain sql.NullInt64
			var rewardScale sql.NullFloat64
			var address, name, owner, location, country, city, street, firstTimestamp, short_country sql.NullString

			for rows.Next() {

				err = rows.Scan(&address, &name, &owner, &lastPocChallenge, &firstBlock, &lastBlock, &firstTimestamp, &nonce, &rewardScale, &elevation, &gain, &location, &country, &city, &street, &short_country)
				if err != nil {
					log.Printf("[ERROR]b %v", err)
				}

				if address.Valid && name.Valid && owner.Valid && firstTimestamp.Valid {

					firstTimestampInt := timestamptzConverter(firstTimestamp.String)

					hotspots = append(hotspots, HotspotSearch{
						"hotspot",
						address.String,
						name.String,
						owner.String,
						Location{
							location.String,
							country.String,
							short_country.String,
							city.String,
							street.String,
						},
						lastPocChallenge.Int64,
						firstBlock.Int64,
						lastBlock.Int64,
						firstTimestampInt,
						nonce.Int64,
						rewardScale.Float64,
						elevation.Int64,
						gain.Int64,
					})

				}
			}
			rows.Close()

			if err := enc.Encode(hotspots); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 300})

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

func getHotspotStatus(hash string) Active {

	activity := Active{false, 0, "none"}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("hotspot-status-%v", hash)
	cacheData, err := db.MC.Get(cacheName)

	if err != nil {
		if err == memcache.ErrCacheMiss {

			var delayTime int32 = 60

			row := db.DB.QueryRow(`SELECT block, transaction_hash FROM transaction_actors WHERE actor = $1 ORDER BY block DESC LIMIT 1`, hash)

			var block sql.NullInt64
			var tx sql.NullString

			row.Scan(&block, &tx)

			if block.Valid && block.Int64 > 0 {

				row := db.DB.QueryRow(`SELECT time FROM blocks WHERE height = $1`, block.Int64)
				var timestamp sql.NullInt64
				row.Scan(&timestamp)

				if timestamp.Valid && timestamp.Int64 != 0 {
					// Check if delta is less than 36 hours
					delta := time.Now().Unix() - timestamp.Int64
					if delta < 36*60*60 {
						delayTime = 60 * 10 // in case true, increase delay time to 10 minutes
						activity = Active{true, timestamp.Int64, tx.String}
					} else {
						activity = Active{false, timestamp.Int64, tx.String}
					}
				}
			}

			if err := enc.Encode(activity); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: delayTime})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&activity); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return activity
}

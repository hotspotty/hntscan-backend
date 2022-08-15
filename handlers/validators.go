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
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/cznic/mathutil"
	"github.com/labstack/echo/v4"
)

func GetValidators(c echo.Context) error {

	validators := make([]Validator, 0)

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

	cacheName := fmt.Sprintf("validators-%v", offset)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			validators = GetValidatorList()

			// Only show the ones from the pagination
			validators = validators[offset : offset+limit]

			if err := enc.Encode(validators); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&validators); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return c.JSON(200, validators)

}

func GetSingleValidator(c echo.Context) error {

	hash := c.Param("hash")

	validator := SingleValidator{}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("validator-%v", hash)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			row := db.DB.QueryRow(`SELECT i.address, 
									i.name, 
									s.online, 
									i.version_heartbeat, 
									i.last_heartbeat, 
									i.status,
									i.penalty,
									i.penalties,
									i.owner
								FROM validator_inventory i
								INNER JOIN validator_status s
								ON i.address = s.address
								WHERE i.address = $1`, hash)

			var address, name, online, staked, penalties, owner sql.NullString
			var version_heartbeat, last_heartbeat sql.NullInt64
			var penaltyScore sql.NullFloat64

			err = row.Scan(&address, &name, &online, &version_heartbeat, &last_heartbeat, &staked, &penaltyScore, &penalties, &owner)

			if err != nil {
				log.Println("GetSingleValidator-sql", err)
			}

			rewards := getSingleHotspotRewards(hash, 30)

			rewardsPerDay := sortRewardsPerDay(rewards)

			rewards24h := GetSingleHotspot24Rewards(hash)

			penaltieData := convertPenaltiesToStruct(penalties.String)

			validator = SingleValidator{
				DataType:         "validator",
				Address:          address.String,
				Name:             name.String,
				Owner:            owner.String,
				Online:           online.String,
				VersionHeartbeat: version_heartbeat.Int64,
				LastHeartbeat:    last_heartbeat.Int64,
				Staked:           staked.String,
				Rewards:          rewardsPerDay,
				Rewards24H:       rewards24h,
				PenaltyScore:     penaltyScore.Float64,
				Penalties:        penaltieData,
			}

			if err := enc.Encode(validator); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60 * 10})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&validator); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return c.JSON(200, validator)
}

func GetValidatorList() []Validator {

	rows, err := db.DB.Query(`SELECT i.address, 
								i.name, 
								s.online, 
								i.version_heartbeat, 
								i.last_heartbeat, 
								i.status,
								i.penalty,
								i.owner
							FROM validator_inventory i
							INNER JOIN validator_status s
							ON i.address = s.address`)
	if err != nil {
		log.Println("GetValidatorList-sql", err)
	}

	var version_heartbeat, last_heartbeat sql.NullInt64
	var address, name, online, staked, owner sql.NullString
	var penalty sql.NullFloat64

	defer rows.Close()

	var validators []Validator

	temp := 0
	for rows.Next() {
		temp++

		err = rows.Scan(&address, &name, &online, &version_heartbeat, &last_heartbeat, &staked, &penalty, &owner)
		if err != nil {
			log.Println("GetValidatorList-sql", err)
		}

		validators = append(validators, Validator{
			DataType:         "validator",
			Address:          address.String,
			Name:             name.String,
			Owner:            owner.String,
			Online:           online.String,
			VersionHeartbeat: version_heartbeat.Int64,
			LastHeartbeat:    last_heartbeat.Int64,
			Staked:           staked.String,
			PenaltyScore:     penalty.Float64,
		})
	}

	return validators
}

func getValidatorData(hash string) []Validator {

	validator := make([]Validator, 0)

	row := db.DB.QueryRow(`
	SELECT i.address, i.name, s.online, i.version_heartbeat, i.last_heartbeat, i.status
	FROM validator_inventory i
	INNER JOIN validator_status s
	ON i.address = s.address
	WHERE i.address = $1`, hash)

	var version_heartbeat, last_heartbeat sql.NullInt64
	var address, name, online, staked sql.NullString

	err := row.Scan(&address, &name, &online, &version_heartbeat, &last_heartbeat, &staked)
	if err != nil {
		log.Println("getValidatorData-sql", err)
	}

	validator = append(validator, Validator{
		DataType:         "validator",
		Address:          address.String,
		Name:             name.String,
		Online:           online.String,
		VersionHeartbeat: version_heartbeat.Int64,
		LastHeartbeat:    last_heartbeat.Int64,
		Staked:           staked.String,
	})

	return validator
}

func calculateValidatorAPR(numValidators int) float64 {

	preHalvingTokensPerDay := 300000 / 30
	postHalvingTokensPerDay := preHalvingTokensPerDay / 2

	startDate := time.Date(2021, time.Month(8), 1, 0, 0, 0, 0, time.UTC)
	today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)

	days := today.Sub(startDate).Hours() / 24

	daysTilHalving := mathutil.Clamp(int(days), 0, 365)
	daysAfterHalving := 365 - daysTilHalving

	blendedTokensPerDay := preHalvingTokensPerDay*daysTilHalving + daysAfterHalving*postHalvingTokensPerDay

	annualTokensPerValidator := float64(blendedTokensPerDay) / float64(numValidators)
	stake := 10000

	return (annualTokensPerValidator / float64(stake)) / 2
}

func getValidatorDataByName(query string) []Validator {

	validators := make([]Validator, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("search-validators-%v", query)
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
			i.address, i.name, s.online, i.version_heartbeat, i.last_heartbeat, i.status
			FROM validator_inventory i
			INNER JOIN validator_status s
			ON i.address = s.address
			WHERE i.name LIKE ALL(array[`

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

			var version_heartbeat, last_heartbeat sql.NullInt64
			var address, name, online, staked sql.NullString

			for rows.Next() {

				err = rows.Scan(&address, &name, &online, &version_heartbeat, &last_heartbeat, &staked)
				if err != nil {
					log.Printf("[ERROR]b %v", err)
				}

				if address.Valid && name.Valid {

					validators = append(validators, Validator{
						DataType:         "validator",
						Address:          address.String,
						Name:             name.String,
						Online:           online.String,
						VersionHeartbeat: version_heartbeat.Int64,
						LastHeartbeat:    last_heartbeat.Int64,
						Staked:           staked.String,
					})

				}
				rows.Close()

				if err := enc.Encode(validators); err != nil {
					log.Println("Error gob: ", err)
				}

				db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 300})

			}
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

func getValidatorRewards(hash string) {
}

func convertPenaltiesToStruct(pen string) []Penalty {

	penalties := make([]Penalty, 0)
	json.Unmarshal([]byte(pen), &penalties)
	return penalties
}

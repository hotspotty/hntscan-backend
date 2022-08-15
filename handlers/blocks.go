package handlers

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"hntscan/db"
	"log"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/labstack/echo/v4"
)

func GetBlocks(c echo.Context) error {

	var blocks []Block

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

	cacheName := fmt.Sprintf("blocks-%v", offset)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query("SELECT height, time, block_hash, transaction_count FROM blocks ORDER BY height DESC LIMIT $1 OFFSET $2", limit, offset)
			if err != nil {
				log.Println(err)
			}

			var height, time, transactionCount sql.NullInt64
			var blockHash sql.NullString

			defer rows.Close()

			for rows.Next() {

				err = rows.Scan(&height, &time, &blockHash, &transactionCount)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				blocks = append(blocks, Block{
					height.Int64,
					time.Int64,
					blockHash.String,
					transactionCount.Int64,
				})

			}

			if err := enc.Encode(blocks); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&blocks); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return c.JSON(200, blocks)
}

func GetSingleBlock(c echo.Context) error {

	block := c.Param("block")

	intBlock, err := strconv.Atoi(block)
	if err != nil {
		return c.JSON(400, "Wrong block number")
	}

	blockData := getSingleBlockData(int64(intBlock))

	return c.JSON(200, blockData)
}

func getSingleBlockData(block int64) []BlockData {

	blockData := make([]BlockData, 0)

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	cacheName := fmt.Sprintf("block-%v", block)
	cacheData, err := db.MC.Get(cacheName)
	if err != nil {

		if err == memcache.ErrCacheMiss {

			rows, err := db.DB.Query(`SELECT
										t.block,
										b.block_hash,
										t.hash,
										t.type,
										t.time,
										t.fields
									FROM
										transactions t
										INNER JOIN blocks b
										ON b.height = t.block
									WHERE
										block = $1`, block)
			if err != nil {
				log.Println(err)
			}

			var block, time sql.NullInt64
			var hash, txType, block_hash, fields sql.NullString

			var txCount int64 = 0

			defer rows.Close()

			blockTx := make([]BlockTx, 0)

			for rows.Next() {

				err = rows.Scan(&block, &block_hash, &hash, &txType, &time, &fields)
				if err != nil {
					log.Printf("[ERROR] %v", err)
				}

				// Increment the tx count
				txCount++

				blockTx = append(blockTx, BlockTx{hash.String, txType.String, fields.String})

			}

			blockData = append(blockData, BlockData{"block", block_hash.String, block.Int64, time.Int64, txCount, blockTx})

			if err := enc.Encode(blockData); err != nil {
				log.Println("Error gob: ", err)
			}

			db.MC.Set(&memcache.Item{Key: cacheName, Value: buf.Bytes(), Expiration: 60})
		}

	} else {
		bufDecode := bytes.NewBuffer(cacheData.Value)
		dec := gob.NewDecoder(bufDecode)

		if err := dec.Decode(&blockData); err != nil {
			log.Println("Error decode: ", err)
		}
	}

	return blockData
}

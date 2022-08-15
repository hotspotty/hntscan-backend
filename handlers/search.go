package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func Search(c echo.Context) error {

	query := c.Param("query")

	if query != "" && len(query) > 3 {

		// is this a number?
		if intVal, err := strconv.Atoi(query); err == nil {

			// this is a number, let's look for the block
			blockData := getSingleBlockData(int64(intVal))
			return c.JSON(http.StatusOK, blockData)

		} else {

			// check hotspot name
			hotspotName := getHotspotDataByName(query)
			if len(hotspotName) != 0 {
				return c.JSON(200, hotspotName)
			}

			// check hash for hotspot
			hotspotData := getHotspotData(query)
			if len(hotspotData) != 0 {
				if hotspotData[0].Address != "" {
					return c.JSON(200, hotspotData)
				}
			}

			// check hash for wallet
			walletData := getWalletData(query)
			if len(walletData) != 0 {
				if walletData[0].Balance.HST != -1 {
					return c.JSON(200, walletData)
				}
			}

			// check validator name
			validatorName := getValidatorDataByName(query)
			log.Println(validatorName)
			if len(validatorName) != 0 {
				if validatorName[0].Name != "" {
					return c.JSON(200, validatorName)
				}
			}

			// check hash for validator
			validatorData := getValidatorData(query)
			if len(validatorData) != 0 {
				if validatorData[0].Name != "" {
					return c.JSON(200, validatorData)
				}
			}

			// check hash for transaction
			transactionData := getTransactionData(query)
			if len(transactionData) != 0 {
				if transactionData[0].Hash != "" {
					return c.JSON(200, transactionData)
				}
			}

			return c.String(http.StatusOK, "[]")

		}
	}

	return c.String(http.StatusOK, "[]")
}

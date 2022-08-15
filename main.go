/*
	HNTScan.io
	Powered by Hotspotty.net
*/

package main

import (
	"flag"
	"hntscan/db"
	"hntscan/handlers"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	// Start database connection
	db.Start()

	// Get parameter to know if running on dev or production
	DEV := flag.Bool("dev", false, "Run in development mode")
	flag.Parse()

	serverPort := ":1122"
	if *DEV {
		serverPort = ":8082"
	}

	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} [${method}][${uri}] - ${remote_ip} - ${user_agent}  ${status} - ${latency_human}\n",
		CustomTimeFormat: "2006-01-02 15:04:05",
	}))

	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "HNTScan V1")
	})

	apiGroup := e.Group("/api/v1")

	/*
		SEARCH
		* This you can search for hotspot name, hotspot hash, wallet hash, transaction hash, block number
		* In the future want to also add other features such as hex and address lookups
	*/
	apiGroup.GET("/search/:query/", handlers.Search)

	/* HOMEPAGE STATS */
	apiGroup.GET("/stats/overview/", handlers.GetStatsOverview)

	/* BLOCKS */
	apiGroup.GET("/blocks/", handlers.GetBlocks)
	apiGroup.GET("/blocks/:block/", handlers.GetSingleBlock)

	/* TRANSACTIONS */
	apiGroup.GET("/transactions/", handlers.GetTransactions)
	apiGroup.GET("/transactions/:tx/", handlers.GetSingleTransaction)
	apiGroup.GET("/transactions/:tx/rewards/", handlers.GetRewardTxPagination) // this is to paginate the data for, rewards

	/* HOTSPOTS */
	apiGroup.GET("/hotspots/", handlers.GetHotspots)
	apiGroup.GET("/hotspots/:hash/", handlers.GetSingleHotspot)
	apiGroup.GET("/hotspots/activities/:hash/", handlers.GetSingleHotspotActivities)
	apiGroup.GET("/hotspots/avgbeacons/:hash/", handlers.GetSingleHotspotAvgBeacons)
	apiGroup.GET("/hotspots/status/:hash/", handlers.GetSingleHotspotStatus)
	apiGroup.POST("/hotspots/status/", handlers.GetMultipleHotspotStatus)
	apiGroup.GET("/hotspots/rewards/:hash/:days/", handlers.GetSingleHotspotRewards)

	/* WALLETS */
	apiGroup.GET("/wallets/", handlers.GetWallets)
	apiGroup.GET("/wallets/:hash/", handlers.GetSingleWallets)
	apiGroup.GET("/wallets/:hash/hotspots/", handlers.GetSingleWalletHotspots)
	apiGroup.GET("/wallets/:hash/validators/", handlers.GetSingleWalletValidators)

	/* VALIDATORS */
	apiGroup.GET("/validators/", handlers.GetValidators)
	apiGroup.GET("/validators/:hash/", handlers.GetSingleValidator)

	/* PRICES */
	apiGroup.GET("/price/oracle/", handlers.GetOraclePrices)

	e.Logger.Fatal(e.Start(serverPort))
}

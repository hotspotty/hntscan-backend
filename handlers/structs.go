package handlers

import "time"

type Stats struct {
	Hotspots          Hotspots       `json:"hotspots"`
	HNTPrice          HNTPrice       `json:"hnt_price"`
	Height            BlockHeight    `json:"block"`
	DCSpent           int64          `json:"dc_spent"`
	ValidatorCount    ValidatorStats `json:"validator"`
	Challenges        int            `json:"challenges"`
	OUICount          int            `json:"oui_count"`
	Countries         int            `json:"countries"`
	Cities            int            `json:"cities"`
	CirculatingSupply float64        `json:"circulating_supply"`
	MarketCap         int            `json:"market_cap"`
	MarketCapRank     int            `json:"market_cap_rank"`
	LastHotspot       LastHotspot    `json:"last_hotspot"`
	LastMaker         LastMaker      `json:"last_maker"`
}

type StatsInventory struct {
	Blocks            int `json:"blocks"`
	Challenges        int `json:"challenges"`
	Cities            int `json:"cities"`
	CoingeckoPriceEUR int `json:"coingecko_price_eur"`
	CoingeckoPriceGBP int `json:"coingecko_price_gbp"`
	CoingeckoPriceUSD int `json:"coingecko_price_usd"`
	ConsensusGroups   int `json:"consensus_groups"`
	Countries         int `json:"countries"`
	Hotspots          int `json:"hotspots"`
	HotspotsDataOnly  int `json:"hotspots_dataonly"`
	HotspotsOnline    int `json:"hotspots_online"`
	OUIs              int `json:"ouis"`
	Transactions      int `json:"transactions"`
	Validators        int `json:"validators"`
}

type VarsInventory struct {
	BlockTime               int64 `json:"block_time"`
	DCPayloadSize           int64 `json:"dc_payload_size"`
	MonthlyRewards          int64 `json:"monthly_rewards"`
	ConsensusNumber         int64 `json:"consensus_number"`
	StakeWithdrawalCooldown int64 `json:"stake_withdraw_cooldown"`
	ValidatorMinimumStake   int64 `json:"validator_minimum_stake"`
}

type ValidatorStats struct {
	Count                 int64           `json:"count"`
	ConsensusNumber       int64           `json:"consensus_number"`
	StakeWithdrawCooldown int64           `json:"stake_withdraw_cooldown"`
	ValidatorMinimumStake int64           `json:"validator_minimum_stake"`
	Online                int64           `json:"online"`
	Versions              map[int64]int64 `json:"versions"`
	APR                   float64         `json:"apr"`
}

// var version_heartbeat, last_heartbeat sql.NullInt64
// var address, name, online, version_heartbeat, last_heartbeat sql.NullString

type Validator struct {
	DataType         string  `json:"data_type"`
	Address          string  `json:"address"`
	Owner            string  `json:"owner"`
	Name             string  `json:"name"`
	Online           string  `json:"online"`
	VersionHeartbeat int64   `json:"version_heartbeat"`
	LastHeartbeat    int64   `json:"last_heartbeat"`
	Staked           string  `json:"staked"`
	PenaltyScore     float64 `json:"penalty_score"`
}

type SingleValidator struct {
	DataType         string          `json:"data_type"`
	Address          string          `json:"address"`
	Owner            string          `json:"owner"`
	Name             string          `json:"name"`
	Online           string          `json:"online"`
	VersionHeartbeat int64           `json:"version_heartbeat"`
	LastHeartbeat    int64           `json:"last_heartbeat"`
	Staked           string          `json:"staked"`
	Rewards          map[int64]int64 `json:"rewards"`
	Rewards24H       map[int64]int64 `json:"rewards_24h"`
	PenaltyScore     float64         `json:"penalty_score"`
	Penalties        []Penalty       `json:"penalties"`
}

type Penalty struct {
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
	Height int64   `json:"height"`
}

type OraclePrices struct {
	Min    float64           `json:"min"`
	Max    float64           `json:"max"`
	Prices map[int64]float64 `json:"prices"`
}

type HNTPrice struct {
	Price      float64 `json:"price"`
	Percentage float64 `json:"percentage"`
	Oracle     int     `json:"oracle"`
}

type BlockHeight struct {
	Height int `json:"height"`
	Change int `json:"change"`
}

type HotspotTrend struct {
	LastDays map[string]int `json:"last_days"`
	Start    int            `json:"start"`
	End      int            `json:"end"`
}

type LastHotspot struct {
	Hash         string `json:"hash"`
	Name         string `json:"name"`
	Location     string `json:"location"`
	Country      string `json:"country"`
	ShortCountry string `json:"short_country"`
}

type LastMaker struct {
	Count   int    `json:"count"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Block struct {
	Height           int64  `json:"height"`
	Time             int64  `json:"time"`
	Hash             string `json:"hash"`
	TransactionCount int64  `json:"transaction_count"`
}

type BlockData struct {
	DataType string    `json:"data_type"`
	Hash     string    `json:"hash"`
	Height   int64     `json:"height"`
	Time     int64     `json:"time"`
	TxCount  int64     `json:"transaction_count"`
	BlockTx  []BlockTx `json:"block_transactions"`
}

type BlockTx struct {
	Hash   string `json:"hash"`
	Type   string `json:"type"`
	Fields string `json:"fields"`
}

type Location struct {
	Location     string `json:"location"`
	Country      string `json:"country"`
	ShortCountry string `json:"short_country"`
	City         string `json:"city"`
	Street       string `json:"street"`
}

type SingleHotspot struct {
	DataType          string   `json:"data_type"`
	Address           string   `json:"address"`
	Name              string   `json:"name"`
	Owner             string   `json:"owner"`
	Location          Location `json:"location"`
	LastPocChallenge  int64    `json:"last_poc_challenge"`
	FirstBlock        int64    `json:"first_block"`
	LastBlock         int64    `json:"last_block"`
	FirstTimestamp    int64    `json:"first_timestamp"`
	Nonce             int64    `json:"nonce"`
	RewardScale       float64  `json:"reward_scale"`
	Elevation         int64    `json:"elevation"`
	Gain              int64    `json:"gain"`
	Maker             string   `json:"maker"`
	Payer             string   `json:"payer"`
	WitnessCount      int      `json:"witness_count"`
	Active            bool     `json:"active"`
	ActivityTimestamp int64    `json:"activity_timestamp"`
	ActivityTX        string   `json:"activity_tx"`
}

type Hotspot struct {
	DataType          string   `json:"data_type"`
	Address           string   `json:"address"`
	Name              string   `json:"name"`
	Owner             string   `json:"owner"`
	Location          Location `json:"location"`
	LastPocChallenge  int64    `json:"last_poc_challenge"`
	FirstBlock        int64    `json:"first_block"`
	LastBlock         int64    `json:"last_block"`
	FirstTimestamp    int64    `json:"first_timestamp"`
	Nonce             int64    `json:"nonce"`
	RewardScale       float64  `json:"reward_scale"`
	Elevation         int64    `json:"elevation"`
	Gain              int64    `json:"gain"`
	Maker             string   `json:"maker"`
	Payer             string   `json:"payer"`
	Active            bool     `json:"active"`
	ActivityTimestamp int64    `json:"activity_timestamp"`
	ActivityTX        string   `json:"activity_tx"`
}

type HotspotSearch struct {
	DataType         string   `json:"data_type"`
	Address          string   `json:"address"`
	Name             string   `json:"name"`
	Owner            string   `json:"owner"`
	Location         Location `json:"location"`
	LastPocChallenge int64    `json:"last_poc_challenge"`
	FirstBlock       int64    `json:"first_block"`
	LastBlock        int64    `json:"last_block"`
	FirstTimestamp   int64    `json:"first_timestamp"`
	Nonce            int64    `json:"nonce"`
	RewardScale      float64  `json:"reward_scale"`
	Elevation        int64    `json:"elevation"`
	Gain             int64    `json:"gain"`
}

type Hotspots struct {
	Total  int64        `json:"total"`
	Online int64        `json:"online"`
	Trend  HotspotTrend `json:"trend"`
}

type Reward struct {
	Days       int             `json:"days"`
	Rewards    map[int64]int64 `json:"rewards"`
	Rewards24H map[int64]int64 `json:"rewards_24h"`
}

type SevenDayAvgBeacons struct {
	Beacons int `json:"7d_average_beacons"`
}

type Maker struct {
	Name  string `json:"name"`
	Payer string `json:"payer"`
}

type WalletBalance struct {
	HST    int64 `json:"hst"`
	DC     int64 `json:"dc"`
	HNT    int64 `json:"hnt"`
	STAKE  int64 `json:"stake"`
	MOBILE int64 `json:"mobile"`
	IOT    int64 `json:"iot"`
}

type Wallet struct {
	DataType       string          `json:"data_type"`
	Address        string          `json:"address"`
	HotspotCount   int             `json:"hotspot_count"`
	Balance        WalletBalance   `json:"balance"`
	Rewards        map[int64]int64 `json:"rewards"`
	Rewards24H     map[int64]int64 `json:"rewards_24h"`
	ValidatorCount int             `json:"validator_count"`
	LastBlock      int64           `json:"last_block"`
}

type WalletList struct {
	DataType       string        `json:"data_type"`
	Address        string        `json:"address"`
	HotspotCount   int           `json:"hotspot_count"`
	ValidatorCount int           `json:"validator_count"`
	Balance        WalletBalance `json:"balance"`
	LastBlock      int64         `json:"last_block"`
}

type Coingecko struct {
	BlockTimeInMinutes  int     `json:"block_time_in_minutes"`
	MarketCapRank       int     `json:"market_cap_rank"`
	CoingeckoRank       int     `json:"coingecko_rank"`
	PublicInterestScore float64 `json:"public_interest_score"`
	MarketData          struct {
		CurrentPrice struct {
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Usd float64 `json:"usd"`
		} `json:"current_price"`
		Ath struct {
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Usd float64 `json:"usd"`
		} `json:"ath"`
		AthChangePercentage struct {
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Usd float64 `json:"usd"`
		} `json:"ath_change_percentage"`
		AthDate struct {
			Eur time.Time `json:"eur"`
			Gbp time.Time `json:"gbp"`
			Usd time.Time `json:"usd"`
		} `json:"ath_date"`
		MarketCap struct {
			Eur int `json:"eur"`
			Gbp int `json:"gbp"`
			Usd int `json:"usd"`
		} `json:"market_cap"`
		TotalVolume struct {
			Eur int `json:"eur"`
			Gbp int `json:"gbp"`
			Usd int `json:"usd"`
		} `json:"total_volume"`
		PriceChange24H               float64 `json:"price_change_24h"`
		PriceChangePercentage24H     float64 `json:"price_change_percentage_24h"`
		PriceChangePercentage7D      float64 `json:"price_change_percentage_7d"`
		PriceChangePercentage30D     float64 `json:"price_change_percentage_30d"`
		MarketCapChange24H           float64 `json:"market_cap_change_24h"`
		MarketCapChangePercentage24H float64 `json:"market_cap_change_percentage_24h"`
		PriceChange24HInCurrency     struct {
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Usd float64 `json:"usd"`
		} `json:"price_change_24h_in_currency"`
		PriceChangePercentage24HInCurrency struct {
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Usd float64 `json:"usd"`
		} `json:"price_change_percentage_24h_in_currency"`
		PriceChangePercentage30DInCurrency struct {
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Usd float64 `json:"usd"`
		} `json:"price_change_percentage_30d_in_currency"`
		MarketCapChange24HInCurrency struct {
			Usd float64 `json:"usd"`
		} `json:"market_cap_change_24h_in_currency"`
		MarketCapChangePercentage24HInCurrency struct {
			Eur float64 `json:"eur"`
			Gbp float64 `json:"gbp"`
			Usd float64 `json:"usd"`
		} `json:"market_cap_change_percentage_24h_in_currency"`
		TotalSupply       float64   `json:"total_supply"`
		MaxSupply         float64   `json:"max_supply"`
		CirculatingSupply float64   `json:"circulating_supply"`
		LastUpdated       time.Time `json:"last_updated"`
	} `json:"market_data"`
	Tickers []struct {
		Base   string `json:"base"`
		Target string `json:"target"`
		Market struct {
			Name                string `json:"name"`
			Identifier          string `json:"identifier"`
			HasTradingIncentive bool   `json:"has_trading_incentive"`
		} `json:"market"`
		Last                   float64     `json:"last"`
		Volume                 float64     `json:"volume"`
		TrustScore             string      `json:"trust_score"`
		BidAskSpreadPercentage float64     `json:"bid_ask_spread_percentage"`
		Timestamp              time.Time   `json:"timestamp"`
		LastTradedAt           time.Time   `json:"last_traded_at"`
		LastFetchAt            time.Time   `json:"last_fetch_at"`
		IsAnomaly              bool        `json:"is_anomaly"`
		IsStale                bool        `json:"is_stale"`
		TradeURL               string      `json:"trade_url"`
		TokenInfoURL           interface{} `json:"token_info_url"`
		CoinID                 string      `json:"coin_id"`
		TargetCoinID           string      `json:"target_coin_id,omitempty"`
	} `json:"tickers"`
}

type Transaction struct {
	DataType string `json:"data_type"`
	Height   int64  `json:"height"`
	Hash     string `json:"hash"`
	Type     string `json:"type"`
	Time     int64  `json:"time"`
	Fields   string `json:"fields"`
}

type SingleReward struct {
	Type    string      `json:"type"`
	Amount  int         `json:"amount"`
	Account string      `json:"account"`
	Gateway interface{} `json:"gateway"`
}

type RewardV2 struct {
	Hash       string         `json:"hash"`
	Type       string         `json:"type"`
	Rewards    []SingleReward `json:"rewards"`
	EndEpoch   int64          `json:"end_epoch"`
	StartEpoch int64          `json:"start_epoch"`
}

type TxRewardV2 struct {
	DataType string `json:"data_type"`
	Height   int64  `json:"height"`
	Hash     string `json:"hash"`
	Type     string `json:"type"`
	Time     int64  `json:"time"`
	Rewards  []struct {
		Type    string      `json:"type"`
		Amount  int         `json:"amount"`
		Account string      `json:"account"`
		Gateway interface{} `json:"gateway"`
	} `json:"rewards"`
}

type WitnessStruct struct {
	Fee  int    `json:"fee"`
	Hash string `json:"hash"`
	Path []struct {
		Receipt struct {
			Snr       float64     `json:"snr"`
			Data      string      `json:"data"`
			Origin    string      `json:"origin"`
			Signal    int         `json:"signal"`
			Channel   int         `json:"channel"`
			Gateway   string      `json:"gateway"`
			Datarate  interface{} `json:"datarate"`
			TxPower   int         `json:"tx_power"`
			Frequency int         `json:"frequency"`
			Timestamp int64       `json:"timestamp"`
		} `json:"receipt"`
		Witnesses []struct {
			Snr           float64 `json:"snr"`
			Owner         string  `json:"owner"`
			Signal        int     `json:"signal"`
			Channel       int     `json:"channel"`
			Gateway       string  `json:"gateway"`
			Datarate      string  `json:"datarate"`
			IsValid       bool    `json:"is_valid"`
			Location      string  `json:"location"`
			Frequency     float64 `json:"frequency"`
			Timestamp     int64   `json:"timestamp"`
			PacketHash    string  `json:"packet_hash"`
			InvalidReason string  `json:"invalid_reason,omitempty"`
		} `json:"witnesses"`
		Challengee         string `json:"challengee"`
		ChallengeeOwner    string `json:"challengee_owner"`
		ChallengeeLocation string `json:"challengee_location"`
	} `json:"path"`
	Type               string `json:"type"`
	Secret             string `json:"secret"`
	Challenger         string `json:"challenger"`
	OnionKeyHash       string `json:"onion_key_hash"`
	ChallengerOwner    string `json:"challenger_owner"`
	RequestBlockHash   string `json:"request_block_hash"`
	ChallengerLocation string `json:"challenger_location"`
}

type WitnessData struct {
	Hash          string  `json:"hash"`
	Distance      int     `json:"distance"`
	Datarate      string  `json:"datarate"`
	RSSI          int     `json:"rssi"`
	SNR           float64 `json:"snr"`
	Frequency     float64 `json:"frequency"`
	Valid         bool    `json:"valid"`
	Timestamp     int64   `json:"timestamp"`
	Channel       int     `json:"channel"`
	Location      string  `json:"location"`
	InvalidReason string  `json:"invalid_reason,omitempty"`
}

type WitnessParsed struct {
	Hash               string        `json:"hash"`
	Time               int64         `json:"time"`
	Block              int64         `json:"block"`
	Distance           int           `json:"distance"`
	Challenger         string        `json:"challenger"`
	ChallengerLocation string        `json:"challenger_location"`
	Beaconer           string        `json:"beaconer"`
	BeaconerLocation   string        `json:"beaconer_location"`
	Valid              bool          `json:"valid"`
	ValidCount         int           `json:"valid_count"`
	InvalidCount       int           `json:"invalid_count"`
	WintessList        []WitnessData `json:"witnesses"`
}

type ChallengeeStruct struct {
	Fee  int    `json:"fee"`
	Hash string `json:"hash"`
	Path []struct {
		Receipt struct {
			Snr       float64     `json:"snr"`
			Data      string      `json:"data"`
			Origin    string      `json:"origin"`
			Signal    int         `json:"signal"`
			Channel   int         `json:"channel"`
			Gateway   string      `json:"gateway"`
			Datarate  interface{} `json:"datarate"`
			TxPower   int         `json:"tx_power"`
			Frequency int         `json:"frequency"`
			Timestamp int64       `json:"timestamp"`
		} `json:"receipt"`
		Witnesses []struct {
			Snr           float64 `json:"snr"`
			Owner         string  `json:"owner"`
			Signal        int     `json:"signal"`
			Channel       int     `json:"channel"`
			Gateway       string  `json:"gateway"`
			Datarate      string  `json:"datarate"`
			IsValid       bool    `json:"is_valid"`
			Location      string  `json:"location"`
			Frequency     float64 `json:"frequency"`
			Timestamp     int64   `json:"timestamp"`
			PacketHash    string  `json:"packet_hash"`
			InvalidReason string  `json:"invalid_reason,omitempty"`
		} `json:"witnesses"`
		Challengee         string `json:"challengee"`
		ChallengeeOwner    string `json:"challengee_owner"`
		ChallengeeLocation string `json:"challengee_location"`
	} `json:"path"`
	Type               string `json:"type"`
	Secret             string `json:"secret"`
	Challenger         string `json:"challenger"`
	OnionKeyHash       string `json:"onion_key_hash"`
	ChallengerOwner    string `json:"challenger_owner"`
	RequestBlockHash   string `json:"request_block_hash"`
	ChallengerLocation string `json:"challenger_location"`
}

type ChallengerParsed struct {
	Hash               string        `json:"hash"`
	Time               int64         `json:"time"`
	Block              int64         `json:"block"`
	Challenger         string        `json:"challenger"`
	ChallengerLocation string        `json:"challenger_location"`
	Beaconer           string        `json:"beaconer"`
	BeaconerLocation   string        `json:"beaconer_location"`
	ValidCount         int           `json:"valid_count"`
	InvalidCount       int           `json:"invalid_count"`
	WintessList        []WitnessData `json:"witnesses"`
}

type GatewayParsed struct {
	Fee        int    `json:"fee"`
	Gain       int    `json:"gain"`
	Hash       string `json:"hash"`
	Type       string `json:"type"`
	Nonce      int    `json:"nonce"`
	Owner      string `json:"owner"`
	Payer      string `json:"payer"`
	Gateway    string `json:"gateway"`
	Location   string `json:"location"`
	Elevation  int    `json:"elevation"`
	StakingFee int    `json:"staking_fee"`
	Timestamp  int64  `json:"timestamp"`
}

type RewardStruct struct {
	Hash    string `json:"hash"`
	Type    string `json:"type"`
	Rewards []struct {
		Type    string      `json:"type"`
		Amount  int         `json:"amount"`
		Account string      `json:"account"`
		Gateway interface{} `json:"gateway"`
	} `json:"rewards"`
	EndEpoch   int `json:"end_epoch"`
	StartEpoch int `json:"start_epoch"`
}

type RewardParsed struct {
	Hash   string `json:"hash"`
	Time   int64  `json:"time"`
	Block  int64  `json:"block"`
	Amount int    `json:"amount"`
}

type ChallengerStruct struct {
	Fee  int    `json:"fee"`
	Hash string `json:"hash"`
	Path []struct {
		Receipt struct {
			Snr       float64     `json:"snr"`
			Data      string      `json:"data"`
			Origin    string      `json:"origin"`
			Signal    int         `json:"signal"`
			Channel   int         `json:"channel"`
			Gateway   string      `json:"gateway"`
			Datarate  interface{} `json:"datarate"`
			TxPower   int         `json:"tx_power"`
			Frequency float64     `json:"frequency"`
			Timestamp int64       `json:"timestamp"`
		} `json:"receipt"`
		Witnesses []struct {
			Snr           float64 `json:"snr"`
			Owner         string  `json:"owner"`
			Signal        int     `json:"signal"`
			Channel       int     `json:"channel"`
			Gateway       string  `json:"gateway"`
			Datarate      string  `json:"datarate"`
			IsValid       bool    `json:"is_valid"`
			Location      string  `json:"location"`
			Frequency     float64 `json:"frequency"`
			Timestamp     int64   `json:"timestamp"`
			PacketHash    string  `json:"packet_hash"`
			InvalidReason string  `json:"invalid_reason,omitempty"`
		} `json:"witnesses"`
		Challengee         string `json:"challengee"`
		ChallengeeOwner    string `json:"challengee_owner"`
		ChallengeeLocation string `json:"challengee_location"`
	} `json:"path"`
	Type               string `json:"type"`
	Secret             string `json:"secret"`
	Challenger         string `json:"challenger"`
	OnionKeyHash       string `json:"onion_key_hash"`
	ChallengerOwner    string `json:"challenger_owner"`
	RequestBlockHash   string `json:"request_block_hash"`
	ChallengerLocation string `json:"challenger_location"`
}

type ChallengeeParsed struct {
	Hash               string        `json:"hash"`
	Time               int64         `json:"time"`
	Block              int64         `json:"block"`
	Challenger         string        `json:"challenger"`
	ChallengerLocation string        `json:"challenger_location"`
	Beaconer           string        `json:"beaconer"`
	BeaconerLocation   string        `json:"beaconer_location"`
	ValidCount         int           `json:"valid_count"`
	InvalidCount       int           `json:"invalid_count"`
	WintessList        []WitnessData `json:"witnesses"`
}

type DataPacket struct {
	Hash         string `json:"hash"`
	Type         string `json:"type"`
	Closer       string `json:"closer"`
	StateChannel struct {
		ID        string `json:"id"`
		Nonce     int    `json:"nonce"`
		Owner     string `json:"owner"`
		State     string `json:"state"`
		RootHash  string `json:"root_hash"`
		Summaries []struct {
			Owner      string `json:"owner,omitempty"`
			Client     string `json:"client"`
			NumDcs     int    `json:"num_dcs"`
			Location   string `json:"location,omitempty"`
			NumPackets int    `json:"num_packets"`
		} `json:"summaries"`
		ExpireAtBlock int `json:"expire_at_block"`
	} `json:"state_channel"`
	ConflictsWith interface{} `json:"conflicts_with"`
}

type DataPacketParsed struct {
	Hash       string `json:"hash"`
	Timestamp  int64  `json:"time"`
	Block      int64  `json:"block"`
	NumDcs     int    `json:"num_dcs"`
	Location   string `json:"location"`
	NumPackets int    `json:"num_packets"`
}

type ActivityResponsePayload struct {
	Limit    int                         `json:"limit"`
	Page     int                         `json:"page"`
	Activity ActivityResponsePayloadData `json:"activity"`
}

type ActivityResponsePayloadData struct {
	Witnesses   []WitnessParsed    `json:"witnesses"`
	Challengers []ChallengerParsed `json:"challengers"`
	Challengees []ChallengeeParsed `json:"challengees"`
	Rewards     []RewardParsed     `json:"rewards"`
	DataPackets []DataPacketParsed `json:"data_packets"`
	GatewayData []GatewayParsed    `json:"gateway_data"`
}

type HotspotStruct struct {
	DataType          string  `json:"data_type"`
	Address           string  `json:"address"`
	Block             int     `json:"block_height"`
	BlockAdded        int     `json:"block_added"`
	Place             string  `json:"place"`
	LastPocChallenge  int     `json:"last_poc_challenge"`
	Location          string  `json:"location"`
	Name              string  `json:"name"`
	Nonce             int     `json:"nonce"`
	Owner             string  `json:"owner"`
	Reward_scale      float64 `json:"reward_scale"`
	Timestamp_added   int     `json:"timestamp_added"`
	Elevation         int     `json:"elevation"`
	Gain              int     `json:"gain"`
	LastUpdate        int     `json:"last_update"`
	Entity            string  `json:"entity"`
	LastAssetion      int     `json:"last_assertion"`
	Payer             string  `json:"payer"`
	Mode              string  `json:"mode"`
	Maker             string  `json:"maker"`
	Active            bool    `json:"active"`
	ActivityTimestamp int64   `json:"activity_timestamp"`
	ActivityTX        string  `json:"activity_tx"`
}

type Active struct {
	Active    bool   `json:"active"`
	Timestamp int64  `json:"timestamp"`
	TX        string `json:"tx"`
}

type GeoCode struct {
	ShortStreet  string `json:"short_street"`
	ShortState   string `json:"short_state"`
	ShortCountry string `json:"short_country"`
	ShortCity    string `json:"short_city"`
	LongStreet   string `json:"long_street"`
	LongState    string `json:"long_state"`
	LongCountry  string `json:"long_country"`
	LongCity     string `json:"long_city"`
	CityID       string `json:"city_id"`
}

type InvalidReasons struct {
	Reason string `json:"type"`
	Count  int    `json:"count"`
}

type HotspotsValid struct {
	Address string `json:"id"`
	Count   int    `json:"count"`
}

type HotspotsInvalid struct {
	Address string           `json:"id"`
	Count   int              `json:"count"`
	Reasons []InvalidReasons `json:"reasons"`
}

type ValidStruct struct {
	Address string `json:"id"`
}

type InvalidStruct struct {
	Address string `json:"id"`
	Reason  string `json:"reason"`
}

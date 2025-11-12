package polymarketrealtime

type Topic string

const (
	// TopicActivity is the topic for activity
	TopicActivity Topic = "activity"

	// TopicComments is the topic for comments
	TopicComments Topic = "comments"

	// TopicRfq is the topic for RFQ
	TopicRfq Topic = "rfq"

	// TopicCryptoPrices is the topic for crypto prices
	TopicCryptoPrices Topic = "crypto_prices"

	// TopicCryptoPricesChainlink is the topic for crypto prices from Chainlink
	TopicCryptoPricesChainlink Topic = "crypto_prices_chainlink"

	// TopicEquityPrices is the topic for equity prices
	TopicEquityPrices Topic = "equity_prices"

	// TopicClobUser is the topic for CLOB user data
	TopicClobUser Topic = "clob_user"

	// TopicClobMarket is the topic for CLOB market data
	TopicClobMarket Topic = "clob_market"

	// Aliases for backward compatibility and clarity
	TopicCLOBUser   = TopicClobUser
	TopicCLOBMarket = TopicClobMarket
)

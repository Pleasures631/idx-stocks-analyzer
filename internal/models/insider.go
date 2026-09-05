package models

// InsiderBrokerMapping adalah pasangan emiten dan broker dari aktivitas insider KSEI.
type InsiderBrokerMapping struct {
	StockCode  string `json:"stock_code"`
	BrokerCode string `json:"broker_code"`
}

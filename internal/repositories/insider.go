package repositories

import (
	"sort"
	"strings"

	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"

	"github.com/jmoiron/sqlx"
)

type stockMarketMakerRow struct {
	StockCode   string `db:"stock_code"`
	MarketMaker string `db:"market_maker"`
}

// UpsertInsiderMarketMakers merges KSEI broker codes into both the requested
// stock column and the existing normalized market-maker table.
func UpsertInsiderMarketMakers(mappings []models.InsiderBrokerMapping) (int, error) {
	unique := make(map[string]models.InsiderBrokerMapping, len(mappings))
	for _, mapping := range mappings {
		stock := strings.ToUpper(strings.TrimSpace(mapping.StockCode))
		broker := strings.ToUpper(strings.TrimSpace(mapping.BrokerCode))
		if stock == "" || broker == "" {
			continue
		}
		unique[stock+"|"+broker] = models.InsiderBrokerMapping{StockCode: stock, BrokerCode: broker}
	}
	if len(unique) == 0 {
		return 0, nil
	}

	clean := make([]models.InsiderBrokerMapping, 0, len(unique))
	for _, mapping := range unique {
		clean = append(clean, mapping)
	}
	sort.Slice(clean, func(i, j int) bool {
		if clean[i].StockCode == clean[j].StockCode {
			return clean[i].BrokerCode < clean[j].BrokerCode
		}
		return clean[i].StockCode < clean[j].StockCode
	})

	stockCodes := make([]string, 0, len(clean))
	seenStocks := make(map[string]bool)
	for _, mapping := range clean {
		if !seenStocks[mapping.StockCode] {
			seenStocks[mapping.StockCode] = true
			stockCodes = append(stockCodes, mapping.StockCode)
		}
	}
	query, args, err := sqlx.In(`SELECT stock_code, COALESCE(market_maker, '') AS market_maker FROM m_list_stocks WHERE stock_code IN (?)`, stockCodes)
	if err != nil {
		return 0, err
	}
	query = database.DB.Rebind(query)
	var rows []stockMarketMakerRow
	if err := database.DB.Select(&rows, query, args...); err != nil {
		return 0, err
	}

	current := make(map[string]map[string]bool, len(rows))
	for _, row := range rows {
		current[row.StockCode] = parseBrokerCodes(row.MarketMaker)
	}
	tx, err := database.DB.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	newCodes := 0
	for _, mapping := range clean {
		brokers, exists := current[mapping.StockCode]
		if !exists {
			continue
		}
		if brokers[mapping.BrokerCode] {
			continue
		}
		brokers[mapping.BrokerCode] = true
		codes := make([]string, 0, len(brokers))
		for code := range brokers {
			codes = append(codes, code)
		}
		sort.Strings(codes)
		if _, err := tx.Exec(`UPDATE m_list_stocks SET market_maker = ? WHERE stock_code = ?`, strings.Join(codes, ","), mapping.StockCode); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(`INSERT IGNORE INTO m_market_maker (stock_code, broker_code) VALUES (?, ?)`, mapping.StockCode, mapping.BrokerCode); err != nil {
			return 0, err
		}
		newCodes++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return newCodes, nil
}

func parseBrokerCodes(value string) map[string]bool {
	result := make(map[string]bool)
	for _, code := range strings.Split(value, ",") {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code != "" {
			result[code] = true
		}
	}
	return result
}

package repositories

import (
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"
)

func UpsertStockbitForeignDomestic(row models.StockbitForeignDomesticDB) error {
	const query = `
	INSERT INTO t_stockbit_foreign_domestic (
		symbol, trade_date, market_type,
		foreign_buy_value, foreign_sell_value, domestic_buy_value, domestic_sell_value,
		foreign_net_value, domestic_net_value,
		foreign_buy_volume, foreign_sell_volume, domestic_buy_volume, domestic_sell_volume,
		foreign_buy_frequency, foreign_sell_frequency, domestic_buy_frequency, domestic_sell_frequency,
		last_updated, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	ON DUPLICATE KEY UPDATE
		foreign_buy_value = VALUES(foreign_buy_value), foreign_sell_value = VALUES(foreign_sell_value),
		domestic_buy_value = VALUES(domestic_buy_value), domestic_sell_value = VALUES(domestic_sell_value),
		foreign_net_value = VALUES(foreign_net_value), domestic_net_value = VALUES(domestic_net_value),
		foreign_buy_volume = VALUES(foreign_buy_volume), foreign_sell_volume = VALUES(foreign_sell_volume),
		domestic_buy_volume = VALUES(domestic_buy_volume), domestic_sell_volume = VALUES(domestic_sell_volume),
		foreign_buy_frequency = VALUES(foreign_buy_frequency), foreign_sell_frequency = VALUES(foreign_sell_frequency),
		domestic_buy_frequency = VALUES(domestic_buy_frequency), domestic_sell_frequency = VALUES(domestic_sell_frequency),
		last_updated = VALUES(last_updated), updated_at = NOW()`
	_, err := database.DB.Exec(query, row.Symbol, row.TradeDate, row.MarketType,
		row.ForeignBuyValue, row.ForeignSellValue, row.DomesticBuyValue, row.DomesticSellValue,
		row.ForeignNetValue, row.DomesticNetValue, row.ForeignBuyVolume, row.ForeignSellVolume,
		row.DomesticBuyVolume, row.DomesticSellVolume, row.ForeignBuyFreq, row.ForeignSellFreq,
		row.DomesticBuyFreq, row.DomesticSellFreq, row.LastUpdated)
	return err
}

func GetStockbitForeignDomestic(symbol, from, to string) ([]models.StockbitForeignDomesticDB, error) {
	query := `
	SELECT id, symbol, trade_date, market_type, foreign_buy_value, foreign_sell_value,
		domestic_buy_value, domestic_sell_value, foreign_net_value, domestic_net_value,
		foreign_buy_volume, foreign_sell_volume, domestic_buy_volume, domestic_sell_volume,
		foreign_buy_frequency, foreign_sell_frequency, domestic_buy_frequency, domestic_sell_frequency,
		last_updated, created_at, updated_at
	FROM t_stockbit_foreign_domestic
	WHERE symbol = ?`
	args := []any{symbol}
	if from != "" {
		query += " AND trade_date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND trade_date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY trade_date"
	rows := []models.StockbitForeignDomesticDB{}
	err := database.DB.Select(&rows, query, args...)
	return rows, err
}

package repositories

import (
	"indonesia-stocks-api/internal/database"
	"indonesia-stocks-api/internal/models"
)

func UpsertStockbitIHSGChart(points []models.StockbitIHSGChartPoint) error {
	if len(points) == 0 {
		return nil
	}
	const query = "INSERT INTO t_stockbit_ihsg_chart (symbol, trade_date, `interval`, observed_at, xlabel, value, percentage, change_value, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE xlabel = VALUES(xlabel), value = VALUES(value), percentage = VALUES(percentage), change_value = VALUES(change_value), updated_at = NOW()"
	for _, point := range points {
		if _, err := database.DB.Exec(query, point.Symbol, point.TradeDate, point.Interval, point.ObservedAt, point.XLabel, point.Value, point.Percentage, point.Change); err != nil {
			return err
		}
	}
	return nil
}

func GetStockbitIHSGChart(symbol, from, to string) ([]models.StockbitIHSGChartPoint, error) {
	query := "SELECT id, symbol, trade_date, `interval`, observed_at, xlabel, value, percentage, change_value, created_at, updated_at FROM t_stockbit_ihsg_chart WHERE symbol = ?"
	args := []any{symbol}
	if from != "" {
		query += " AND trade_date >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND trade_date <= ?"
		args = append(args, to)
	}
	query += " ORDER BY observed_at"
	rows := []models.StockbitIHSGChartPoint{}
	err := database.DB.Select(&rows, query, args...)
	return rows, err
}

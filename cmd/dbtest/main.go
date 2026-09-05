package main

import (
	"fmt"
	"os"
	"sort"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type brokerRow struct {
	BrokerCode string  `db:"broker_code"`
	BrokerType string  `db:"broker_type"`
	NetValue   float64 `db:"net_value"`
}

type feat struct {
	status      string
	foreignNet  float64
	top3Sum     float64
	top1Asing   bool
	all3Pos     bool
	dayPct      float64 // (close-open)/open*100
	upperHalf   bool    // close > (high+low)/2
	closeOpen   float64
}

func main() {
	godotenv.Load()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FJakarta",
		os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	details := []struct {
		StockCode  string `db:"stock_code"`
		SignalDate string `db:"signal_date"`
		Status     string `db:"status"`
	}{}
	err = db.Select(&details, "SELECT stock_code, DATE_FORMAT(signal_date,'%Y-%m-%d') AS signal_date, status FROM t_backtest_detail WHERE backtest_run_id = 15")
	if err != nil {
		panic(err)
	}

	seen := map[string]bool{}
	feats := []feat{}
	for _, d := range details {
		key := d.StockCode + "|" + d.SignalDate
		if seen[key] {
			continue
		}
		seen[key] = true

		var brokers []brokerRow
		err = db.Select(&brokers, `
			SELECT broker_code, MAX(broker_type) AS broker_type,
				SUM(IF(side='BUY',value,0)) - SUM(IF(side='SELL',value,0)) AS net_value
			FROM t_exodus_broker_summary
			WHERE stock_code = ? AND trade_date = ?
			GROUP BY broker_code
			ORDER BY net_value DESC`, d.StockCode, d.SignalDate)
		if err != nil {
			continue
		}

		sort.Slice(brokers, func(i, j int) bool { return brokers[i].NetValue > brokers[j].NetValue })

		f := feat{status: d.Status}
		for _, b := range brokers {
			if b.BrokerType == "Asing" {
				f.foreignNet += b.NetValue
			}
		}
		if len(brokers) >= 3 {
			for i := 0; i < 3; i++ {
				f.top3Sum += brokers[i].NetValue
				if brokers[i].NetValue <= 0 {
					f.all3Pos = false
				}
			}
			f.top1Asing = brokers[0].BrokerType == "Asing" && brokers[0].NetValue > 0
		} else {
			for _, b := range brokers {
				f.top3Sum += b.NetValue
			}
		}
		if len(brokers) < 3 {
			f.all3Pos = false
		}

		var ts struct {
			Open  float64 `db:"open_price"`
			High  float64 `db:"high_price"`
			Low   float64 `db:"low_price"`
			Close float64 `db:"close_price"`
		}
		err = db.Get(&ts, "SELECT open_price, high_price, low_price, close_price FROM t_trading_summary WHERE stock_code = ? AND trade_date = ?", d.StockCode, d.SignalDate)
		if err == nil && ts.Open > 0 {
			f.dayPct = (ts.Close - ts.Open) / ts.Open * 100
			f.closeOpen = ts.Close / ts.Open
			f.upperHalf = ts.Close > (ts.High+ts.Low)/2
		}

		feats = append(feats, f)
	}

	sum := func(ff []feat, pick func(feat) float64) (n int, sum float64) {
		for _, f := range ff {
			sum += pick(f)
			n++
		}
		return
	}

	group := func(status string) {
		var ff []feat
		for _, f := range feats {
			if f.status == status {
				ff = append(ff, f)
			}
		}
		if len(ff) == 0 {
			return
		}
		_, fn := sum(ff, func(f feat) float64 { return f.foreignNet })
		_, t3 := sum(ff, func(f feat) float64 { return f.top3Sum })
		_, dp := sum(ff, func(f feat) float64 { return f.dayPct })
		_, co := sum(ff, func(f feat) float64 { return f.closeOpen })
		top1AsingCnt, all3PosCnt, upperHalfCnt := 0, 0, 0
		for _, f := range ff {
			if f.top1Asing {
				top1AsingCnt++
			}
			if f.all3Pos {
				all3PosCnt++
			}
			if f.upperHalf {
				upperHalfCnt++
			}
		}
		n := float64(len(ff))
		fmt.Printf("%-8s n=%2d | foreignNet=%.0f top3=%.0f day%%=%.2f close/open=%.3f | top1Asing=%d/%.0f all3Pos=%d/%.0f upperHalf=%d/%.0f\n",
			status, len(ff), fn/n, t3/n, dp/n, co/n,
			top1AsingCnt, n, all3PosCnt, n, upperHalfCnt, n)
	}

	group("WIN")
	group("LOSS")
	group("EXPIRED")
}

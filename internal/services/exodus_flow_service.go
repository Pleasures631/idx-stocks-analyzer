package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"indonesia-stocks-api/internal/helpers"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
)

func AnalyzeExodusBrokerFlowService(symbol, fromStr, toStr string) (*models.ExodusFlowPhase, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	now := time.Now()
	if fromStr == "" {
		fromStr = now.AddDate(0, 0, -6).Format("2006-01-02")
	}
	if toStr == "" {
		toStr = now.Format("2006-01-02")
	}

	startDate, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return nil, fmt.Errorf("invalid from date, format: YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return nil, fmt.Errorf("invalid to date, format: YYYY-MM-DD")
	}

	if startDate.After(endDate) {
		return nil, fmt.Errorf("from cannot be after to")
	}

	flows, err := repositories.GetExodusBrokerFlow(symbol, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	// Ambil breakdown per broker group (RETAIL, FOREIGN, INSTITUTIONAL, LOCAL_MID)
	groupNet, err := repositories.GetExodusBrokerFlowGrouped(symbol, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	result := &models.ExodusFlowPhase{
		Symbol:    symbol,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		TotalDays: int(endDate.Sub(startDate).Hours()/24) + 1,

		// Populate broker group breakdown
		RetailNet:        groupNet["RETAIL"],
		InstitutionalNet: groupNet["INSTITUTIONAL"],
		LocalMidNet:      groupNet["LOCAL_MID"],
	}

	for i := range flows {
		f := &flows[i]
		result.TotalBuyValue += f.BuyValue
		result.TotalSellValue += f.SellValue
		result.NetValue += f.NetValue

		switch f.BrokerType {
		case "Asing":
			result.ForeignNetValue += f.NetValue
		case "Pemerintah":
			result.GovernmentNet += f.NetValue
		case "Lokal":
			result.LocalNetValue += f.NetValue
		}

		f.FormattedNetValue = helpers.FormatBigNumber(f.NetValue)

		ratio := 0.0
		if f.BuyValue+f.SellValue > 0 {
			ratio = (f.NetValue / (f.BuyValue + f.SellValue)) * 100
		}

		f.DisplayStatus = fmt.Sprintf("%s | Net: %s | %s",
			f.BrokerCode, f.FormattedNetValue, flowSideLabel(ratio))
	}

	sorted := flows
	result.TotalBrokers = len(flows)

	acc := []models.ExodusBrokerFlow{}
	dist := []models.ExodusBrokerFlow{}

	for _, f := range sorted {
		if f.NetValue >= 0 {
			acc = append(acc, f)
		} else {
			dist = append(dist, f)
		}
	}

	acc = sortDesc(acc, false)
	dist = sortDesc(dist, true)

	if len(acc) > 5 {
		acc = acc[:5]
	}
	if len(dist) > 5 {
		dist = dist[:5]
	}

	result.BrokersAccumulation = acc
	result.BrokersDistribution = dist

	result.FormattedBuyValue = helpers.FormatBigNumber(result.TotalBuyValue)
	result.FormattedSellValue = helpers.FormatBigNumber(result.TotalSellValue)
	result.FormattedNetValue = helpers.FormatBigNumber(result.NetValue)
	result.FormattedForeignNet = helpers.FormatBigNumber(result.ForeignNetValue)

	// Compute metrics
	totalAbsValue := result.TotalBuyValue + result.TotalSellValue
	if totalAbsValue > 0 {
		smartMoneyNet := result.ForeignNetValue + result.InstitutionalNet
		result.SmartMoneyRatio = (smartMoneyNet / totalAbsValue) * 100
		result.RetailDominance = (result.RetailNet / totalAbsValue) * 100
	}

	// --- New: Smart money consistency & momentum ---
	enrichGroupProfile(result)
	// --- New: Price & volume confirmation ---
	enrichPriceVolume(result)
	// --- New: Anomaly detection (per-broker, Modified Z-Score) ---
	enrichAnomalies(result)

	// Top 1 concentration (top1_net / top3_net)
	if len(acc) > 0 {
		top1Net := acc[0].NetValue
		top3Net := 0.0
		for i := 0; i < len(acc) && i < 3; i++ {
			top3Net += acc[i].NetValue
		}
		if top3Net > 0 {
			result.Top1Concentration = top1Net / top3Net
		}
		// Foreign leadership if top1 broker is FOREIGN group (needs broker_group in query — simplify: check broker_type for now)
		result.ForeignLeadership = acc[0].BrokerType == "Asing"
	}

	result.Phase = determinePhaseEnhanced(*result)
	result.DisplayStatus = fmt.Sprintf("%s | %s | Asing: %s | Retail: %s",
		result.Phase, result.FormattedNetValue, result.FormattedForeignNet, helpers.FormatBigNumber(result.RetailNet))

	return result, nil
}

// determinePhaseEnhanced uses broker group breakdown + metrics to detect
// DISTRIBUTION MASIF ke retail (retail FOMO) vs genuine ACCUMULATION.
func determinePhaseEnhanced(p models.ExodusFlowPhase) string {
	// Pergerakan harga signifikan didahulukan: NAV besar (mis. markup/rebound)
	// sering disertai net flow yang hampir balance antar broker, jadi tidak boleh
	// langsung dikategorikan SIDEWAYS hanya karena ratio net kecil.
	if p.PriceChangePct >= 5 {
		return fmt.Sprintf("MARK-UP / REBOUND (Harga %+.1f%% selama %d hari)",
			p.PriceChangePct, p.TotalDays)
	}
	if p.PriceChangePct <= -5 {
		return fmt.Sprintf("MARK-DOWN / KOREKSI (Harga %+.1f%% selama %d hari)",
			p.PriceChangePct, p.TotalDays)
	}

	// Net flow sangat kecil relatif terhadap total turnover -> harga moving
	// sideways, bukan distribusi/akumulasi masif. Ini prioritas pertama supaya
	// buy & sell yang hampir seimbang tidak salah dikategorikan.
	totalTurnover := p.TotalBuyValue + p.TotalSellValue
	if totalTurnover > 0 {
		netRatio := math.Abs(p.NetValue) / totalTurnover * 100
		if netRatio < 2 {
			return fmt.Sprintf("SIDEWAYS / NETRAL (Net flow %s cuma %.2f%% dari turnover %s)",
				helpers.FormatBigNumber(p.NetValue), netRatio, helpers.FormatBigNumber(totalTurnover))
		}
	}

	// Net positive tapi retail dominan + smart money ratio rendah = DISTRIBUSI MASIF (retail chasing)
	if p.NetValue > 0 {
		// Foreign exiting (negative) + retail buying (positive) = classic distribution
		if p.ForeignNetValue < 0 && p.RetailNet > 0 && p.RetailDominance > 25 {
			return "DISTRIBUSI MASIF (Retail FOMO, Foreign/Smart Money Exiting)"
		}
		if p.RetailDominance > 50 && p.SmartMoneyRatio < 30 {
			return "DISTRIBUSI MASIF (Retail Driven, Institution Exiting)"
		}
		if p.ForeignNetValue > 0 && p.SmartMoneyRatio > 60 {
			if p.SmartMoneyConsistency >= 60 && p.MomentumAccelerating && p.PriceConfirms {
				return "AKUMULASI SANGAT KUAT (SM Konsisten + Accelerating + Harga Konfirmasi)"
			}
			if p.SmartMoneyConsistency >= 60 {
				return "AKUMULASI KUAT (Smart Money Konsisten)"
			}
			return "AKUMULASI KUAT (Smart Money Leading)"
		}
		if p.ForeignNetValue > 0 {
			return "AKUMULASI (Institusi & Asing Net Buy)"
		}
		if len(p.Anomalies) > 0 {
			return "AKUMULASI EVENT-DRIVEN (Ada Broker Anomali Naik)"
		}
		return "AKUMULASI (Retail Driven)"
	}

	// Net negative + retail net negative (retail panic) = bisa ACCUMULATION BOTTOM (institution buy)
	if p.NetValue < 0 {
		if p.RetailNet < 0 && p.ForeignNetValue > 0 {
			return "AKUMULASI BOTTOM (Retail Panic, Foreign Buying)"
		}
		if p.ForeignNetValue < 0 {
			if p.SmartMoneyConsistency >= 60 {
				return "DISTRIBUSI KUAT (Asing Net Sell, Konsisten)"
			}
			return "DISTRIBUSI KUAT (Asing Net Sell)"
		}
		return "DISTRIBUSI"
	}

	return "NETRAL"
}

func determinePhase(p models.ExodusFlowPhase) string {
	if p.NetValue > 0 {
		if p.ForeignNetValue > 0 {
			return "AKUMULASI (Institusi & Asing Net Buy)"
		}
		return "AKUMULASI"
	}

	if p.NetValue < 0 {
		if p.ForeignNetValue < 0 {
			return "DISTRIBUSI (Asing Net Sell)"
		}
		return "DISTRIBUSI"
	}

	return "NETRAL"
}

func sortDesc(flows []models.ExodusBrokerFlow, ascendingForDist bool) []models.ExodusBrokerFlow {
	out := make([]models.ExodusBrokerFlow, len(flows))
	copy(out, flows)

	sort.SliceStable(out, func(i, j int) bool {
		if ascendingForDist {
			return out[i].NetValue < out[j].NetValue
		}
		return out[i].NetValue > out[j].NetValue
	})

	return out
}

func flowSideLabel(ratio float64) string {
	if ratio >= 50 {
		return "Akumulasi Kuat"
	}
	if ratio > 20 {
		return "Akumulasi Ringan"
	}
	if ratio <= -50 {
		return "Distribusi Kuat"
	}
	if ratio < -20 {
		return "Distribusi Ringan"
	}
	return "Netral"
}

// enrichGroupProfile menghitung konsistensi & momentum smart money (Asing+Inst)
// berdasarkan net flow harian. Konsistensi = hari net buy / hari aktif; momentum
// = net second half − net first half (membesar = accelerating).
func enrichGroupProfile(p *models.ExodusFlowPhase) {
	daily, err := repositories.GetExodusDailyGroupFlow(p.Symbol, p.StartDate, p.EndDate)
	if err != nil {
		return
	}

	// Net harian smart money & total hari
	var smartDaily []float64
	for _, d := range daily {
		if d.BrokerGroup == "FOREIGN" || d.BrokerGroup == "INSTITUTIONAL" {
			smartDaily = append(smartDaily, d.NetValue)
		}
	}
	if len(smartDaily) == 0 {
		return
	}

	// Konsistensi: berapa hari smart money net > 0
	buyDays := 0
	for _, v := range smartDaily {
		if v > 0 {
			buyDays++
		}
	}
	p.SmartMoneyActiveDays = buyDays
	if p.TotalDays > 0 {
		p.SmartMoneyConsistency = float64(buyDays) / float64(p.TotalDays) * 100
	}

	// Momentum: second half − first half net, dibagi berdasarkan tanggal aktual.
	// Batas paruh = titik tengah rentang window.
	winStart, _ := time.Parse("2006-01-02", p.StartDate)
	winEnd, _ := time.Parse("2006-01-02", p.EndDate)
	mid := winStart.Add(winEnd.Sub(winStart) / 2)

	firstHalf := 0.0
	secondHalf := 0.0
	firstDays := 0
	secondDays := 0
	for _, d := range daily {
		if d.BrokerGroup != "FOREIGN" && d.BrokerGroup != "INSTITUTIONAL" {
			continue
		}
		dt, _ := time.Parse("2006-01-02", d.TradeDate)
		if !dt.After(mid) {
			firstHalf += d.NetValue
			firstDays++
		} else {
			secondHalf += d.NetValue
			secondDays++
		}
	}
	if firstDays > 0 || secondDays > 0 {
		p.FirstHalfNet = firstHalf
		p.SecondHalfNet = secondHalf
		// Tanggal representatif: awal paruh pertama & akhir paruh kedua
		if firstDays > 0 {
			p.FirstHalfDate = fmt.Sprintf("%s → %s", p.StartDate, mid.Format("2006-01-02"))
		}
		if secondDays > 0 {
			p.SecondHalfDate = fmt.Sprintf("%s → %s", mid.AddDate(0, 0, 1).Format("2006-01-02"), p.EndDate)
		}
	}
	p.SmartMoneyMomentum = secondHalf - firstHalf
	p.MomentumAccelerating = secondHalf > firstHalf && firstHalf != 0
}

// enrichPriceVolume menghitung perubahan harga selama window & volume spike
// dibanding baseline rata-rata volume sebelum window.
func enrichPriceVolume(p *models.ExodusFlowPhase) {
	// Ambil harga window (dengan tambahan baseline ~30 hari sebelum start utk avg volume)
	start, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return
	}
	end, err := time.Parse("2006-01-02", p.EndDate)
	if err != nil {
		return
	}
	baselineStart := start.AddDate(0, 0, -30).Format("2006-01-02")

	prices, err := repositories.GetPriceWindow(p.Symbol, baselineStart, p.EndDate)
	if err != nil || len(prices) == 0 {
		return
	}

	// Baseline avg volume = 30 hari sebelum window (hari tanpa data dimulai setelah start)
	var baselineVols []float64
	for _, px := range prices {
		d, _ := time.Parse("2006-01-02", px.TradeDate)
		if d.Before(start) {
			baselineVols = append(baselineVols, px.Volume)
		}
	}

	// Window prices
	var winClose []float64
	var winVol []float64
	for _, px := range prices {
		d, _ := time.Parse("2006-01-02", px.TradeDate)
		if !d.Before(start) && !d.After(end) {
			winClose = append(winClose, px.Close)
			winVol = append(winVol, px.Volume)
		}
	}
	if len(winClose) == 0 {
		return
	}

	if len(winClose) >= 2 {
		prev := winClose[0]
		last := winClose[len(winClose)-1]
		if prev > 0 {
			p.PriceChangePct = (last - prev) / prev * 100
		}
	}

	// Volume spike
	baseAvg := 0.0
	if len(baselineVols) > 0 {
		s := 0.0
		for _, v := range baselineVols {
			s += v
		}
		baseAvg = s / float64(len(baselineVols))
	}
	winAvg := 0.0
	if len(winVol) > 0 {
		s := 0.0
		for _, v := range winVol {
			s += v
		}
		winAvg = s / float64(len(winVol))
	}
	if baseAvg > 0 {
		p.VolumeSpikeRatio = winAvg / baseAvg
		p.HasVolumeSpike = p.VolumeSpikeRatio >= 1.5
	}

	// Price confirmation: smart money net searah dengan pergerakan harga.
	smartNet := p.ForeignNetValue + p.InstitutionalNet
	if p.PriceChangePct >= 1 && smartNet > 0 {
		p.PriceConfirms = true
	}
	if p.PriceChangePct <= -1 && smartNet < 0 {
		p.PriceConfirms = true
	}
}

// enrichAnomalies mendeteksi broker dengan net harian sangat anomali memakai
// Modified Z-Score (median + MAD) yang robust terhadap outlier.
func enrichAnomalies(p *models.ExodusFlowPhase) {
	// Baseline history lebih panjang (60 hari sebelum end) supaya median stabil.
	end, err := time.Parse("2006-01-02", p.EndDate)
	if err != nil {
		return
	}
	baselineFrom := end.AddDate(0, 0, -60).Format("2006-01-02")

	rows, err := repositories.GetExodusDailyBrokerNet(p.Symbol, baselineFrom, p.EndDate)
	if err != nil {
		return
	}

	// Kelompokkan per broker
	byBroker := map[string][]models.DailyBrokerNet{}
	for _, r := range rows {
		byBroker[r.BrokerCode] = append(byBroker[r.BrokerCode], r)
	}

	// Ambil daftar market maker saham ini
	mm, _ := repositories.GetMarketMakers(p.Symbol)

	// Threshold: nilai transaksi anomali minimal (Rp) untuk dianggap signifikan.
	const minAnomalyValue = 2e9 // Rp 2 Miliar

	// Hanya flag anomali yang terjadi di dalam window analisis; baseline
	// (lebih panjang) hanya dipakai untuk menghitung median/MAD yang stabil.
	winStart, _ := time.Parse("2006-01-02", p.StartDate)
	winEnd, _ := time.Parse("2006-01-02", p.EndDate)

	var anomalies []models.ExodusAnomaly
	for _, series := range byBroker {
		if len(series) < 3 {
			continue
		}
		vals := make([]float64, len(series))
		for i, s := range series {
			vals[i] = s.NetValue
		}
		sorted := make([]float64, len(vals))
		copy(sorted, vals)
		sort.Float64s(sorted)

		median := medianOf(sorted)
		// MAD
		absDev := make([]float64, len(sorted))
		for i, v := range sorted {
			absDev[i] = math.Abs(v - median)
		}
		sort.Float64s(absDev)
		mad := medianOf(absDev)

		for _, s := range series {
			if s.NetValue < minAnomalyValue {
				continue
			}
			// Lewati anomali di luar window analisis
			d, err := time.Parse("2006-01-02", s.TradeDate)
			if err != nil || d.Before(winStart) || d.After(winEnd) {
				continue
			}
			var zMod float64
			if mad > 0 {
				zMod = 0.6745 * (s.NetValue - median) / mad
			} else {
				// fallback jika MAD = 0 (semua nilai sama) -> gunakan std kecil
				zMod = 0
			}
			if zMod >= 3 {
				anomalies = append(anomalies, models.ExodusAnomaly{
					StockCode:    p.Symbol,
					BrokerCode:   s.BrokerCode,
					BrokerType:   s.BrokerType,
					TradeDate:    s.TradeDate,
					NetValue:     s.NetValue,
					FormattedNet: helpers.FormatBigNumber(s.NetValue),
					ZScore:       math.Round(zMod*10) / 10,
					IsMarketMaker: mm[s.BrokerCode],
				})
			}
		}
	}

	// Urutkan: anomaly paling anomali (Z-score tertinggi) dulu
	sort.Slice(anomalies, func(i, j int) bool {
		return anomalies[i].ZScore > anomalies[j].ZScore
	})
	if len(anomalies) > 5 {
		anomalies = anomalies[:5]
	}
	p.Anomalies = anomalies
}

func medianOf(sorted []float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

// GetStockDetail merangkum data detail sebuah saham untuk halaman ticker:
// chart harga, volume per broker, dan summary buy/sell per broker per hari.
// rangeStr (1m|3m|1y) menentukan rentang waktu; dari/to eksplisit mengalahkan
// range jika keduanya diisi.
func GetStockDetail(symbol, fromStr, toStr, rangeStr string) (*models.TickerDetail, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	now := time.Now()

	rangeDays := 0
	rangeStr = strings.ToLower(strings.TrimSpace(rangeStr))
	switch rangeStr {
	case "1m":
		rangeDays = 30
	case "3m":
		rangeDays = 90
	case "1y":
		rangeDays = 365
	case "":
	default:
		return nil, fmt.Errorf("invalid range, use 1m|3m|1y")
	}

	if rangeDays > 0 {
		fromStr = now.AddDate(0, 0, -rangeDays).Format("2006-01-02")
		toStr = now.Format("2006-01-02")
	}
	if strings.TrimSpace(fromStr) == "" {
		fromStr = now.AddDate(0, 0, -29).Format("2006-01-02")
	}
	if strings.TrimSpace(toStr) == "" {
		toStr = now.Format("2006-01-02")
	}

	startDate, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return nil, fmt.Errorf("invalid from date, format: YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return nil, fmt.Errorf("invalid to date, format: YYYY-MM-DD")
	}
	if startDate.After(endDate) {
		return nil, fmt.Errorf("from cannot be after to")
	}

	from := startDate.Format("2006-01-02")
	to := endDate.Format("2006-01-02")

	chart, err := repositories.GetTickerPriceBars(symbol, from, to)
	if err != nil {
		return nil, err
	}
	byBroker, err := repositories.GetTickerBrokerVolume(symbol, from, to)
	if err != nil {
		return nil, err
	}
	brokerDaily, err := repositories.GetTickerBrokerSummary(symbol, from, to)
	if err != nil {
		return nil, err
	}

	stockName, _ := repositories.GetStockName(symbol)

	return &models.TickerDetail{
		Symbol:      symbol,
		StockName:   stockName,
		Range:       rangeStr,
		From:        from,
		To:          to,
		PriceChart:  chart,
		ByBroker:    byBroker,
		BrokerDaily: brokerDaily,
	}, nil
}

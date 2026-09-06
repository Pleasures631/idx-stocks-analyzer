package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"indonesia-stocks-api/internal/constants"
	"indonesia-stocks-api/internal/handlers"
	"indonesia-stocks-api/internal/helpers"
	"indonesia-stocks-api/internal/models"
	"indonesia-stocks-api/internal/repositories"
	"indonesia-stocks-api/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

type TelegramBot struct {
	api *tgbotapi.BotAPI
	db  *sqlx.DB
}

func NewTelegramBot(db *sqlx.DB) (*TelegramBot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN tidak ditemukan di .env")
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{
		api: botAPI,
		db:  db,
	}, nil
}

// StartBot menjalankan listener Telegram secara asinkron
func (b *TelegramBot) StartBot() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	log.Printf("🤖 Telegram Bot %s aktif!", b.api.Self.UserName)

	for update := range updates {
		// Handle Callback Query (Klik Tombol Inline)
		if update.CallbackQuery != nil {
			go b.handleCallbackQuery(update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			go b.handleCommand(update.Message)
		} else if update.Message.Text != "" {
			go b.handleTextMessage(update.Message)
		}
	}
}

func (b *TelegramBot) handleCommand(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	command := strings.ToLower(msg.Command())
	args := msg.CommandArguments()

	switch command {
	case "start":
		replyText := "📊 *Bot Indonesia Stocks API*\n\nPerintah tersedia:\n- `/analyze <TICKER>` : Analisis Broker Flow (Contoh: `/analyze CUAN`)"
		b.sendMessage(chatID, replyText, nil)

	case "analyze", "analyze_flow", "analyzeflow", "flow":
		b.handleAnalyzeFlowResponse(chatID, args)

	case "fetchdaily":
		go b.handleFetchDaily(chatID)

	default:
		b.sendMessage(chatID, "❌ Perintah tidak dikenali. Ketik `/start` untuk bantuan.", nil)
	}
}

func (b *TelegramBot) handleTextMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	text := strings.TrimSpace(msg.Text)
	lowerText := strings.ToLower(text)

	if strings.HasPrefix(lowerText, "analyze ") || strings.HasPrefix(lowerText, "analyze_flow ") {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) > 1 {
			b.handleAnalyzeFlowResponse(chatID, parts[1])
		}
	}
}

func (b *TelegramBot) handleAnalyzeFlowResponse(chatID int64, args string) {
	parts := strings.Fields(args)
	if len(parts) == 0 {
		b.sendMessage(chatID, "⚠️ Format salah. Gunakan: `/analyze <TICKER>` atau `analyze CUAN`\nContoh: `/analyze CUAN` atau `/analyze CUAN 2026-02-01 2026-02-17`", nil)
		return
	}

	symbol := strings.ToUpper(strings.TrimSpace(parts[0]))
	fromStr := ""
	toStr := ""

	if len(parts) >= 2 {
		fromStr = parts[1]
	}
	if len(parts) >= 3 {
		toStr = parts[2]
	}

	flow, err := services.AnalyzeExodusBrokerFlowService(symbol, fromStr, toStr)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("⚠️ *Gagal menganalisis flow %s*\nError: %s", symbol, err.Error()), nil)
		return
	}

	// JIKA DATA TIDAK DITEMUKAN: Bikin tombol penawaran Fetch
	if flow.TotalBrokers == 0 {
		msgText := fmt.Sprintf("⚠️ Data broker flow untuk *%s* tidak ditemukan dalam periode `%s` s/d `%s`.\n\nApakah kamu ingin men-fetch datanya sekarang?", symbol, flow.StartDate, flow.EndDate)

		// Create inline keyboard buttons
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📥 Bulan Ini (MTD)", fmt.Sprintf("fetch:%s:mtd", symbol)),
				tgbotapi.NewInlineKeyboardButtonData("📥 30 Hari Terakhir", fmt.Sprintf("fetch:%s:30", symbol)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "cancel_fetch"),
			),
		)

		b.sendMessage(chatID, msgText, &keyboard)
		return
	}

	// JIKA DATA ADA: Tampilkan analisis biasa
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 *ANALISIS BROKER FLOW - %s*\n", flow.Symbol))
	sb.WriteString(fmt.Sprintf("🗓 *Periode*: `%s` s/d `%s` (%d hari)\n\n", flow.StartDate, flow.EndDate, flow.TotalDays))
	sb.WriteString(fmt.Sprintf("📌 *Fase*: *%s*\n", flow.Phase))
	sb.WriteString(fmt.Sprintf("💰 *Net Flow*: `%s` (Buy: %s | Sell: %s)\n", flow.FormattedNetValue, flow.FormattedBuyValue, flow.FormattedSellValue))
	sb.WriteString(fmt.Sprintf("🌍 *Foreign Net*: `%s` | *Lokal*: `%s`\n", flow.FormattedForeignNet, helpers.FormatBigNumber(flow.LocalNetValue)))

	// --- Konfirmasi & Nuansa Baru ---
	if flow.SmartMoneyActiveDays > 0 {
		sb.WriteString("\n📊 *SINYAL SMART MONEY*\n")

		// Konsistensi
		consLabel := "⚠️ Jarang (sporadis)"
		if flow.SmartMoneyConsistency >= 75 {
			consLabel = "🟢 Konsisten"
		} else if flow.SmartMoneyConsistency >= 50 {
			consLabel = "🟡 Sedang"
		}
		sb.WriteString(fmt.Sprintf("▫️ *Konsistensi*: %s (%d/%d hari net buy, %.0f%%)\n",
			consLabel, flow.SmartMoneyActiveDays, flow.TotalDays, flow.SmartMoneyConsistency))

		// Momentum
		momLabel := "❄️ Melambat"
		momEmoji := "📉"
		momDetail := "net smart money di paruh kedua MENGECIL dibanding paruh pertama"
		if flow.MomentumAccelerating {
			momLabel = "🔥 Accelerating"
			momEmoji = "📈"
			momDetail = "net smart money di paruh kedua MEMBESAR dibanding paruh pertama"
		}
		sb.WriteString(fmt.Sprintf("▫️ *Momentum*: %s %s\n", momEmoji, momLabel))
		sb.WriteString(fmt.Sprintf("   └ %s\n", momDetail))
		if flow.FirstHalfDate != "" {
			sb.WriteString(fmt.Sprintf("   ├ Paruh 1 (%s): net `%s`\n", flow.FirstHalfDate, helpers.FormatBigNumber(flow.FirstHalfNet)))
		}
		if flow.SecondHalfDate != "" {
			sb.WriteString(fmt.Sprintf("   └ Paruh 2 (%s): net `%s`\n", flow.SecondHalfDate, helpers.FormatBigNumber(flow.SecondHalfNet)))
		}

		// Price confirmation
		pxLabel := "⚠️ Tidak searah"
		pxEmoji := "❌"
		pxDetail := "harga bergerak berlawanan/mendatar vs net smart money (perhatian)"
		if flow.PriceConfirms {
			pxLabel = "✅ Searah"
			pxEmoji = "✅"
			pxDetail = "harga bergerak sejalan dengan net smart money"
		}
		sb.WriteString(fmt.Sprintf("▫️ *Harga*: %s %s\n", pxEmoji, pxLabel))
		sb.WriteString(fmt.Sprintf("   └ %s, perubahan harga %s%% selama %d hari\n",
			pxDetail, fmt.Sprintf("%.1f", flow.PriceChangePct), flow.TotalDays))

		// Volume spike
		if flow.HasVolumeSpike {
			sb.WriteString(fmt.Sprintf("▫️ *Volume*: 🔥 Spike %.1fx dari rata-rata (aktivitas tinggi)\n", flow.VolumeSpikeRatio))
		}
	}

	// --- Anomali Deteksi ---
	if len(flow.Anomalies) > 0 {
		sb.WriteString("\n🚨 *DETEKSI ANOMALI* (Z-Score):\n")
		for _, a := range flow.Anomalies {
			mmTag := ""
			if a.IsMarketMaker {
				mmTag = " 🏦 MM"
			}
			sb.WriteString(fmt.Sprintf("• *%s* (%s) `%s` @%s (Z=%s)%s\n", a.BrokerCode, a.BrokerType, a.FormattedNet, a.TradeDate, fmt.Sprintf("%.1f", a.ZScore), mmTag))
		}
	}

	sb.WriteString("\n🟢 *Top Accumulation*:\n")
	if len(flow.BrokersAccumulation) == 0 {
		sb.WriteString("  _(Tidak ada)_\n")
	} else {
		for _, bk := range flow.BrokersAccumulation {
			sb.WriteString(fmt.Sprintf("• *%s* (%s) : `%s`\n", bk.BrokerCode, bk.BrokerType, bk.FormattedNetValue))
		}
	}

	sb.WriteString("\n🔴 *Top Distribution*:\n")
	if len(flow.BrokersDistribution) == 0 {
		sb.WriteString("  _(Tidak ada)_\n")
	} else {
		for _, bk := range flow.BrokersDistribution {
			sb.WriteString(fmt.Sprintf("• *%s* (%s) : `%s`\n", bk.BrokerCode, bk.BrokerType, bk.FormattedNetValue))
		}
	}

	b.sendMessage(chatID, sb.String(), nil)
}

// Handler khusus Callback Query (Tombol diklik)
func (b *TelegramBot) handleCallbackQuery(cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	data := cb.Data

	// Hentikan loading indicator di tombol Telegram
	callbackAnswer := tgbotapi.NewCallback(cb.ID, "")
	b.api.Request(callbackAnswer)

	if data == "cancel_fetch" {
		b.sendMessage(chatID, "🚫 Proses fetching dibatalkan.", nil)
		return
	}

	if strings.HasPrefix(data, "fetch:") {
		parts := strings.Split(data, ":")
		if len(parts) == 3 {
			symbol := parts[1]
			days := parts[2]

			b.sendMessage(chatID, fmt.Sprintf("⏳ Sedang mengambil data *%s* untuk %s hari terakhir, mohon tunggu...", symbol, days), nil)

			// Panggil fungsi eksekusi sync secara langsung
			go b.executeFetchAndReanalyze(chatID, symbol, days)
		}
	}
}

// Menjalankan Fetching dan otomatis langsung Re-Analyze hasilnya
func (b *TelegramBot) executeFetchAndReanalyze(chatID int64, symbol string, mode string) {
	now := time.Now()
	endDate := now
	var startDate time.Time

	if mode == "mtd" {
		// Set tanggal ke awal bulan berjalan jam 00:00:00 (misal: 2026-08-01)
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	} else {
		var days int
		fmt.Sscanf(mode, "%d", &days)
		if days <= 0 {
			days = 30
		}
		startDate = endDate.AddDate(0, 0, -days)
	}

	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	b.sendMessage(chatID, fmt.Sprintf("⏳ Menarik data *%s* dari `%s` s/d `%s`...", symbol, startDateStr, endDateStr), nil)

	successDays := 0

	// Loop sync harian
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")

		data, err := services.FetchExodusMarketDetector(symbol, dateStr, dateStr)
		if err != nil {
			continue
		}

		if len(data.BrokerSummary.BrokersBuy) == 0 && len(data.BrokerSummary.BrokersSell) == 0 {
			continue
		}

		// Panggil mapper
		rows := handlers.MapExodusBrokerSummaryToModel(data.BrokerSummary)

		if err := repositories.UpsertExodusBrokerSummary(rows); err != nil {
			continue
		}

		successDays++
	}

	if successDays == 0 {
		b.sendMessage(chatID, fmt.Sprintf("❌ Data *%s* tidak ditemukan di server sumber untuk periode tersebut.", symbol), nil)
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("✅ Berhasil sync %d hari data transaksi *%s*!", successDays, symbol), nil)

	// Jalankan ulang analisis broker flow secara otomatis
	b.handleAnalyzeFlowResponse(chatID, fmt.Sprintf("%s %s %s", symbol, startDateStr, endDateStr))
}

// handleFetchDaily menjalankan sync harian untuk semua data
func (b *TelegramBot) handleFetchDaily(chatID int64) {
	start := time.Now()

	// 1. Ambil tanggal terakhir sync dari database
	lastSyncDate, err := repositories.GetLastSyncDate("daily_sync")
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Gagal mengambil data last sync: %s", err.Error()), nil)
		return
	}

	// 2. Tentukan tanggal mulai sync
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startDate := today

	if lastSyncDate != nil {
		// Jika sudah pernah sync, mulai dari hari setelah last sync
		startDate = lastSyncDate.AddDate(0, 0, 1)
	}

	// 3. Cek apakah sudah sync hari ini
	if !startDate.After(today) {
		// Hitung jumlah hari yang perlu di-sync
		daysToSync := int(today.Sub(startDate).Hours()/24) + 1
		b.sendMessage(chatID, fmt.Sprintf("⏳ Memulai sync harian dari `%s` s/d `%s` (%d hari)...", startDate.Format("2006-01-02"), today.Format("2006-01-02"), daysToSync), nil)
	} else {
		b.sendMessage(chatID, fmt.Sprintf("✅ Sudah sync sampai tanggal `%s`. Tidak ada data baru.", today.Format("2006-01-02")), nil)
		return
	}

	startDateStr := startDate.Format("2006-01-02")
	endDateStr := today.Format("2006-01-02")

	// 4. Sync IDX Stocks dan Brokers (tidak butuh parameter tanggal)
	b.sendMessage(chatID, "🔄 Sync IDX Stocks...", nil)
	if err := b.syncIDXStocks(); err != nil {
		log.Printf("[fetchDaily] Gagal sync IDX stocks: %v", err)
	} else {
		b.sendMessage(chatID, "✅ IDX Stocks synced", nil)
	}

	// 4b. Ambil broker insider KSEI melalui browser session Stockbit.
	b.sendMessage(chatID, "🔄 Sync Stockbit Insider KSEI...", nil)
	insiderResult, err := b.syncInsiderMarketMakers()
	if err != nil {
		log.Printf("[fetchDaily] Gagal sync Stockbit Insider: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("⚠️ Gagal sync Insider KSEI: %s", err.Error()), nil)
	} else {
		b.sendMessage(chatID, fmt.Sprintf("✅ Insider KSEI: %d mapping baru", insiderResult.NewMappings), nil)
	}

	b.sendMessage(chatID, "🔄 Sync IDX Brokers...", nil)
	if err := b.syncIDXBrokers(); err != nil {
		log.Printf("[fetchDaily] Gagal sync IDX brokers: %v", err)
	} else {
		b.sendMessage(chatID, "✅ IDX Brokers synced", nil)
	}

	// 5. Sync Exodus Broker Summary (bulkfetch)
	b.sendMessage(chatID, fmt.Sprintf("🔄 Sync Exodus Broker Summary (%s s/d %s)...", startDateStr, endDateStr), nil)
	exodusResult, err := b.syncExodusBrokerSummary(startDateStr, endDateStr)
	if err != nil {
		log.Printf("[fetchDaily] Gagal sync Exodus broker summary: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Gagal sync Exodus: %s", err.Error()), nil)
	} else {
		b.sendMessage(chatID, fmt.Sprintf("✅ Exodus Broker Summary: %d hari success, %d skipped", exodusResult.SuccessDays, exodusResult.SkippedDays), nil)
	}

	// 6. Sync Trading Summary
	b.sendMessage(chatID, fmt.Sprintf("🔄 Sync Trading Summary (%s s/d %s)...", startDateStr, endDateStr), nil)
	tradingResult, err := b.syncTradingSummary(startDateStr, endDateStr)
	if err != nil {
		log.Printf("[fetchDaily] Gagal sync Trading Summary: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("❌ Gagal sync Trading Summary: %s", err.Error()), nil)
	} else {
		b.sendMessage(chatID, fmt.Sprintf("✅ Trading Summary: %d hari success, %d failed", tradingResult.SuccessDays, len(tradingResult.FailedDays)), nil)
	}

	// 7. Update last sync date ke database
	b.sendMessage(chatID, fmt.Sprintf("🔄 Sync IHSG Price Chart (%s s/d %s)...", startDateStr, endDateStr), nil)
	if chartPoints, err := services.SyncStockbitIHSGChart(context.Background(), "IHSG", startDateStr, endDateStr); err != nil {
		log.Printf("[fetchDaily] Gagal sync IHSG chart: %v", err)
		b.sendMessage(chatID, fmt.Sprintf("⚠️ Gagal sync IHSG Price Chart: %s", err.Error()), nil)
	} else {
		b.sendMessage(chatID, fmt.Sprintf("✅ IHSG Price Chart synced: %d titik", chartPoints), nil)
	}

	// 8. Update last sync date ke database
	if err := repositories.UpdateLastSyncDate("daily_sync", today); err != nil {
		log.Printf("[fetchDaily] Gagal update last sync date: %v", err)
	}

	duration := time.Since(start)
	b.sendMessage(chatID, fmt.Sprintf("✅ *Sync harian selesai!*\n\n📊 Ringkasan:\n- Exodus: %d hari\n- Trading: %d hari\n- Waktu: %s", exodusResult.SuccessDays, tradingResult.SuccessDays, duration.Round(time.Second)), nil)
}

// syncIDXStocks melakukan sync data saham dari IDX
func (b *TelegramBot) syncIDXStocks() error {
	data, err := services.FetchIDX[models.IDXStock](constants.IDXBaseURL, constants.ModuleStockData, constants.ServiceStocksList)
	if err != nil {
		return err
	}

	stocks := make([]models.StocksList, 0, len(data))
	for _, d := range data {
		stocks = append(stocks, handlers.MapIDXStockToModel(d))
	}

	return repositories.UpsertStocks(stocks)
}

type InsiderSyncResult struct {
	Pages       int
	Records     int
	NewMappings int
}

func (b *TelegramBot) syncInsiderMarketMakers() (*InsiderSyncResult, error) {
	records, pages, err := services.CrawlKSEIInsiderWithTimeout()
	if err != nil {
		return nil, err
	}
	newMappings, err := repositories.UpsertInsiderMarketMakers(records)
	if err != nil {
		return nil, err
	}
	return &InsiderSyncResult{Pages: pages, Records: len(records), NewMappings: newMappings}, nil
}

// syncIDXBrokers melakukan sync data broker dari IDX
func (b *TelegramBot) syncIDXBrokers() error {
	data, err := services.FetchIDX[models.IDXBroker](constants.IDXBaseURL, constants.ModuleExchangeMember, constants.ServiceBrokerList)
	if err != nil {
		return err
	}

	brokers := make([]models.BrokerList, 0, len(data))
	for _, d := range data {
		brokers = append(brokers, handlers.MapIDXBrokerToModel(d))
	}

	return repositories.UpsertBrokers(brokers)
}

type ExodusSyncResult struct {
	SuccessDays int
	SkippedDays int
	FailedDays  int
}

// syncExodusBrokerSummary melakukan sync broker summary dari Exodus
func (b *TelegramBot) syncExodusBrokerSummary(startDate, endDate string) (*ExodusSyncResult, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}

	stocks, err := repositories.GetActiveStockCodes()
	if err != nil {
		return nil, err
	}

	// Cek data yang sudah ada
	existingPairs, err := repositories.GetExistingExodusStockDates(startDate, endDate)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]map[string]struct{})
	for _, p := range existingPairs {
		dateKey := p.TradeDate.Format("2006-01-02")
		if existing[p.StockCode] == nil {
			existing[p.StockCode] = make(map[string]struct{})
		}
		existing[p.StockCode][dateKey] = struct{}{}
	}

	result := &ExodusSyncResult{}

	for _, symbol := range stocks {
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			if _, ok := existing[symbol][dateStr]; ok {
				result.SkippedDays++
				continue
			}

			data, err := services.FetchExodusMarketDetector(symbol, dateStr, dateStr)
			if err != nil {
				result.FailedDays++
				continue
			}

			if len(data.BrokerSummary.BrokersBuy) == 0 && len(data.BrokerSummary.BrokersSell) == 0 {
				continue
			}

			rows := handlers.MapExodusBrokerSummaryToModel(data.BrokerSummary)
			if err := repositories.UpsertExodusBrokerSummary(rows); err != nil {
				result.FailedDays++
				continue
			}

			result.SuccessDays++
		}
	}

	return result, nil
}

type TradingSyncResult struct {
	SuccessDays int
	FailedDays  []string
}

// syncTradingSummary melakukan sync trading summary dari IDX
func (b *TelegramBot) syncTradingSummary(startDate, endDate string) (*TradingSyncResult, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}

	result := &TradingSyncResult{}

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("20060102") // IDX format YYYYMMDD

		data, err := services.FetchIDX[models.TradingSummary](
			constants.IDXBaseURL,
			constants.ModuleTradingSummary,
			constants.ServiceStockSummary,
			dateStr,
		)
		if err != nil {
			result.FailedDays = append(result.FailedDays, d.Format("2006-01-02"))
			continue
		}

		tradingSummary := make([]models.TradingSummaryDB, 0, len(data))
		for _, d := range data {
			tradingSummary = append(tradingSummary, handlers.MapIDXTradingSummaryToModel(d))
		}

		if err := repositories.InsertTradingSummary(tradingSummary); err != nil {
			result.FailedDays = append(result.FailedDays, d.Format("2006-01-02"))
			continue
		}

		result.SuccessDays++
		time.Sleep(300 * time.Millisecond) // Rate limit
	}

	return result, nil
}

func (b *TelegramBot) sendMessage(chatID int64, text string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	response := tgbotapi.NewMessage(chatID, text)
	response.ParseMode = "Markdown"
	if keyboard != nil {
		response.ReplyMarkup = keyboard
	}

	_, err := b.api.Send(response)
	if err != nil {
		// Fallback ke plain text jika Markdown parsing gagal
		response.ParseMode = ""
		b.api.Send(response)
	}
}

func (b *TelegramBot) SendNotification(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	_, err := b.api.Send(msg)
	return err
}

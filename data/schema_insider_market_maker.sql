-- Migration: store KSEI insider broker codes per stock.
-- Run once after the existing m_list_stocks table is available.
ALTER TABLE m_list_stocks
  ADD COLUMN market_maker TEXT NULL AFTER is_active;

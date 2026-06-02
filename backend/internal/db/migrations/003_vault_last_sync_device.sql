-- Track which device last synced each vault (for dashboard vault cards)
ALTER TABLE vaults ADD COLUMN last_sync_device_id TEXT;
ALTER TABLE vaults ADD COLUMN last_sync_device_name TEXT;

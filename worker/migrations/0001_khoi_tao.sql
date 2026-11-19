-- Cau hinh: 1 dong JSON duy nhat (k='cau_hinh'). Dem dang nhap sai: k='sai_pin'.
CREATE TABLE IF NOT EXISTS cai_dat (k TEXT PRIMARY KEY, v TEXT NOT NULL);
-- May con bao ve moi lan hoi lenh (~30 giay).
CREATE TABLE IF NOT EXISTS may (
  id TEXT PRIMARY KEY, ten TEXT, phien_ban TEXT, lan_cuoi INTEGER, dang_chan INTEGER, ghi_chu TEXT
);
CREATE TABLE IF NOT EXISTS nhat_ky (
  id INTEGER PRIMARY KEY AUTOINCREMENT, ts INTEGER NOT NULL, may TEXT, loai TEXT NOT NULL, noi_dung TEXT
);
CREATE INDEX IF NOT EXISTS idx_nhat_ky_ts ON nhat_ky(ts);

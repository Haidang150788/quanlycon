#!/usr/bin/env bash
# Dung file cai cho may Windows cua con. Can Go >= 1.22.
# Doc cau hinh tu ./cau-hinh.env (khong dua len git), vd:
#   SERVER=https://quanlycon.ten-ban.workers.dev
#   CAU_HOI="Hỏi bố mẹ nếu muốn chơi."
#   TOKEN=<AGENT_TOKEN>            # co thi dung them file CaiDat-QuanLyCon.exe bam dup la cai
# Ra: dist/quanlycon.exe (khong chua ma) va dist/CaiDat-QuanLyCon.exe (chua ma — KHONG chia se).
set -euo pipefail
GOC="$(cd "$(dirname "$0")/.." && pwd)"
[ -f "$GOC/cau-hinh.env" ] && set -a && . "$GOC/cau-hinh.env" && set +a
SERVER="${SERVER:-}"; CAU_HOI="${CAU_HOI:-}"; TOKEN="${TOKEN:-}"
mkdir -p "$GOC/dist"
cd "$GOC/agent"
go test ./...
GOOS=windows GOARCH=amd64 go vet ./...
LD="-s -w"
[ -n "$SERVER" ] && LD="$LD -X 'main.ServerMacDinh=$SERVER'"
[ -n "$CAU_HOI" ] && LD="$LD -X 'main.CauHoiBo=$CAU_HOI'"
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$LD" -o "$GOC/dist/quanlycon.exe" .
echo "Đã dựng dist/quanlycon.exe"
if [ -n "$TOKEN" ]; then
  GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$LD -X main.tokenNhung=QLCTOKEN=$TOKEN" -o "$GOC/dist/CaiDat-QuanLyCon.exe" .
  echo "Đã dựng dist/CaiDat-QuanLyCon.exe (CHỨA MÃ — không gửi công khai, cài xong thì xoá)"
fi

#!/usr/bin/env bash
# Dong goi extension YouTube theo kenh thanh file .zip de tai len Chrome Web Store / Edge Add-ons.
# Chay bo kiem luat truoc. Ra: dist/quanlycon-youtube-<phien ban>.zip
set -euo pipefail
GOC="$(cd "$(dirname "$0")/.." && pwd)"
(cd "$GOC/worker" && npm test >/dev/null) || { echo "Bộ kiểm luật hỏng — chưa đóng gói"; exit 1; }
PB=$(node -p "require('$GOC/extension/manifest.json').version")
mkdir -p "$GOC/dist"
OUT="$GOC/dist/quanlycon-youtube-$PB.zip"
rm -f "$OUT"
cd "$GOC/extension"
zip -q -X "$OUT" manifest.json nen.js noi-dung.js trinh-phat.js luat-kenh.js icon-16.png icon-48.png icon-128.png
echo "Đã đóng gói $OUT"

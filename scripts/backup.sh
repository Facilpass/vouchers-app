#!/bin/bash
# Backup GFS rotation do volume vouchers-data
# Retenção: 7 daily + 4 weekly + 12 monthly + 1 yearly = ~365 dias funcional
# Cron: 0 3 * * *
set -euo pipefail

BACKUP_DIR=/opt/backups/vouchers
VOLUME=vouchers-data
DATE=$(date +%Y-%m-%d)
DOW=$(date +%u)
DOM=$(date +%d)
DOY=$(date +%j)

mkdir -p "$BACKUP_DIR"/{daily,weekly,monthly,yearly}

docker run --rm \
  -v "$VOLUME":/src:ro \
  -v "$BACKUP_DIR/daily":/dst \
  alpine:3.20 \
  tar czf "/dst/vouchers-$DATE.tar.gz" -C /src .

[ "$DOW" = "7" ]   && cp "$BACKUP_DIR/daily/vouchers-$DATE.tar.gz" "$BACKUP_DIR/weekly/"
[ "$DOM" = "01" ]  && cp "$BACKUP_DIR/daily/vouchers-$DATE.tar.gz" "$BACKUP_DIR/monthly/"
[ "$DOY" = "001" ] && cp "$BACKUP_DIR/daily/vouchers-$DATE.tar.gz" "$BACKUP_DIR/yearly/"

find "$BACKUP_DIR/daily"   -name "*.tar.gz" -mtime +7   -delete
find "$BACKUP_DIR/weekly"  -name "*.tar.gz" -mtime +31  -delete
find "$BACKUP_DIR/monthly" -name "*.tar.gz" -mtime +365 -delete

echo "$(date -Iseconds) backup ok" >> /var/log/vouchers-backup.log

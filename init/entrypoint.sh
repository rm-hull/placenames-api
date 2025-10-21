#!/bin/bash
set -e

DATABASE="/data/placenames.db"
ZIP_URL="https://www.arcgis.com/sharing/rest/content/items/208d9884575647c29f0dd5a1184e711a/data"
ZIP_FILE="/app/data.zip"

echo "Downloading CSV ZIP..."
wget -O "$ZIP_FILE" "$ZIP_URL"

echo "Extracting IPN_GB_2024.csv..."
unzip -p "$ZIP_FILE" "IPN_GB_2024.csv" > /app/data.csv

echo "Rewiring tempcode column..."
awk -F',' 'BEGIN{OFS=","} NR>1{$1=NR-1; print}' /app/data.csv | sponge /app/data.csv

echo "Creating SQLite database..."
rm -f $DATABASE
sqlite3 $DATABASE < /app/init.sql

echo "Done! Database is at $DATABASE"

#!/usr/bin/env bash
###############################################################################
# auto-test.sh — Rastgele hava kirliliği verisi üretir, ingest API'ine yollar
#
# Parametreler (tamamı isteğe bağlı):
#   --duration=60           # Çalışma süresi (saniye)
#   --rate=5                # Saniyedeki istek sayısı
#   --anomaly-chance=0.15   # Anomali olasılığı (0‑1)
#   --center-lat=41         # Merkez enlem
#   --center-lon=29         # Merkez boylam
#   --radius=25             # Koordinat yarıçapı (km)
#   --url=http://localhost:8080/ingest
#
# Örnek:
#   ./auto-test.sh --duration=120 --rate=5 --anomaly-chance=0.15
###############################################################################

######################## 1) Locale sabitle ####################################
export LC_ALL=C           # ← Sayılarda ondalık ayırıcı hep "."

######################## 2) Varsayılanlar ####################################
duration=60
rate=5
anomaly_chance=0.15
center_lat=40.7
center_lon=30
radius_km=25
url="http://localhost:8080/api/ingest"

######################## 3) Argümantaları ayrıştır ###########################
for arg in "$@"; do
  case $arg in
    --duration=*)        duration="${arg#*=}" ;;
    --rate=*)            rate="${arg#*=}" ;;
    --anomaly-chance=*)  anomaly_chance="${arg#*=}" ;;
    --center-lat=*)      center_lat="${arg#*=}" ;;
    --center-lon=*)      center_lon="${arg#*=}" ;;
    --radius=*)          radius_km="${arg#*=}" ;;
    --url=*)             url="${arg#*=}" ;;
    *) echo "Bilinmeyen argüman $arg"; exit 1 ;;
  esac
done

printf "▶️  Süre %ss | Rate %s req/s | Anomali %s%%\n" \
       "$duration" "$rate" "$(awk -v val="$anomaly_chance" 'BEGIN{printf "%.0f", val*100}')"

echo "Gönderilen JSON:"
######################## 4) Yardımcı fonksiyonlar #############################
rand() { awk -v s="$RANDOM" 'BEGIN{srand(s); printf "%.8f\n", rand()}'; }

random_point() {               # merkez (lat,lon) ve yarıçap (km) → nokta
  local latC=$1 lonC=$2 rad_km=$3
  local earth_radius=6371
  local u=$(rand) v=$(rand)
  local w=$(awk -v r="$rad_km" -v u="$u" -v er="$earth_radius" \
      'BEGIN{printf "%.10f\n", sqrt(u)*(r/er)}')
  local t=$(awk -v v="$v" 'BEGIN{printf "%.10f\n", 2*3.141592653*v}')
  local dlat=$(awk -v w="$w" -v t="$t" 'BEGIN{printf "%.10f\n", w*cos(t)}')
  local dlon=$(awk -v w="$w" -v t="$t" -v latC="$latC" \
      'BEGIN{printf "%.10f\n", w*sin(t)/cos(latC*3.141592653/180)}')
  local lat=$(awk -v latC="$latC" -v d="$dlat" \
      'BEGIN{printf "%.6f\n", latC + d*180/3.141592653}')
  local lon=$(awk -v lonC="$lonC" -v d="$dlon" \
      'BEGIN{printf "%.6f\n", lonC + d*180/3.141592653}')
  printf "%s %s\n" "$lat" "$lon"
}

######################## 5) Parametre/baz değer tabloları #####################
params=("PM2.5" "PM10" "NO2" "SO2" "O3")
baselines=(15 40 20 5 60)

######################## 6) Döngü ############################################
total_requests=$(awk -v d="$duration" -v r="$rate" 'BEGIN{print int(d*r)}')
interval=$(awk -v r="$rate" 'BEGIN{printf "%.3f\n", 1/r}')

for ((i=1; i<=total_requests; i++)); do
  # 6.1 Parametre seç
  idx=$((RANDOM % ${#params[@]}))
  param=${params[$idx]}
  base=${baselines[$idx]}

  # 6.2 Değer (anomali olasılığı)
  if awk "BEGIN{exit !($(rand) < $anomaly_chance)}"; then
    factor=$(awk -v r=$(rand) 'BEGIN{printf "%.2f\n", 3 + r*2}')   # 3‑5 kat
  else
    factor=$(awk -v r=$(rand) 'BEGIN{printf "%.2f", 0.8 + r*0.4}') # 0.8‑1.2
  fi
  value=$(awk -v b="$base" -v f="$factor" 'BEGIN{printf "%.2f\n", b*f}')

  # 6.3 Koordinat
  read lat lon < <(random_point "$center_lat" "$center_lon" "$radius_km")

  # 6.4 Zaman damgası
  ts=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

  # 6.5 JSON oluştur
  json=$(printf '{"latitude":%.6f,"longitude":%.6f,"parameter":"%s","value":%s,"timestamp":"%s"}' \
                "$lat" "$lon" "$param" "$value" "$ts")

  # 6.6 Gönder
curl -v -X POST -H "Content-Type: application/json" -d "$json" "$url" \
    || echo "❌ HTTP hata"

  # 6.7 Bekle
  sleep "$interval"
done

echo "✅ $total_requests istek gönderildi"
#!/bin/bash

# Default values for options
DURATION=60
RATE=1
ANOMALY_CHANCE=10

# Parse command-line arguments
for arg in "$@"; do
  case $arg in
    --duration=*)
      DURATION="${arg#*=}"
      shift
      ;;
    --rate=*)
      RATE="${arg#*=}"
      shift
      ;;
    --anomaly-chance=*)
      ANOMALY_CHANCE="${arg#*=}"
      shift
      ;;
    *)
      echo "Unknown option: $arg"
      exit 1
      ;;
  esac
done

# Initialize random seed globally
RANDOM_SEED=$(date +%s)
export RANDOM_SEED

# Set locale to ensure proper decimal formatting
export LC_ALL=C

# Function to generate random latitude and longitude
generate_latitude() {
  LAT=$(awk -v min=-90 -v max=90 'BEGIN{srand(); print min+rand()*(max-min)}')
  printf "%.6f\n" "$LAT"
}
generate_longitude() {
  LON=$(awk -v min=-180 -v max=180 'BEGIN{srand(); print min+rand()*(max-min)}')
  printf "%.6f\n" "$LON"
}

# Function to generate pollution value
generate_pollution_value() {
  if (( RANDOM % 100 < ANOMALY_CHANCE )); then
    # Generate an anomalous value
    echo $((RANDOM % 500 + 500))
  else
    # Generate a normal value
    echo $((RANDOM % 100 + 1))
  fi
}

# Function to send data to the API
send_data() {
  local LAT=$1
  local LON=$2
  local PARAM=$3
  local VALUE=$4
  echo "Sending data: Latitude: $LAT, Longitude: $LON, Parameter: $PARAM, Value: $VALUE"

  curl -X POST "http://localhost:8080/api/ingest" \
       -H "Content-Type: application/json" \
       -d "{
            \"latitude\": \"$LAT\",
            \"longitude\": \"$LON\",
            \"parameter\": \"$PARAM\",
            \"value\": \"$VALUE\"
           }"
}

# Main loop
END_TIME=$((SECONDS + DURATION))
while (( SECONDS < END_TIME )); do
  LAT=$(generate_latitude)
  LON=$(generate_longitude)
  echo "Generated coordinates: Latitude: $LAT, Longitude: $LON"

  for PARAM in "pm2.5" "pm10" "no2" "o3" "so2"; do
    POLLUTION=$(generate_pollution_value)
    send_data "$LAT" "$LON" "$PARAM" "$POLLUTION"
  done

  sleep $((1 / RATE))
done

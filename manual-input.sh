#!/bin/zsh

# Check parameter count
if [[ $# -ne 4 ]]; then
  echo "Usage: $0 <latitude> <longitude> <parameter> <value>"
  exit 1
fi

# Check if first and second arguments are valid numbers (latitude, longitude)
if [[ ! "$1" =~ ^-?[0-9]+(\.[0-9]+)?$ ]]; then
  echo "Error: First argument (latitude) must be a valid number."
  exit 1
fi

if [[ ! "$2" =~ ^-?[0-9]+(\.[0-9]+)?$ ]]; then
  echo "Error: Second argument (longitude) must be a valid number."
  exit 1
fi

# Check third argument is equal to "pm2.5", "pm10", "no2", "o3", "so2"
if [[ ! "$3" =~ ^(pm2\.5|pm10|no2|o3|so2)$ ]]; then
  echo "Error: Third argument (parameter) must be one of: pm2.5, pm10, no2, o3, so2."
  exit 1
fi

# Check if fourth argument is a valid number
if [[ ! "$4" =~ ^-?[0-9]+(\.[0-9]+)?$ ]]; then
  echo "Error: Fourth argument (value) must be a valid number."
  exit 1
fi



# Send request using curl
curl -X POST "http://localhost:8080/api/ingest" \
     -H "Content-Type: application/json" \
     -d "{
          \"latitude\": \"$1\",
          \"longitude\": \"$2\",
          \"parameter\": \"$3\",
          \"value\": \"$4\"
         }"

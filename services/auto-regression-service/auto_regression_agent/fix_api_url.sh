#!/bin/bash

# Fix the API URL in openAPISample.json to point to real Petstore API

echo "Fixing API URL in openAPISample.json..."

# Backup original
cp openAPISample.json openAPISample.json.backup

# Update the servers URL to point to real Petstore API
cat openAPISample.json | python3 -c "
import sys, json
data = json.load(sys.stdin)

# Update servers to point to real API
data['servers'] = [
    {
        'url': 'https://petstore3.swagger.io/api/v3',
        'description': 'Swagger Petstore Server'
    }
]

print(json.dumps(data, indent=2))
" > openAPISample.json.tmp

mv openAPISample.json.tmp openAPISample.json

echo "✅ Updated API URL to: https://petstore3.swagger.io/api/v3"
echo ""
echo "Now run a new test:"
echo "  ./run_full_test.sh"


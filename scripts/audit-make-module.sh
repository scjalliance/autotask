#!/bin/bash
#
# audit-make-module.sh - Inspect Make module configuration
#
# Usage: ./scripts/audit-make-module.sh <module_name>
#
# This script fetches a specific module from the Make API and displays:
# - Basic Info (name, label, description, typeId, public)
# - API Configuration (HTTP method, endpoint, etc.)
# - Expect section (input field definitions)
# - Interface section (output field definitions)
#

set -e

# Check if module name is provided
if [[ -z "$1" ]]; then
    echo "Error: Module name required"
    echo "Usage: $0 <module_name>"
    echo "Example: $0 companiesCreate"
    exit 1
fi

MODULE_NAME="$1"

# Load environment variables from .secrets/.env
if [[ ! -f .secrets/.env ]]; then
    echo "Error: .secrets/.env file not found"
    exit 1
fi

source .secrets/.env

# Check for required environment variable
if [[ -z "$MAKE_API_KEY" ]]; then
    echo "Error: MAKE_API_KEY not set in .secrets/.env"
    exit 1
fi

if [[ -z "$MAKE_API_URL" ]]; then
    echo "Error: MAKE_API_URL not set in .secrets/.env"
    exit 1
fi

# The Make custom app uses SDK endpoint with app ID
# App ID: scj-autotask-nn8loi, Version: 1
APP_ID="scj-autotask-nn8loi"
APP_VERSION="1"
BASE_ENDPOINT="${MAKE_API_URL}/sdk/apps/${APP_ID}/${APP_VERSION}/modules/${MODULE_NAME}"

echo "Auditing module: $MODULE_NAME"
echo ""

# Fetch Basic Info
RESPONSE=$(curl -s -H "Authorization: Token ${MAKE_API_KEY}" "${BASE_ENDPOINT}")

if [[ -z "$RESPONSE" ]]; then
    echo "Error: curl failed or no data returned"
    exit 1
fi

if echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    echo "API Error fetching basic info:"
    echo "$RESPONSE" | jq '.'
    exit 1
fi

# Extract appModule from response
BASIC_INFO=$(echo "$RESPONSE" | jq '.appModule')

# Check if module exists
if echo "$BASIC_INFO" | jq -e '.name' > /dev/null 2>&1; then
    echo "=== Basic Info ==="
    echo "$BASIC_INFO" | jq '{name: .name, label: .label, description: .description, typeId: .typeId, public: .public}'
    echo ""
else
    echo "Error: Module '$MODULE_NAME' not found"
    exit 1
fi

# Fetch API Configuration
echo "=== API Configuration ==="
API_CONFIG=$(curl -s -H "Authorization: Token ${MAKE_API_KEY}" "${BASE_ENDPOINT}/api")
if [[ -z "$API_CONFIG" ]]; then
    echo "No API configuration available"
else
    echo "$API_CONFIG" | jq '.'
fi
echo ""

# Fetch Expect (Input Fields)
echo "=== Expect (Input Fields) ==="
EXPECT=$(curl -s -H "Authorization: Token ${MAKE_API_KEY}" "${BASE_ENDPOINT}/expect")
if [[ -z "$EXPECT" ]]; then
    echo "No expect configuration available"
else
    echo "$EXPECT" | jq '.'
fi
echo ""

# Fetch Interface (Output Fields)
echo "=== Interface (Output Fields) ==="
INTERFACE=$(curl -s -H "Authorization: Token ${MAKE_API_KEY}" "${BASE_ENDPOINT}/interface")
if [[ -z "$INTERFACE" ]]; then
    echo "No interface configuration available"
else
    echo "$INTERFACE" | jq '.'
fi
echo ""

echo "=== Full Module Configuration ==="
echo "$RESPONSE" | jq '.'

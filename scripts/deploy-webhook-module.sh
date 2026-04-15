#!/bin/bash
#
# deploy-webhook-module.sh - Deploy watchTicketNotes module configuration to Make
#
# Usage: envwith -f .secrets/.env -- ./scripts/deploy-webhook-module.sh
#
# This script updates the watchTicketNotes module in the Make custom app with
# proper webhook lifecycle management configuration.
#

set -e

# Check for required environment variables
if [[ -z "$MAKE_API_KEY" ]]; then
    echo "Error: MAKE_API_KEY not set in environment"
    exit 1
fi

if [[ -z "$MAKE_API_URL" ]]; then
    echo "Error: MAKE_API_URL not set in environment"
    exit 1
fi

# Make API configuration
APP_ID="scj-autotask-nn8loi"
APP_VERSION="1"
MODULE_NAME="watchTicketNotes"

echo "Deploying webhook module configuration for: $MODULE_NAME"
echo ""

# Step 1: Update Expect (Input Fields)
echo "=== Updating Expect Configuration ==="
EXPECT_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X PUT \
    -H "Authorization: Token ${MAKE_API_KEY}" \
    -H "Content-Type: application/json" \
    -d @make-webhook-module/watchTicketNotes-expect.json \
    "${MAKE_API_URL}/sdk/apps/${APP_ID}/${APP_VERSION}/modules/${MODULE_NAME}/expect")

HTTP_CODE=$(echo "$EXPECT_RESPONSE" | tail -n1 | sed 's/HTTP_CODE://')
RESPONSE_BODY=$(echo "$EXPECT_RESPONSE" | sed '$ d')

if [[ "$HTTP_CODE" == "200" ]]; then
    echo "✅ Expect configuration updated successfully"
    echo "$RESPONSE_BODY" | jq '.'
else
    echo "❌ Failed to update expect configuration (HTTP $HTTP_CODE)"
    echo "$RESPONSE_BODY"
    exit 1
fi
echo ""

# Step 2: Update API (Webhook Lifecycle)
echo "=== Updating API Configuration ==="
API_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X PUT \
    -H "Authorization: Token ${MAKE_API_KEY}" \
    -H "Content-Type: application/json" \
    -d @make-webhook-module/watchTicketNotes-api.json \
    "${MAKE_API_URL}/sdk/apps/${APP_ID}/${APP_VERSION}/modules/${MODULE_NAME}/api")

HTTP_CODE=$(echo "$API_RESPONSE" | tail -n1 | sed 's/HTTP_CODE://')
RESPONSE_BODY=$(echo "$API_RESPONSE" | sed '$ d')

if [[ "$HTTP_CODE" == "200" ]]; then
    echo "✅ API configuration updated successfully"
    echo "$RESPONSE_BODY" | jq '.'
else
    echo "❌ Failed to update API configuration (HTTP $HTTP_CODE)"
    echo "$RESPONSE_BODY"
    exit 1
fi
echo ""

echo "🎉 Module deployment completed successfully!"
echo ""
echo "Next steps:"
echo "1. Test the module in a Make scenario"
echo "2. Verify webhook registration in Autotask"
echo "3. Test event triggering with ticket note creation"
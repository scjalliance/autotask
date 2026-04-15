#!/bin/bash
#
# fix-webhook-modules.sh - Set webhook references for all watch modules
#
# Usage: envwith -f .secrets/.env -- ./scripts/fix-webhook-modules.sh
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

APP_ID="scj-autotask-nn8loi"
APP_VERSION="1"

# List of watch modules that need webhook references
WATCH_MODULES=(
    "watchTickets"
    "watchCompanies"
    "watchContacts"
    "watchConfigurationItems"
    "watchTicketNotes"
)

echo "Setting webhook references for Autotask watch modules..."
echo ""

for MODULE in "${WATCH_MODULES[@]}"; do
    echo "=== Updating $MODULE ==="

    RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
        -X PATCH \
        -H "Authorization: Token ${MAKE_API_KEY}" \
        -H "Content-Type: application/json" \
        -d '{"webhook": "scj-autotask-nn8loi"}' \
        "${MAKE_API_URL}/sdk/apps/${APP_ID}/${APP_VERSION}/modules/${MODULE}")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1 | sed 's/HTTP_CODE://')
    RESPONSE_BODY=$(echo "$RESPONSE" | sed '$ d')

    if [[ "$HTTP_CODE" == "200" ]]; then
        WEBHOOK_REF=$(echo "$RESPONSE_BODY" | jq -r '.appModule.webhook // "null"')
        echo "✅ $MODULE webhook reference set to: $WEBHOOK_REF"
    else
        echo "❌ Failed to update $MODULE (HTTP $HTTP_CODE)"
        echo "$RESPONSE_BODY"
    fi
    echo ""
done

echo "🎉 Webhook reference updates completed!"
echo ""
echo "Next steps:"
echo "1. Re-save webhook modules in Make scenarios"
echo "2. Verify webhook registration in Autotask"
echo "3. Test with ticket note creation"
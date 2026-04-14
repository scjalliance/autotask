#!/bin/bash
#
# get-autotask-field-info.sh - Fetch Autotask field definitions and picklist values
#
# Usage: ./scripts/get-autotask-field-info.sh <entity_name> [field_name]
#
# This script fetches field information from the Autotask REST API.
# If field_name is provided, shows picklist values for that field.
# If field_name is not provided, shows all fields for the entity.
#
# Examples:
#   ./scripts/get-autotask-field-info.sh Ticket            # All Ticket fields
#   ./scripts/get-autotask-field-info.sh Ticket status     # Picklist for Ticket.status
#

set -e

# Check if entity name is provided
if [[ -z "$1" ]]; then
    echo "Error: Entity name required"
    echo "Usage: $0 <entity_name> [field_name]"
    echo "Examples:"
    echo "  $0 Ticket"
    echo "  $0 Ticket status"
    exit 1
fi

ENTITY_NAME="$1"
FIELD_NAME="$2"

# Load environment variables from .secrets/.env
if [[ ! -f .secrets/.env ]]; then
    echo "Error: .secrets/.env file not found"
    exit 1
fi

source .secrets/.env

# Check for required environment variables
if [[ -z "$AUTOTASK_USERNAME" ]]; then
    echo "Error: AUTOTASK_USERNAME not set in .secrets/.env"
    exit 1
fi

if [[ -z "$AUTOTASK_SECRET" ]]; then
    echo "Error: AUTOTASK_SECRET not set in .secrets/.env"
    exit 1
fi

if [[ -z "$AUTOTASK_INTEGRATION_CODE" ]]; then
    echo "Error: AUTOTASK_INTEGRATION_CODE not set in .secrets/.env"
    exit 1
fi

# Construct API credentials and URL
API_BASE_URL="https://webservices5.autotask.net/ATServicesRest/V1.0"

# Determine if we're fetching all fields or a specific field's picklist
if [[ -z "$FIELD_NAME" ]]; then
    # Fetch all fields for the entity
    ENDPOINT="${API_BASE_URL}/${ENTITY_NAME}s/entityInformation/fields"

    echo "Fetching field information for entity: $ENTITY_NAME"
    echo "Endpoint: $ENDPOINT"
    echo ""

    RESPONSE=$(curl -s \
        -H "ApiIntegrationCode: ${AUTOTASK_INTEGRATION_CODE}" \
        -H "UserName: ${AUTOTASK_USERNAME}" \
        -H "Secret: ${AUTOTASK_SECRET}" \
        "$ENDPOINT")

    # Check if curl succeeded and response is not empty
    if [[ -z "$RESPONSE" ]]; then
        echo "Error: curl failed or no data returned"
        exit 1
    fi

    # Check if response contains an error
    if echo "$RESPONSE" | jq -e '.pageDetails.errorDetail' > /dev/null 2>&1; then
        ERROR=$(echo "$RESPONSE" | jq -r '.pageDetails.errorDetail')
        echo "API Error: $ERROR"
        exit 1
    fi

    # Display field information
    echo "=== Fields for $ENTITY_NAME ==="
    echo "$RESPONSE" | jq '.fields | sort_by(.name) | .[] | {name, dataType, picklistValues}'

else
    # Fetch picklist values for a specific field
    ENDPOINT="${API_BASE_URL}/${ENTITY_NAME}s/entityInformation/fields"

    echo "Fetching picklist values for: $ENTITY_NAME.$FIELD_NAME"
    echo "Endpoint: $ENDPOINT"
    echo ""

    RESPONSE=$(curl -s \
        -H "ApiIntegrationCode: ${AUTOTASK_INTEGRATION_CODE}" \
        -H "UserName: ${AUTOTASK_USERNAME}" \
        -H "Secret: ${AUTOTASK_SECRET}" \
        "$ENDPOINT")

    # Check if curl succeeded and response is not empty
    if [[ -z "$RESPONSE" ]]; then
        echo "Error: curl failed or no data returned"
        exit 1
    fi

    # Check if response contains an error
    if echo "$RESPONSE" | jq -e '.pageDetails.errorDetail' > /dev/null 2>&1; then
        ERROR=$(echo "$RESPONSE" | jq -r '.pageDetails.errorDetail')
        echo "API Error: $ERROR"
        exit 1
    fi

    # Extract the specific field and display its picklist values
    FIELD_INFO=$(echo "$RESPONSE" | jq ".fields[] | select(.name == \"$FIELD_NAME\")")

    if [[ -z "$FIELD_INFO" ]]; then
        echo "Error: Field '$FIELD_NAME' not found for entity '$ENTITY_NAME'"
        exit 1
    fi

    echo "=== Field Information ==="
    echo "$FIELD_INFO" | jq '{name, dataType, description}'
    echo ""

    echo "=== Picklist Values ==="
    echo "$FIELD_INFO" | jq '.picklistValues'
fi

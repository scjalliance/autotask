#!/bin/bash
#
# test-module.sh - Test a Make module configuration and functionality
#
# Usage: envwith -f .secrets/.env -- ./scripts/test-module.sh <module_name> [test_type]
#
# Test types:
#   config    - Test module configuration (dropdowns, fields, defaults)
#   api       - Test API functionality with sample data
#   full      - Run both config and api tests (default)
#
# This script validates:
# - Field configurations load correctly
# - Dropdown options populate as expected
# - Default values are set appropriately
# - API endpoints respond correctly
# - Error handling works properly
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check arguments
MODULE_NAME="$1"
TEST_TYPE="${2:-full}"

if [[ -z "$MODULE_NAME" ]]; then
    echo -e "${RED}Error: Module name required${NC}"
    echo "Usage: envwith -f .secrets/.env -- $0 <module_name> [test_type]"
    echo "Example: envwith -f .secrets/.env -- $0 ticketsCreate config"
    echo ""
    echo "Test types:"
    echo "  config    - Test module configuration (dropdowns, fields, defaults)"
    echo "  api       - Test API functionality with sample data"
    echo "  full      - Run both config and api tests (default)"
    exit 1
fi

# Validate test type
case "$TEST_TYPE" in
    config|api|full)
        ;;
    *)
        echo -e "${RED}Error: Invalid test type '$TEST_TYPE'${NC}"
        echo "Valid options: config, api, full"
        exit 1
        ;;
esac

# Check for required environment variables
if [[ -z "$MAKE_API_KEY" ]]; then
    echo -e "${RED}Error: MAKE_API_KEY not set in environment${NC}"
    exit 1
fi

if [[ -z "$MAKE_API_URL" ]]; then
    echo -e "${RED}Error: MAKE_API_URL not set in environment${NC}"
    exit 1
fi

# Configuration
APP_ID="scj-autotask-nn8loi"
APP_VERSION="1"
BASE_ENDPOINT="${MAKE_API_URL}/sdk/apps/${APP_ID}/${APP_VERSION}/modules/${MODULE_NAME}"

# Test results
CONFIG_TESTS_PASSED=0
CONFIG_TESTS_FAILED=0
API_TESTS_PASSED=0
API_TESTS_FAILED=0

echo -e "${BLUE}Testing module: $MODULE_NAME${NC}"
echo -e "${BLUE}Test type: $TEST_TYPE${NC}"
echo ""

# Function to log test results
log_test_result() {
    local test_name="$1"
    local result="$2"
    local message="$3"

    if [[ "$result" == "PASS" ]]; then
        echo -e "  ${GREEN}✓${NC} $test_name: $message"
        if [[ "$test_name" == *"Config"* ]]; then
            ((CONFIG_TESTS_PASSED++))
        else
            ((API_TESTS_PASSED++))
        fi
    else
        echo -e "  ${RED}✗${NC} $test_name: $message"
        if [[ "$test_name" == *"Config"* ]]; then
            ((CONFIG_TESTS_FAILED++))
        else
            ((API_TESTS_FAILED++))
        fi
    fi
}

# Function to test module configuration
test_module_config() {
    echo -e "${YELLOW}=== Configuration Tests ===${NC}"

    # Test 1: Module exists and is accessible
    echo "Testing module accessibility..."
    MODULE_RESPONSE=$(curl -s -H "Authorization: Token ${MAKE_API_KEY}" "${BASE_ENDPOINT}")

    if [[ -z "$MODULE_RESPONSE" ]]; then
        log_test_result "Config - Accessibility" "FAIL" "No response from module endpoint"
        return
    fi

    if echo "$MODULE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        ERROR_MSG=$(echo "$MODULE_RESPONSE" | jq -r '.error.message // "Unknown error"')
        log_test_result "Config - Accessibility" "FAIL" "API error: $ERROR_MSG"
        return
    fi

    if ! echo "$MODULE_RESPONSE" | jq -e '.appModule.name' > /dev/null 2>&1; then
        log_test_result "Config - Accessibility" "FAIL" "Module not found or malformed response"
        return
    fi

    log_test_result "Config - Accessibility" "PASS" "Module found and accessible"

    # Test 2: Basic module info
    MODULE_NAME_ACTUAL=$(echo "$MODULE_RESPONSE" | jq -r '.appModule.name')
    MODULE_LABEL=$(echo "$MODULE_RESPONSE" | jq -r '.appModule.label')
    MODULE_TYPE=$(echo "$MODULE_RESPONSE" | jq -r '.appModule.typeId')

    # Debug: print values if needed
    # echo "DEBUG: Name=$MODULE_NAME_ACTUAL, Label=$MODULE_LABEL, Type=$MODULE_TYPE"

    if [[ "$MODULE_NAME_ACTUAL" == "$MODULE_NAME" ]]; then
        log_test_result "Config - Name Match" "PASS" "Name matches: $MODULE_NAME"
    else
        log_test_result "Config - Name Match" "FAIL" "Expected $MODULE_NAME, got $MODULE_NAME_ACTUAL"
    fi

    if [[ -n "$MODULE_LABEL" && "$MODULE_LABEL" != "null" ]]; then
        log_test_result "Config - Label" "PASS" "Label defined: '$MODULE_LABEL'"
    else
        log_test_result "Config - Label" "FAIL" "Missing or empty label"
    fi

    # Test 3: Expect (Input) fields
    echo ""
    echo "Testing input field configuration..."
    EXPECT_RESPONSE=$(curl -s -H "Authorization: Token ${MAKE_API_KEY}" "${BASE_ENDPOINT}/expect")

    if [[ -z "$EXPECT_RESPONSE" ]]; then
        log_test_result "Config - Input Fields" "FAIL" "No expect configuration"
        return
    fi

    # Debug API errors
    if echo "$EXPECT_RESPONSE" | jq -e '.error.message' > /dev/null 2>&1; then
        ERROR_MSG=$(echo "$EXPECT_RESPONSE" | jq -r '.error.message')
        log_test_result "Config - Input Fields" "FAIL" "API error: $ERROR_MSG"
        return
    fi

    if echo "$EXPECT_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        ERROR_MSG=$(echo "$EXPECT_RESPONSE" | jq -r '.error.message // "Unknown error"')
        log_test_result "Config - Input Fields" "FAIL" "Error fetching fields: $ERROR_MSG"
        return
    fi

    # Check if it's a valid array
    if ! echo "$EXPECT_RESPONSE" | jq -e 'type == "array"' > /dev/null 2>&1; then
        log_test_result "Config - Input Fields" "FAIL" "Expect configuration is not an array"
        return
    fi

    FIELD_COUNT=$(echo "$EXPECT_RESPONSE" | jq 'length')
    log_test_result "Config - Input Fields" "PASS" "$FIELD_COUNT fields defined"

    # Test 4: Check for dropdown fields with options
    echo ""
    echo "Testing dropdown field configurations..."
    SELECT_FIELDS=$(echo "$EXPECT_RESPONSE" | jq '[.[] | select(.type == "select")]')
    SELECT_COUNT=$(echo "$SELECT_FIELDS" | jq 'length')

    if [[ "$SELECT_COUNT" -gt 0 ]]; then
        log_test_result "Config - Dropdown Fields" "PASS" "Found $SELECT_COUNT dropdown fields"

        # Test each dropdown field for valid options
        for i in $(seq 0 $((SELECT_COUNT - 1))); do
            FIELD=$(echo "$SELECT_FIELDS" | jq ".[$i]")
            FIELD_NAME=$(echo "$FIELD" | jq -r '.name')
            OPTIONS_COUNT=$(echo "$FIELD" | jq '.options | length')

            if [[ "$OPTIONS_COUNT" -gt 0 ]]; then
                log_test_result "Config - Dropdown Options" "PASS" "$FIELD_NAME has $OPTIONS_COUNT options"

                # Check if options have proper structure
                FIRST_OPTION=$(echo "$FIELD" | jq '.options[0]')
                if echo "$FIRST_OPTION" | jq -e '.value and .label' > /dev/null 2>&1; then
                    log_test_result "Config - Option Structure" "PASS" "$FIELD_NAME options properly formatted"
                else
                    log_test_result "Config - Option Structure" "FAIL" "$FIELD_NAME options missing value/label"
                fi
            else
                log_test_result "Config - Dropdown Options" "FAIL" "$FIELD_NAME has no options"
            fi
        done
    else
        log_test_result "Config - Dropdown Fields" "INFO" "No dropdown fields found"
    fi

    # Test 5: Check for required fields
    echo ""
    echo "Testing required field configuration..."
    REQUIRED_FIELDS=$(echo "$EXPECT_RESPONSE" | jq '[.[] | select(.required == true)]')
    REQUIRED_COUNT=$(echo "$REQUIRED_FIELDS" | jq 'length')

    if [[ "$REQUIRED_COUNT" -gt 0 ]]; then
        log_test_result "Config - Required Fields" "PASS" "Found $REQUIRED_COUNT required fields"
    else
        log_test_result "Config - Required Fields" "INFO" "No required fields found"
    fi

    # Test 6: Interface (Output) configuration
    echo ""
    echo "Testing output field configuration..."
    INTERFACE_RESPONSE=$(curl -s -H "Authorization: Token ${MAKE_API_KEY}" "${BASE_ENDPOINT}/interface")

    if [[ -n "$INTERFACE_RESPONSE" ]] && ! echo "$INTERFACE_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        if echo "$INTERFACE_RESPONSE" | jq -e 'type == "array"' > /dev/null 2>&1; then
            OUTPUT_COUNT=$(echo "$INTERFACE_RESPONSE" | jq 'length')
            log_test_result "Config - Output Fields" "PASS" "$OUTPUT_COUNT output fields defined"
        else
            log_test_result "Config - Output Fields" "FAIL" "Interface configuration is not an array"
        fi
    else
        log_test_result "Config - Output Fields" "INFO" "No interface configuration or error"
    fi
}

# Function to test API functionality
test_module_api() {
    echo -e "${YELLOW}=== API Functionality Tests ===${NC}"

    # Test 1: API configuration exists
    echo "Testing API configuration..."
    API_RESPONSE=$(curl -s -H "Authorization: Token ${MAKE_API_KEY}" "${BASE_ENDPOINT}/api")

    if [[ -z "$API_RESPONSE" ]]; then
        log_test_result "API - Configuration" "FAIL" "No API configuration"
        return
    fi

    if echo "$API_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
        ERROR_MSG=$(echo "$API_RESPONSE" | jq -r '.error.message // "Unknown error"')
        log_test_result "API - Configuration" "FAIL" "Error fetching API config: $ERROR_MSG"
        return
    fi

    log_test_result "API - Configuration" "PASS" "API configuration accessible"

    # Test 2: API method and URL
    if echo "$API_RESPONSE" | jq -e '.method' > /dev/null 2>&1; then
        METHOD=$(echo "$API_RESPONSE" | jq -r '.method')
        log_test_result "API - HTTP Method" "PASS" "Method defined: $METHOD"
    else
        log_test_result "API - HTTP Method" "FAIL" "No HTTP method defined"
    fi

    if echo "$API_RESPONSE" | jq -e '.url' > /dev/null 2>&1; then
        URL=$(echo "$API_RESPONSE" | jq -r '.url')
        log_test_result "API - Endpoint URL" "PASS" "URL defined: $URL"
    else
        log_test_result "API - Endpoint URL" "FAIL" "No endpoint URL defined"
    fi

    # Test 3: Headers configuration
    if echo "$API_RESPONSE" | jq -e '.headers' > /dev/null 2>&1; then
        HEADERS_COUNT=$(echo "$API_RESPONSE" | jq '.headers | length')
        log_test_result "API - Headers" "PASS" "$HEADERS_COUNT headers configured"
    else
        log_test_result "API - Headers" "INFO" "No headers configuration"
    fi

    # Test 4: Response handling
    if echo "$API_RESPONSE" | jq -e '.response' > /dev/null 2>&1; then
        log_test_result "API - Response Config" "PASS" "Response handling configured"
    else
        log_test_result "API - Response Config" "INFO" "No response configuration"
    fi

    # Note: We're not doing live API calls here as they would require valid Autotask credentials
    # and could have side effects. Instead, we test the configuration completeness.
    echo ""
    echo -e "${YELLOW}Note: Live API testing skipped to avoid side effects${NC}"
    echo -e "${YELLOW}To test actual API calls, use the module in Make with test data${NC}"
}

# Function to print summary
print_summary() {
    echo ""
    echo -e "${BLUE}=== Test Summary ===${NC}"

    local total_passed=0
    local total_failed=0

    if [[ "$TEST_TYPE" == "config" || "$TEST_TYPE" == "full" ]]; then
        echo -e "Configuration Tests: ${GREEN}$CONFIG_TESTS_PASSED passed${NC}, ${RED}$CONFIG_TESTS_FAILED failed${NC}"
        ((total_passed += CONFIG_TESTS_PASSED))
        ((total_failed += CONFIG_TESTS_FAILED))
    fi

    if [[ "$TEST_TYPE" == "api" || "$TEST_TYPE" == "full" ]]; then
        echo -e "API Tests: ${GREEN}$API_TESTS_PASSED passed${NC}, ${RED}$API_TESTS_FAILED failed${NC}"
        ((total_passed += API_TESTS_PASSED))
        ((total_failed += API_TESTS_FAILED))
    fi

    echo ""
    echo -e "Total: ${GREEN}$total_passed passed${NC}, ${RED}$total_failed failed${NC}"

    if [[ $total_failed -eq 0 ]]; then
        echo -e "${GREEN}✓ All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}✗ Some tests failed${NC}"
        exit 1
    fi
}

# Run tests based on type
case "$TEST_TYPE" in
    config)
        test_module_config
        ;;
    api)
        test_module_api
        ;;
    full)
        test_module_config
        echo ""
        test_module_api
        ;;
esac

print_summary
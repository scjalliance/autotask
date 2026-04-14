# Module Testing and Validation Checklist

## Overview

This document provides a systematic approach to testing and validating Make module configurations and functionality after improvements. Use this checklist to ensure all module changes are properly tested before deployment.

## Quick Test Command

```bash
# Test any module configuration and API setup
envwith -f .secrets/.env -- ./scripts/test-module.sh <module_name> [config|api|full]

# Examples:
envwith -f .secrets/.env -- ./scripts/test-module.sh ticketsCreate config
envwith -f .secrets/.env -- ./scripts/test-module.sh companiesQuery full
```

## Pre-Testing Setup

### Environment Verification
- [ ] `.secrets/.env` file exists with required API keys
- [ ] `MAKE_API_KEY` is set and valid
- [ ] `MAKE_API_URL` is set correctly
- [ ] `envwith` tool is installed and accessible
- [ ] Testing scripts have execute permissions

### Module Identification
- [ ] Module name is confirmed (check inventory: `docs/make-app-modules-inventory.md`)
- [ ] Module type is known (Create/Query/Trigger/Special)
- [ ] Previous improvements are documented and ready to test

## Core Configuration Testing

### Module Accessibility
- [ ] Module responds to API calls
- [ ] No authentication errors
- [ ] Module name matches expected value
- [ ] Module label is defined and user-friendly
- [ ] Module type ID is correct

### Input Field Configuration
- [ ] Expect configuration loads without errors
- [ ] Field count matches expected number
- [ ] All field names are defined
- [ ] Field labels are user-friendly (not technical)
- [ ] Field types are appropriate for data

### Dropdown Field Validation
For modules with select/dropdown fields:
- [ ] All dropdown fields have `type: "select"`
- [ ] Each dropdown has options array
- [ ] Options have proper `value` and `label` structure
- [ ] Option values are valid (match API expectations)
- [ ] Option labels are user-friendly
- [ ] Default values are set appropriately
- [ ] Default values exist in options array

### Required Field Handling
- [ ] Required fields are marked `required: true`
- [ ] Required fields have clear labels
- [ ] Optional fields don't have unnecessary requirements
- [ ] Field validation rules are logical

### Help Text and Descriptions
- [ ] Complex fields have helpful descriptions
- [ ] Field labels explain purpose clearly
- [ ] Examples provided where helpful
- [ ] Technical terms are avoided or explained

## API Configuration Testing

### Endpoint Configuration
- [ ] HTTP method is defined and appropriate
- [ ] API endpoint URL is valid
- [ ] URL parameters are properly configured
- [ ] Headers are set correctly (Authorization, Content-Type)

### Request Handling
- [ ] Request body structure is valid
- [ ] Parameter mapping works correctly
- [ ] Optional parameters handled properly
- [ ] Required parameters validated

### Response Handling
- [ ] Response parsing is configured
- [ ] Output fields are mapped correctly
- [ ] Error handling is implemented
- [ ] Status codes handled appropriately

## Module-Specific Testing

### Create/Add Modules (typeId: 7)
**Examples: ticketsCreate, companiesCreate, contactsCreate**

#### Dropdown Field Testing
- [ ] Priority fields show user-friendly options (Critical, High, Medium, Low)
- [ ] Status fields show workflow-appropriate options
- [ ] Type fields show all available categories
- [ ] Reference ID fields use dropdowns instead of number inputs

#### Field Organization
- [ ] Related fields are grouped logically
- [ ] Required fields are prominent
- [ ] Optional fields don't clutter the interface
- [ ] Field order follows logical workflow

#### Default Values
- [ ] Sensible defaults are set based on usage patterns
- [ ] Default values match Autotask API expectations
- [ ] Defaults improve workflow efficiency

### Query/Search Modules (typeId: 9)
**Examples: ticketsQuery, companiesQuery, timeEntriesQuery**

#### Search Field Configuration
- [ ] Search field label is entity-specific ("Ticket Filters" not "Search")
- [ ] Filter options are comprehensive but organized
- [ ] Field labels are clear (not technical IDs)
- [ ] Common filters are easy to find

#### Field Organization
- [ ] Related fields grouped with consistent prefixes
- [ ] Date fields grouped together ("Date - Created", "Date - Modified")
- [ ] Address fields grouped ("Address - Street 1", "Address - City")
- [ ] Status and type fields grouped logically

#### Performance Settings
- [ ] Default limit is appropriate for entity type
- [ ] Pagination is configured correctly
- [ ] Large result sets are handled properly

### Trigger/Watcher Modules (typeId: 10)
**Examples: ticketsWatch, companiesWatch**

#### Webhook Configuration
- [ ] Webhook endpoint is properly configured
- [ ] Event types are mapped correctly
- [ ] Filter criteria work as expected
- [ ] Output format matches requirements

#### Real-time Testing
- [ ] Webhook receives events correctly
- [ ] Event data is parsed properly
- [ ] Duplicate events are handled
- [ ] Error conditions trigger appropriate responses

### Special Operations (typeId: 1-3, 6, 8, 11)
**Examples: authenticate, listPicklistValues**

#### Authentication Modules
- [ ] Credentials are validated correctly
- [ ] Connection status is reported accurately
- [ ] Error messages are helpful
- [ ] Security requirements are met

#### Utility Modules
- [ ] Data transformation works correctly
- [ ] Reference data is up-to-date
- [ ] Performance is acceptable
- [ ] Output format is consistent

## Manual Testing in Make Interface

### Module Loading
- [ ] Module appears in correct category
- [ ] Module icon and description are appropriate
- [ ] Module loads without errors
- [ ] All configured fields appear

### Field Interaction
- [ ] Dropdown menus populate correctly
- [ ] Default values appear automatically
- [ ] Required field validation works
- [ ] Help text displays when available

### Data Flow Testing
- [ ] Test with minimal required data
- [ ] Test with complex data scenarios
- [ ] Verify output format is as expected
- [ ] Check error handling with invalid data

### Integration Testing
- [ ] Module works in typical scenarios
- [ ] Connects properly with other modules
- [ ] Performance is acceptable
- [ ] No unexpected side effects

## Common Issues and Solutions

### Dropdown Fields Not Loading
**Symptoms**: Dropdown shows as number field or empty options
- Check options array structure (value/label pairs)
- Verify field type is set to "select"
- Confirm options values match API expectations
- Test with audit script: `./scripts/audit-make-module.sh <module>`

### API Connection Errors
**Symptoms**: Module fails to load or returns authentication errors
- Verify MAKE_API_KEY is valid
- Check API endpoint configuration
- Confirm module exists in correct app version
- Test with curl: `curl -H "Authorization: Token $MAKE_API_KEY" "$MAKE_API_URL/sdk/apps/scj-autotask-nn8loi/1/modules/<module>"`

### Field Validation Issues
**Symptoms**: Required fields not enforced, unexpected validation errors
- Review field definitions in expect configuration
- Check required flag is set correctly
- Verify validation rules are logical
- Test edge cases (empty values, special characters)

### Performance Problems
**Symptoms**: Slow loading, timeouts, excessive data
- Adjust default limit values
- Optimize field organization
- Review API endpoint efficiency
- Consider pagination for large datasets

## Automated Testing Results

### Test Script Results
Document results from `./scripts/test-module.sh`:

#### Configuration Tests
- Module Accessibility: ✓/✗
- Name Match: ✓/✗
- Label: ✓/✗
- Input Fields: ✓/✗
- Dropdown Fields: ✓/✗
- Dropdown Options: ✓/✗
- Option Structure: ✓/✗
- Required Fields: ✓/✗
- Output Fields: ✓/✗

#### API Tests
- Configuration: ✓/✗
- HTTP Method: ✓/✗
- Endpoint URL: ✓/✗
- Headers: ✓/✗
- Response Config: ✓/✗

## Testing Schedule and Priorities

### Immediate Testing (Post-Implementation)
1. **ticketsCreate** - Already fixed, needs validation
2. **Query modules** - Designed improvements ready for testing
3. **New dropdown conversions** - Any recently modified modules

### Regular Testing (Weekly/Monthly)
1. **High-usage modules** - ticketsCreate, ticketsQuery, companiesQuery
2. **Recently modified modules** - Any modules with recent changes
3. **Regression testing** - Previously fixed modules

### Full Suite Testing (Quarterly)
1. **All 117 modules** - Complete inventory validation
2. **Performance benchmarks** - Response times and reliability
3. **User experience audit** - Field organization and clarity

## Documentation Requirements

### Test Results Documentation
For each tested module, document:
- Test date and tester
- Test script results (pass/fail counts)
- Manual testing notes
- Issues found and resolved
- Performance observations
- User experience feedback

### Issue Tracking
- **Critical issues**: API failures, authentication problems
- **Major issues**: Broken dropdowns, missing required fields
- **Minor issues**: Unclear labels, suboptimal defaults
- **Enhancements**: Improvement opportunities identified

### Success Criteria
A module passes testing when:
- All automated tests pass
- Dropdowns load with proper options
- Required fields are clearly marked
- API endpoints respond correctly
- Manual testing confirms good user experience
- No critical or major issues remain

## Integration with Development Workflow

### Pre-deployment Testing
1. Run automated tests for all modified modules
2. Perform manual testing for key scenarios
3. Document any issues and resolutions
4. Get approval for production deployment

### Post-deployment Validation
1. Re-run automated tests in production
2. Monitor for user-reported issues
3. Verify performance metrics
4. Update documentation as needed

### Continuous Improvement
1. Collect user feedback on module usability
2. Identify patterns in common issues
3. Enhance testing scripts and procedures
4. Refine improvement patterns and standards

---

**Usage**: Use this checklist systematically for all module testing. Combine automated testing with manual validation for comprehensive coverage. Document all results to track improvement patterns and catch regressions early.
# Autotask Make App - User Guide and Reference

> **App Version**: `scj-autotask-nn8loi` v1
> **Last Updated**: April 14, 2026
> **Total Modules**: 117 (60 public, 57 private)

## Table of Contents

1. [Quick Start](#quick-start)
2. [Module Types Overview](#module-types-overview)
3. [Common Patterns and Best Practices](#common-patterns-and-best-practices)
4. [Module Categories](#module-categories)
5. [Field Value Reference](#field-value-reference)
6. [Troubleshooting](#troubleshooting)
7. [Advanced Configuration](#advanced-configuration)
8. [Testing and Validation](#testing-and-validation)

## Quick Start

### Essential Setup Steps

1. **Connect to Autotask**: Use the Authentication module first to establish connection
2. **Start with Common Modules**: Begin with `ticketsCreate`, `ticketsQuery`, `companiesQuery`
3. **Use Improved UX**: Look for dropdown fields instead of number inputs
4. **Check Required Fields**: Red asterisks indicate required fields
5. **Review Output**: Use test runs to understand data structure

### Most Used Modules
- **`ticketsCreate`** - Create new support tickets (✅ Enhanced UX)
- **`ticketsQuery`** - Search for tickets with filters (✅ Enhanced UX)
- **`companiesQuery`** - Search for companies (✅ Enhanced UX)
- **`contactsCreate`** - Create new contacts
- **`timeEntriesQuery`** - Search time entries (✅ Enhanced UX)
- **`watchTickets`** - Real-time ticket change notifications

## Module Types Overview

### Actions (69 modules)
**Purpose**: Create, update, or modify Autotask data
**Pattern**: `entityCreate`, `entityUpdate`
**Examples**: `ticketsCreate`, `companiesUpdate`, `taskNoteCreate`

**Key Features**:
- Dropdown fields for picklist values (Critical, High, Medium, Low)
- Sensible default values
- Required field validation
- Related entity selection

### Search/Query (42 modules)
**Purpose**: Find and retrieve Autotask data
**Pattern**: `entityQuery`, `entityGet`
**Examples**: `ticketsQuery`, `companiesQuery`, `timeEntriesQuery`

**Key Features**:
- Entity-specific filter labels ("Ticket Filters", not "Search")
- Organized field groups (Date -, Address -, Billing -)
- Optimized pagination (25-100 records)
- Multiple filter operators (equals, contains, in, exists)

### Instant Triggers (6 modules)
**Purpose**: Real-time notifications via webhooks
**Pattern**: `watchEntity`, `entityEvents`
**Examples**: `watchTickets`, `watchCompanies`, `ticketEvents`

**Key Features**:
- Automatic webhook registration
- Event filtering (create, update, delete)
- HMAC signature validation
- Structured output format

## Common Patterns and Best Practices

### Creating Records

#### ✅ Best Practice: Use Enhanced Dropdown Fields
```
Priority: [Medium ▼] (dropdown)
Status: [New ▼] (dropdown)
Ticket Type: [Service Request ▼] (dropdown)
Source: [Phone ▼] (dropdown)
```

#### ❌ Avoid: Manual Number Entry
```
Priority: [2] (number input - confusing!)
Status: [1] (number input - error-prone)
```

### Searching and Filtering

#### ✅ Best Practice: Use Logical Field Groups
```
Ticket Filters:
├── Date - Created (between dates)
├── Date - Due (before/after)
├── Status - Current (in list)
├── Priority - Level (equals)
├── Company - Name (contains)
└── Assigned - Resource (equals)
```

#### ✅ Best Practice: Common Filter Patterns
- **Recent items**: Date - Created > 7 days ago
- **Open items**: Status in [New, In Progress, Waiting Customer]
- **High priority**: Priority in [Critical, High]
- **My items**: Assigned Resource = {your resource ID}

### Working with Webhooks

#### ✅ Best Practice: Start with Basic Watchers
1. Use `watchTickets` for ticket change notifications
2. Filter events by action type (create/update/delete)
3. Process webhook data systematically
4. Implement retry logic for failed processing

#### ✅ Best Practice: Webhook Security
- Always validate HMAC signatures
- Limit webhook payload size (1MB max)
- Use HTTPS endpoints only
- Implement proper error handling

## Module Categories

### Core Business Entities

#### Tickets (3 modules)
- **`ticketsCreate`** ✅ - Enhanced with 4 dropdown fields
- **`ticketsQuery`** ✅ - Enhanced with organized filters
- **`ticketsUpdate`** - Standard update functionality

**Common Use Cases**:
- Create customer service requests
- Search for tickets by status, priority, company
- Update ticket information and status

**Key Fields** (ticketsCreate):
- **Priority**: Critical (4), High (1), Medium (2), Low (3), Information (5)
- **Status**: New (1), In Progress (8), Complete (5), Waiting Customer (9), etc.
- **Type**: Service Request (1), Incident (2), Problem (3), Change Request (4), Alert (5)
- **Source**: Phone (2), Email (4), Client Portal (-1), Walk-in (1), etc.

#### Companies (3 modules)
- **`companiesCreate`** - Create new client companies
- **`companiesQuery`** ✅ - Enhanced with grouped address/billing filters
- **`companiesUpdate`** - Update company information

**Key Field Groups** (companiesQuery):
- **Address Fields**: Address - Street 1, Address - City, Address - State
- **Billing Fields**: Billing - Address Street 1, Billing - Attention
- **Portal Fields**: Portal - Client Portal Active, Portal - Login
- **Tax Fields**: Tax - Exempt, Tax - Region

#### Contacts (3 modules)
- **`contactsCreate`** - Create new contact records
- **`contactsQuery`** - Search for contacts with filters
- **`contactsUpdate`** - Update contact information

### Project Management

#### Projects (3 modules)
- **`projectsCreate`** - Create new projects
- **`projectsQuery`** - Search for projects
- **`projectsUpdate`** - Update project details

#### Tasks (3 modules)
- **`tasksCreate`** - Create project and ticket tasks
- **`tasksQuery`** - Search for tasks
- **`tasksUpdate`** - Update task information

### Time and Billing

#### Time Entries (3 modules)
- **`timeEntriesCreate`** - Log time entries
- **`timeEntriesQuery`** ✅ - Enhanced with date/billing field groups
- **`timeEntriesUpdate`** - Update time records

**Key Field Groups** (timeEntriesQuery):
- **Date Fields**: Date - Worked, Date - Start Time, Date - End Time
- **Billing Fields**: Billing - Code, Billing - Approval Date, Billing - Rate
- **Resource Fields**: Resource, Role, Ticket, Task

### Notes and Attachments

#### Note Creation (5 modules)
- **`ticketNoteCreate`** - Add notes to tickets
- **`taskNoteCreate`** - Add notes to tasks
- **`companiesNoteCreate`** - Add notes to companies
- **`projectNoteCreate`** - Add notes to projects
- **`configurationItemNoteCreate`** - Add notes to assets

#### Attachment Queries (7 modules)
- **Pattern**: `entityAttachmentsQuery`
- **Entities**: Tickets, Companies, Projects, Tasks, Configuration Items, Opportunities, Time Entries
- **Filters**: Date ranges, file types, uploader

### Special Operations

#### Association Operations
- **`contactGroupContactAdd`** - Add contacts to groups
- **`contractServiceAdd`** - Add services to contracts
- **`ticketChecklistItemCreate`** - Add checklist items

#### Advanced Configuration
- **`ticketEvents`** - Custom webhook field configuration
- **`contractServiceAdjust`** - Adjust service quantities

## Field Value Reference

### Priority Values (All Ticket Modules)
| Display Name | API Value | Usage |
|-------------|-----------|--------|
| Critical | 4 | Emergency issues requiring immediate attention |
| High | 1 | Important issues affecting business operations |
| Medium | 2 | Standard requests and issues (default) |
| Low | 3 | Minor issues that can wait |
| Information | 5 | FYI items, no action required |

### Ticket Status Values
| Display Name | API Value | Workflow Stage |
|-------------|-----------|----------------|
| New | 1 | Just created (default) |
| In Progress | 8 | Being worked on |
| Complete | 5 | Work finished |
| Waiting Customer | 9 | Pending customer response |
| Waiting Vendor | 10 | Pending vendor response |
| Change Approved | 6 | Change management approved |
| Change Request | 7 | Awaiting change approval |

### Ticket Types
| Display Name | API Value | Description |
|-------------|-----------|-------------|
| Service Request | 1 | Standard service requests (default) |
| Incident | 2 | Unplanned service disruptions |
| Problem | 3 | Root cause investigation |
| Change Request | 4 | Planned changes to services |
| Alert | 5 | Automated monitoring alerts |

### Ticket Sources
| Display Name | API Value | Channel |
|-------------|-----------|---------|
| Phone | 2 | Phone call (default) |
| Email | 4 | Email submission |
| Client Portal | -1 | Customer portal |
| Walk-in | 1 | In-person visit |
| Monitoring Alert | 3 | Automated monitoring |

### Common Boolean Values
| Display Name | API Value | Usage |
|-------------|-----------|--------|
| Yes | true | Active, enabled, visible |
| No | false | Inactive, disabled, hidden |

### Date Format Guidelines
- **Input Format**: ISO 8601 (YYYY-MM-DDTHH:MM:SSZ)
- **Display Format**: Local timezone formatting
- **Common Filters**: "last 7 days", "this month", "custom range"

## Troubleshooting

### Common Issues and Solutions

#### Dropdown Not Loading Options
**Symptoms**: Field shows as number input or empty dropdown

**Solutions**:
1. **Check Module Version**: Ensure using updated module with enhanced UX
2. **Verify Connection**: Confirm Autotask connection is active
3. **Refresh Browser**: Clear cache and reload Make interface
4. **Check Permissions**: Verify account has access to required picklists

#### "Invalid Value" Errors
**Symptoms**: Form submission fails with invalid field values

**Solutions**:
1. **Use Dropdown Values**: Select from dropdown instead of typing numbers
2. **Check Required Fields**: Ensure all required fields are filled
3. **Verify References**: Ensure referenced IDs (Company, Contact) exist
4. **Review Field Constraints**: Some fields have specific validation rules

#### Webhook Not Triggering
**Symptoms**: No events received from webhook modules

**Solutions**:
1. **Verify Webhook URL**: Ensure endpoint is accessible and correct
2. **Check HMAC Validation**: Implement proper signature validation
3. **Review Event Filters**: Confirm watching for correct action types
4. **Test Endpoint**: Use webhook testing tools to verify connectivity

#### Query Returns No Results
**Symptoms**: Search modules return empty results unexpectedly

**Solutions**:
1. **Simplify Filters**: Start with basic filters and add complexity
2. **Check Field Names**: Use enhanced field labels instead of technical names
3. **Verify Permissions**: Ensure account can access requested data
4. **Review Date Ranges**: Date filters may be too restrictive

#### Performance Issues
**Symptoms**: Slow loading or timeouts

**Solutions**:
1. **Reduce Result Limits**: Use default optimized limits (25-100)
2. **Optimize Filters**: Use specific filters to reduce data volume
3. **Check Pagination**: Implement proper pagination for large datasets
4. **Review Module Choice**: Use Get modules for single records

### Error Code Reference

#### Authentication Errors (401)
- **Cause**: Invalid or expired API credentials
- **Solution**: Reconnect to Autotask, verify credentials

#### Permission Errors (403)
- **Cause**: Account lacks permissions for operation
- **Solution**: Check Autotask security roles and permissions

#### Not Found Errors (404)
- **Cause**: Referenced entity doesn't exist
- **Solution**: Verify entity IDs, check for deleted records

#### Validation Errors (400)
- **Cause**: Invalid field values or missing required fields
- **Solution**: Review field requirements, use dropdown values

#### Rate Limit Errors (429)
- **Cause**: Too many API requests in short time
- **Solution**: Implement delays, batch operations

### Getting Help

1. **Module Documentation**: Check individual module descriptions in Make
2. **Autotask API Docs**: Reference official Autotask REST API documentation
3. **Test Modules**: Use Make's test functionality to debug issues
4. **Support Channels**: Contact system administrator or Autotask support

## Advanced Configuration

### Custom Webhook Configuration

#### Using ticketEvents Module
The `ticketEvents` module provides advanced webhook customization:

**Field Configuration Array**:
```json
[
  {
    "FieldID": 123,           // Autotask field ID
    "DisplayAlways": true,    // Include in all payloads
    "TriggerOnUpdate": false  // Only trigger when this field changes
  }
]
```

**Common Field IDs**:
- Status: 25
- Priority: 164
- Assigned Resource: 69
- Due Date: 76
- Description: 80

#### Webhook Payload Structure
```json
{
  "action": "update",
  "id": 12345,
  "entityType": "Ticket",
  "eventTime": "2026-04-14T10:30:00Z",
  "sequenceNumber": 67890,
  "personId": 123,
  "guid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "fields": {
    "Status": {"id": 25, "value": "8"},
    "Priority": {"id": 164, "value": "1"}
  }
}
```

### Query Filter Operators

#### Text Field Operators
- **equals** (`eq`): Exact match
- **contains** (`contains`): Substring search
- **begins with** (`beginswith`): Prefix match
- **not equal** (`ne`): Exclusion filter

#### Number Field Operators
- **equals** (`eq`): Exact value match
- **greater than** (`gt`): Value greater than
- **less than** (`lt`): Value less than
- **between** (`gte` and `lte`): Range filter

#### Date Field Operators
- **on** (`eq`): Specific date
- **before** (`lt`): Earlier than date
- **after** (`gt`): Later than date
- **between** (`gte` and `lte`): Date range

#### Array Operators
- **in** (`in`): Value in list
- **not in** (`notin`): Value not in list

#### Existence Operators
- **exists** (`exist`): Field has any value
- **does not exist** (`notexist`): Field is empty/null

### Performance Optimization

#### Query Optimization
1. **Use Specific Filters**: Narrow results with precise criteria
2. **Limit Field Selection**: Only request needed fields
3. **Implement Pagination**: Process results in manageable chunks
4. **Cache Reference Data**: Store frequently used lookups

#### Webhook Optimization
1. **Filter Events**: Only watch for relevant changes
2. **Batch Processing**: Group related events together
3. **Async Processing**: Handle webhooks asynchronously
4. **Error Recovery**: Implement retry logic for failed processing

## Testing and Validation

### Module Testing Tools

The app includes comprehensive testing tools for validation:

#### Quick Module Test
```bash
# Test module configuration and API functionality
envwith -f .secrets/.env -- ./scripts/test-module.sh <module_name> [test_type]

# Examples:
envwith -f .secrets/.env -- ./scripts/test-module.sh ticketsCreate config
envwith -f .secrets/.env -- ./scripts/test-module.sh companiesQuery full
```

#### Module Audit
```bash
# Get complete module configuration
envwith -f .secrets/.env -- ./scripts/audit-make-module.sh <module_name>
```

#### Field Information Lookup
```bash
# Get Autotask field definitions and picklist values
envwith -f .secrets/.env -- ./scripts/get-autotask-field-info.sh <Entity> <fieldName>

# Example:
envwith -f .secrets/.env -- ./scripts/get-autotask-field-info.sh Ticket priority
```

### Testing Checklist

#### Before Using New Modules
- [ ] Run configuration test to verify dropdown fields
- [ ] Check default values are appropriate
- [ ] Verify required fields are clearly marked
- [ ] Test with sample data in Make interface

#### Before Production Deployment
- [ ] Test common use case scenarios
- [ ] Verify error handling with invalid data
- [ ] Check performance with realistic data volumes
- [ ] Validate webhook endpoints if using triggers

#### Regular Maintenance
- [ ] Monthly regression testing of high-usage modules
- [ ] Quarterly review of all 117 modules
- [ ] Performance monitoring and optimization
- [ ] User feedback collection and analysis

### Quality Assurance

#### Module Standards
All enhanced modules follow these quality standards:
- ✅ User-friendly dropdown fields instead of number inputs
- ✅ Logical field organization and grouping
- ✅ Appropriate default values for common scenarios
- ✅ Clear field labels without technical jargon
- ✅ Helpful descriptions for complex fields
- ✅ Optimized pagination and performance settings

#### Success Metrics
- **Field Configuration**: 100% of picklist fields use dropdowns
- **User Experience**: 90%+ of users can complete tasks without documentation
- **Error Reduction**: 75%+ reduction in validation errors
- **Performance**: 95%+ of queries complete under 10 seconds
- **Adoption**: 80%+ of workflows use enhanced modules

---

## Quick Reference Card

### Essential Modules
| Purpose | Module | Status | Notes |
|---------|--------|--------|--------|
| Create Tickets | `ticketsCreate` | ✅ Enhanced | 4 dropdown fields |
| Search Tickets | `ticketsQuery` | ✅ Enhanced | Organized filters |
| Search Companies | `companiesQuery` | ✅ Enhanced | Grouped fields |
| Search Time Entries | `timeEntriesQuery` | ✅ Enhanced | Date/billing groups |
| Watch Ticket Changes | `watchTickets` | ✅ Standard | Real-time webhooks |

### Priority Values Quick Reference
| Critical | High | Medium | Low | Info |
|----------|------|--------|-----|------|
| 4 | 1 | 2 | 3 | 5 |

### Status Values Quick Reference
| New | In Progress | Complete | Waiting Customer | Waiting Vendor |
|-----|-------------|----------|------------------|-----------------|
| 1 | 8 | 5 | 9 | 10 |

### Support Resources
- **Module Inventory**: `/docs/make-app-modules-inventory.md`
- **Testing Guide**: `/docs/module-testing-checklist.md`
- **Implementation Notes**: `/docs/module-review-*.md`
- **Testing Scripts**: `/scripts/test-module.sh`, `/scripts/audit-make-module.sh`

---

*This guide consolidates learnings from the systematic review of all 117 modules in the Autotask Make custom app. For specific module details, use the testing scripts or refer to the detailed review documentation.*
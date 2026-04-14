# Autotask Make App: Trigger and Watcher Module Review

## Executive Summary

Reviewed all 6 trigger/watcher modules (typeId=10) in the Autotask Make custom app. Found two distinct patterns: basic webhook watchers with consistent output structure and a more advanced configurable trigger module. Key improvement opportunities identified around webhook field configuration, action filtering, and user experience for webhook setup.

## Trigger Module Inventory

| Module Name | Description | Input Fields | Pattern |
|-------------|-------------|--------------|---------|
| `watchTickets` | Triggers on ticket create/update/delete | None | Standard watcher |
| `watchCompanies` | Triggers on company create/update/delete | None | Standard watcher |
| `watchContacts` | Triggers on contact create/update/delete | None | Standard watcher |
| `watchConfigurationItems` | Triggers on asset create/update/delete | None | Standard watcher |
| `watchTicketNotes` | Triggers on ticket note create/update/delete | None | Standard watcher |
| `ticketEvents` | Advanced ticket event trigger with field config | `webhookFields[]` array | Configurable trigger |

## Current Architecture Analysis

### Standard Watcher Pattern (5 modules)

All basic watcher modules (`watch*`) share identical structure:

**Input Configuration**:
- No input fields required - automatic webhook registration
- No filtering or customization options

**Output Fields** (consistent across all):
```json
{
  "action": "text",           // create/update/delete
  "id": "number",            // Entity ID
  "entityType": "text",      // Autotask entity type
  "eventTime": "date",       // When event occurred
  "sequenceNumber": "number", // Event sequence
  "personId": "number",      // User who triggered
  "guid": "text",           // Event GUID
  "fields": "collection"     // Entity field data
}
```

### Advanced Configurable Pattern (1 module)

The `ticketEvents` module provides webhook field customization:

**Input Configuration**:
```json
{
  "webhookFields": [
    {
      "FieldID": "number",         // Autotask field ID
      "DisplayAlways": "boolean",  // Include in payload
      "TriggerOnUpdate": "boolean" // Trigger when field changes
    }
  ]
}
```

**Output Fields**: Empty (likely populated dynamically based on configuration)

## User Experience Issues Identified

### 1. Webhook Setup Clarity
- **Issue**: No visible webhook URL or setup instructions in modules
- **Impact**: Users must figure out webhook configuration elsewhere
- **Recommendation**: Add help text explaining webhook endpoint setup

### 2. Action Filtering
- **Issue**: No ability to filter by action type (create/update/delete)
- **Impact**: Users receive all events, must filter in workflow
- **Recommendation**: Add action filter checkboxes

### 3. Field Selection Complexity
- **Issue**: `ticketEvents` requires knowledge of Autotask Field IDs
- **Impact**: Steep learning curve for field configuration
- **Recommendation**: Add field ID lookup or dropdown

### 4. Inconsistent Patterns
- **Issue**: `ticketEvents` works differently from all other watchers
- **Impact**: Confusing user experience
- **Recommendation**: Standardize approach across modules

### 5. Output Field Documentation
- **Issue**: Minimal labels, no descriptions for output fields
- **Impact**: Users unsure what each field contains
- **Recommendation**: Add field descriptions and examples

## Recommended Improvements

### Phase 1: Basic UX Enhancements

1. **Add Action Filtering to All Watchers**
   ```json
   {
     "name": "actions",
     "type": "multiselect",
     "label": "Watch For Actions",
     "help": "Select which actions should trigger this webhook",
     "options": [
       {"label": "Created", "value": "create"},
       {"label": "Updated", "value": "update"},
       {"label": "Deleted", "value": "delete"}
     ],
     "default": ["create", "update", "delete"]
   }
   ```

2. **Enhance Field Descriptions**
   Add help text to all output fields explaining their purpose and typical values.

3. **Add Webhook Setup Help**
   Add help text explaining webhook URL configuration and HMAC validation.

### Phase 2: Advanced Field Configuration

1. **Standardize Field Selection**
   Extend the `ticketEvents` field configuration pattern to other watchers:
   ```json
   {
     "name": "includeFields",
     "type": "array",
     "label": "Additional Fields",
     "help": "Select additional Autotask fields to include in webhook payload",
     "spec": [
       {
         "name": "fieldName",
         "type": "select",
         "label": "Field",
         "options": "{{getAutotaskFields(entityType)}}"
       },
       {
         "name": "triggerOnUpdate",
         "type": "boolean",
         "label": "Trigger on Change",
         "default": false
       }
     ]
   }
   ```

2. **Field ID Resolution**
   Implement dynamic field lookup to avoid requiring users to know Field IDs.

### Phase 3: Polling Fallback

1. **Add Polling Mode**
   For scenarios where webhooks aren't feasible:
   ```json
   {
     "name": "mode",
     "type": "select",
     "label": "Trigger Mode",
     "options": [
       {"label": "Webhook (Recommended)", "value": "webhook"},
       {"label": "Polling", "value": "polling"}
     ],
     "default": "webhook"
   },
   {
     "name": "pollingInterval",
     "type": "number",
     "label": "Polling Interval (minutes)",
     "help": "How often to check for new events",
     "default": 5,
     "required": false
   }
   ```

## Webhook Implementation Notes

Based on the Go SDK webhook implementation (`webhook.go`):

1. **HMAC Validation**: Implements HMAC-SHA1 signature validation with `X-Hook-Signature` header
2. **Entity Mapping**: Maps webhook EntityType to Go structs:
   - Ticket → Ticket
   - Account → Company
   - Contact → Contact
   - InstalledProduct → ConfigurationItem
   - TicketNote → TicketNote

3. **Field Parsing**: Webhook Fields are parsed into typed entities with proper JSON unmarshaling

4. **Security**: 1MB body limit, signature validation, structured error responses

## Testing Recommendations

### Webhook Testing Checklist

1. **Registration Testing**
   - [ ] Verify webhook URL accepts POST requests
   - [ ] Confirm HMAC signature validation works
   - [ ] Test webhook deregistration on module deletion

2. **Event Filtering**
   - [ ] Create entity - verify webhook triggered
   - [ ] Update entity - verify webhook triggered
   - [ ] Delete entity - verify webhook triggered
   - [ ] Verify action filtering works correctly

3. **Payload Validation**
   - [ ] Confirm all output fields populated correctly
   - [ ] Test field data types (dates, numbers, text)
   - [ ] Verify entity-specific ID field naming

4. **Error Handling**
   - [ ] Test invalid HMAC signatures
   - [ ] Test malformed payloads
   - [ ] Verify retry logic for failed webhooks

### Field Configuration Testing

For `ticketEvents` module specifically:

1. **Field ID Validation**
   - [ ] Test valid Autotask field IDs
   - [ ] Test invalid field IDs (error handling)
   - [ ] Verify DisplayAlways flag behavior

2. **Trigger Configuration**
   - [ ] Test TriggerOnUpdate flag
   - [ ] Verify field-specific triggering
   - [ ] Test multiple field configurations

## Conclusion

The trigger modules provide solid webhook foundation but need UX improvements for easier configuration. The biggest opportunity is standardizing the user experience across all watchers while providing appropriate customization options. Focus areas:

1. **Immediate**: Add action filtering and better field documentation
2. **Short-term**: Improve webhook setup guidance and field selection UX
3. **Long-term**: Consider polling fallback and advanced field configuration

The webhook implementation in the Go SDK is robust and provides good security foundations for the Make app integration.
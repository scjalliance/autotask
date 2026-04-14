# Module Review: Special Operations and Edge Cases

## Overview

Task 6 of the systematic module review focuses on special operations that don't follow standard CRUD patterns. These modules require custom UX improvements tailored to their unique workflows and requirements.

## Special Operation Categories

### 1. Adjustment Operations
- **`contractServiceAdjust`**: Adjust quantity of contract service units

**Current Issues:**
- Limited field set (only ID, descriptions, and unit cost)
- Missing quantity adjustment field (the core purpose!)
- No validation for adjustment constraints
- Unclear relationship between fields

**Proposed Improvements:**
- Add quantity field as primary input
- Add current quantity display for context
- Add adjustment reason/notes field
- Implement validation for quantity limits
- Show cost impact calculation
- Add confirmation dialog for significant changes

### 2. Association/Linking Operations
Complex operations that create relationships between entities:

- **`contactGroupContactAdd`**: Add a contact to a contact group
- **`contractServiceAdd`**: Add a service to a recurring contract
- **`quoteItemCreate`**: Add a line item to a quote
- **`ticketChecklistItemCreate`**: Add a checklist item to a ticket

**Current Issues:**
- Raw ID inputs without entity selection
- No validation for duplicate associations
- Missing context about target entities
- Complex field relationships not guided

**Proposed Improvements:**
- Convert ID fields to searchable dropdowns
- Add entity preview/validation
- Implement duplicate checking
- Add bulk operations for multiple items
- Provide templates for common configurations

### 3. Line Item Operations
Specialized modules for adding items with complex pricing:

- **`quoteItemCreate`**: 21 fields including multiple discount types, tax settings, and entity references

**Current Issues:**
- Overwhelming number of fields presented at once
- Complex business logic not guided (discount types, taxation)
- Multiple entity references (Product, Service, Labor, etc.) without clear selection flow
- No calculation preview

**Proposed Improvements:**
- Multi-step wizard interface:
  1. Item Type Selection (Product/Service/Labor/etc.)
  2. Item Selection with search/browse
  3. Pricing Configuration with live calculations
  4. Advanced Options (discounts, taxes)
- Conditional field display based on item type
- Real-time price calculations
- Template/favorites system for common items

### 4. Checklist/Task Management
Operations for structured task workflows:

- **`ticketChecklistItemCreate`**: Create checklist items with position, importance, and knowledge base links

**Current Issues:**
- Position field without context of existing items
- Knowledge base linking by ID only
- No template or standard checklist support

**Proposed Improvements:**
- Visual position selector showing existing items
- Knowledge base search and preview
- Checklist templates for common procedures
- Bulk import from templates
- Drag-and-drop reordering preview

### 5. Attachment Query Operations
Specialized query modules for attachments:

- **`configurationItemAttachmentsQuery`**
- **`companiesAttachmentsQuery`**
- **`opportunityAttachmentsQuery`**
- **`projectAttachmentsQuery`**
- **`taskAttachmentsQuery`**
- **`ticketAttachmentsQuery`**
- **`timeEntryAttachmentsQuery`**

**Current Issues:**
- Limited filter fields (only ID and dates)
- No file type or size filtering
- No attachment preview capabilities
- Generic interface doesn't reflect attachment-specific needs

**Proposed Improvements:**
- Add file type filters (documents, images, etc.)
- Add file size range filtering
- Add attachment metadata search
- Implement attachment preview/thumbnail
- Add bulk download capabilities
- Context-aware filtering (e.g., recent attachments, by uploader)

### 6. Event/Webhook Operations
Advanced trigger operations:

- **`ticketEvents`**: Custom webhook configuration with field-level triggers

**Current Issues:**
- Complex array configuration for webhook fields
- Field ID selection without field name context
- No validation for webhook payload size
- Advanced configuration hidden in array structure

**Proposed Improvements:**
- Visual field selector with preview
- Field grouping and categorization
- Payload size estimation and warnings
- Webhook testing interface
- Configuration templates for common scenarios

## Implementation Priority

### High Priority
1. **Quote Item Creation**: Most complex with highest impact on sales workflows
2. **Contract Service Adjustment**: Missing core functionality (quantity field)
3. **Association Operations**: High frequency of use, significant UX gains

### Medium Priority
4. **Checklist Management**: Good UX wins for service desk workflows
5. **Attachment Queries**: Specialized improvements for document management

### Lower Priority
6. **Event Configuration**: Advanced feature, smaller user base

## Technical Implementation Strategy

### Phase 1: Field Enhancement
- Convert ID fields to searchable dropdowns using existing entity query modules
- Add calculated fields and real-time validation
- Implement conditional field display

### Phase 2: Workflow Optimization
- Multi-step wizards for complex operations
- Template and favorites systems
- Bulk operation capabilities

### Phase 3: Advanced Features
- Visual configuration interfaces
- Preview and testing capabilities
- Integration with related modules

## Common Patterns for Special Operations

### 1. Entity Reference Fields
**Standard Pattern:**
```json
{
  "name": "entityID",
  "type": "number",
  "label": "Entity ID",
  "required": true
}
```

**Enhanced Pattern:**
```json
{
  "name": "entityID",
  "type": "select",
  "label": "Entity",
  "required": true,
  "options": {
    "store": "rpc://queryEntities",
    "searchable": true,
    "label": "{{name}}",
    "value": "{{id}}",
    "hint": "{{description}}"
  }
}
```

### 2. Complex Configuration Arrays
**Standard Pattern:**
```json
{
  "name": "configuration",
  "type": "array",
  "spec": [...]
}
```

**Enhanced Pattern:**
- Visual builder interface
- Drag-and-drop configuration
- Template selection
- Real-time validation

### 3. Calculated Fields
**Standard Pattern:**
- Static number inputs

**Enhanced Pattern:**
- Real-time calculations
- Impact preview
- Range validation
- Currency formatting

## Success Metrics

### User Experience
- Reduced configuration time for complex operations
- Lower error rates in field completion
- Increased adoption of advanced features

### Technical
- Reduced support tickets for configuration issues
- Improved data quality through better validation
- Higher success rates for complex workflows

## Next Steps

1. **Immediate**: Fix critical missing fields (contract service quantity)
2. **Short-term**: Implement entity dropdowns for high-traffic operations
3. **Medium-term**: Build wizard interfaces for complex configurations
4. **Long-term**: Advanced features like templates and bulk operations

This analysis provides the foundation for Phase 2 of the special operations enhancement project.
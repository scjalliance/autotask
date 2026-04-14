# Module Review: Create/Add Actions

## Overview

This document tracks the systematic review and improvement of Autotask Make app modules that create or add new records. These modules represent user-facing workflows where proper field validation and user experience are critical.

## Priority Assessment

After auditing all 31 Create/Add modules, we identified field types that negatively impact user experience:
- **Number fields for picklist values**: Users must guess correct integer codes
- **Missing validation**: No client-side validation for invalid values
- **Poor UX**: Dropdowns are more intuitive than number inputs

## Completed Fixes

### ✅ ticketsCreate (2026-04-13)

**Status**: **COMPLETE**

The `ticketsCreate` module was identified as the highest priority fix due to having 16 dropdown-worthy fields.

**Fields Fixed**:

1. **priority** (number → select)
   - Options: Critical (4), High (1), Medium (2), Low (3), Information (5)
   - Default: Medium (2)

2. **status** (number → select)
   - Options: New (1), In Progress (8), Complete (5), etc.
   - Default: New (1)
   - Includes all 14 status options from Autotask

3. **source** (number → select)
   - Options: Phone (2), Email (4), Client Portal (-1), etc.
   - Default: Phone (2)
   - Includes all 10 source options

4. **ticketType** (number → select)
   - Options: Service Request (1), Incident (2), Problem (3), Change Request (4), Alert (5)
   - Default: Service Request (1)

**Implementation Details**:
- Used Autotask REST API to fetch exact picklist values
- Applied successful pattern from `ticketNoteCreate` fix
- Set sensible defaults based on `isDefaultValue` from API
- Ordered options logically (priority by severity, status by workflow)

**API Calls Used**:
```bash
./scripts/get-autotask-field-info.sh Ticket priority
./scripts/get-autotask-field-info.sh Ticket status
./scripts/get-autotask-field-info.sh Ticket ticketType
./scripts/get-autotask-field-info.sh Ticket source
```

**Verification**: All four fields now display as proper select dropdowns in Make with correct options and defaults.

## Next Targets

Based on impact analysis, the next modules to review are:

1. **projectCreate** - 11 dropdown-worthy fields
2. **resourcesCreate** - 10 dropdown-worthy fields
3. **companiesCreate** - 8 dropdown-worthy fields
4. **contactsCreate** - 8 dropdown-worthy fields

## Module Analysis Summary

Total modules reviewed: 31 Create/Add actions
- **High Priority** (8+ dropdown fields): 5 modules
- **Medium Priority** (3-7 dropdown fields): 15 modules
- **Low Priority** (1-2 dropdown fields): 11 modules

Full inventory available in `/docs/make-app-modules-inventory.md`.

## Implementation Pattern

The successful pattern for fixing dropdown fields:

1. **Get exact picklist values** from Autotask API
   ```bash
   envwith -f .secrets/.env -- ./scripts/get-autotask-field-info.sh <Entity> <field>
   ```

2. **Update module via Make API**:
   ```bash
   curl -X PUT -H "Authorization: Token $MAKE_API_KEY" \
     -H "Content-Type: application/json" \
     "$MAKE_API_URL/sdk/apps/scj-autotask-nn8loi/1/modules/<module>/expect" \
     -d '[updated field definitions]'
   ```

3. **Convert field definition**:
   ```json
   // FROM:
   {"name": "priority", "type": "number", "label": "Priority"}

   // TO:
   {
     "name": "priority",
     "type": "select",
     "label": "Priority",
     "default": "2",
     "options": [
       {"value": "4", "label": "Critical"},
       {"value": "1", "label": "High"},
       {"value": "2", "label": "Medium"},
       {"value": "3", "label": "Low"},
       {"value": "5", "label": "Information"}
     ]
   }
   ```

4. **Test in Make interface** to verify dropdowns load correctly

This pattern provides immediate UX improvements by replacing cryptic number inputs with user-friendly dropdown selections.
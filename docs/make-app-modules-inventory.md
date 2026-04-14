# Autotask Make App Module Inventory

**Generated:** 2026-04-13
**App:** scj-autotask-nn8loi (version 1)
**Total Modules:** 117

## Overview

This document provides a comprehensive inventory of all modules in the Autotask Make custom app, categorized by type, visibility, and entity. This inventory serves as the foundation for systematic review and enhancement of user experience across all modules.

## Module Breakdown by Type

| Type ID | Type Name | Count | Description |
|---------|-----------|-------|-------------|
| 4 | Action | 69 | Modules that perform operations (create, update, delete, etc.) |
| 9 | Search | 42 | Modules that query or search for records |
| 10 | Instant | 6 | Instant trigger/webhook modules |
| **Total** | | **117** | |

## Public vs Private Modules

| Visibility | Count | Percentage |
|-----------|-------|-----------|
| Public | 60 | 51.3% |
| Private | 57 | 48.7% |

**Note:** Public modules are visible and usable by end users. Private modules are typically for internal use or testing.

## Module Categories by Entity Type

The Autotask API covers the following primary entity types, each typically having 3 modules (Create/Get, Update, Search/Watch/Delete/etc.):

### Core Business Entities
- **Tickets** - Support ticket management (3 modules)
- **Tasks** - Project and ticket tasks (3 modules)
- **Contacts** - Contact records (3 modules)
- **Companies** - Company records (3 modules)
- **Opportunities** - Sales opportunities (3 modules)

### Project Management
- **Projects** - Project records (3 modules)
- **Phases** - Project phases (2 modules)
- **Project Notes** - Project note records (1 public module)
- **Project Charges** - Project charge records (1 private module)

### Service & Contract Management
- **Services** - Service records (3 modules)
- **Contracts** - Contract records (3 modules)
- **Contract Services** - Services within contracts (1 module)
- **Contract Charges** - Contract charge records (1 private module)
- **Contract Milestones** - Contract milestone records (1 private module)

### Quote & Product Management
- **Quotes** - Quote records (3 modules)
- **Quote Items** - Items within quotes (1 private module)
- **Products** - Product records (3 modules)

### Asset Management
- **Configuration Items** - IT assets/configuration items (3 modules)
- **Configuration Item Notes** - Asset note records (1 public module)

### Administrative
- **Companies** - Company records (3 modules)
- **Company Locations** - Company location records (3 modules)
- **Contacts** - Contact records (3 modules)
- **Contact Groups** - Contact group management (2 modules)
- **Departments** - Department records (3 modules)
- **Roles** - Role/security role records (3 modules)
- **Resources** - Resource/employee records (2 modules)

### Time & Billing
- **Time Entries** - Time entry records (3 modules)

### Related Records
- **Ticket Notes** - Ticket note records (1 public module)
- **Ticket Checklist Items** - Checklist item management (1 public module)
- **Company Notes** - Company note records (1 public module)
- **Task Notes** - Task note records (1 public module)
- **Contract Notes** - Contract note records (1 private module)

## High-Priority Modules for Review

These modules are flagged for priority review as they are **public-facing action modules** with **Create or Add operations**, which typically have the highest likelihood of UX issues (invalid picklists, poor field types, missing defaults):

### Public Create/Add Actions (16 modules)
1. **Create Ticket** (ticketsCreate)
   - Entity: Ticket
   - Type: Action (4)
   - Public: Yes

2. **Create Task** (tasksCreate)
   - Entity: Task
   - Type: Action (4)
   - Public: Yes

3. **Create Contact** (contactsCreate)
   - Entity: Contact
   - Type: Action (4)
   - Public: Yes

4. **Create Company Location** (companylocationsCreate)
   - Entity: Company Location
   - Type: Action (4)
   - Public: Yes

5. **Create Department** (departmentsCreate)
   - Entity: Department
   - Type: Action (4)
   - Public: Yes

6. **Create Configuration Item** (configurationItemsCreate)
   - Entity: Configuration Item
   - Type: Action (4)
   - Public: Yes

7. **Create Project** (projectsCreate)
   - Entity: Project
   - Type: Action (4)
   - Public: Yes

8. **Create Project Phase** (phaseCreate)
   - Entity: Project Phase
   - Type: Action (4)
   - Public: Yes

9. **Create Time Entry** (timeEntriesCreate)
   - Entity: Time Entry
   - Type: Action (4)
   - Public: Yes

10. **Create Ticket Note** (ticketNoteCreate)
    - Entity: Ticket Note
    - Type: Action (4)
    - Public: Yes

11. **Create Task Note** (taskNoteCreate)
    - Entity: Task Note
    - Type: Action (4)
    - Public: Yes

12. **Create Company Note** (companiesNoteCreate)
    - Entity: Company Note
    - Type: Action (4)
    - Public: Yes

13. **Create Project Note** (projectNoteCreate)
    - Entity: Project Note
    - Type: Action (4)
    - Public: Yes

14. **Create Asset Note** (configurationItemNoteCreate)
    - Entity: Configuration Item Note
    - Type: Action (4)
    - Public: Yes

15. **Add Ticket Checklist Item** (ticketChecklistItemCreate)
    - Entity: Ticket Checklist Item
    - Type: Action (4)
    - Public: Yes

16. **Add Contact to Group** (contactGroupContactAdd)
    - Entity: Contact Group
    - Type: Action (4)
    - Public: Yes

## Module Types

- **Triggers** (typeId=1): 0 modules - Not implemented in this app
- **Actions** (typeId=4): 69 modules - Operations that modify data (create, update, delete, add)
- **Feeders** (typeId=6): 0 modules - Not implemented in this app
- **Search/Query** (typeId=9): 42 modules - Operations that retrieve/search data
- **Instant** (typeId=10): 6 modules - Instant triggers/webhooks
- **Responders** (typeId=11): 0 modules - Not implemented in this app
- **Universal** (typeId=12): 0 modules - Not implemented in this app

## Review Strategy

The module review will proceed in phases:

1. **Phase 1: High-Priority Modules** - Focus on the 16 public Create/Add action modules identified above
2. **Phase 2: Search/Query Modules** - Review the 42 search modules for appropriate field types and filters
3. **Phase 3: Other Action Modules** - Review remaining private and less-critical action modules
4. **Phase 4: Special Cases** - Handle edge cases and instant trigger modules

## Module Files

The following files contain detailed module metadata:

- `modules-raw.json` - Raw JSON export of all module metadata
- `modules-inventory.csv` - CSV format: name, label, typeId, public

## Data Quality Considerations

When reviewing modules, look for:

1. **Invalid Picklist Values** - Dropdowns with values that don't match the API
2. **Field Type Mismatches** - Text fields when numbers are needed, numbers when text is needed
3. **Missing Defaults** - Required fields without sensible defaults
4. **Unclear Documentation** - Field labels and descriptions that don't match Autotask documentation
5. **API Parameter Misalignment** - Parameters that don't match the actual Autotask API specifications
6. **User Experience Issues** - Fields in wrong order, confusing grouping, missing help text
7. **Validation Gaps** - Missing validation for field constraints

## Next Steps

1. ✅ Generate complete module inventory
2. → Build audit framework with field type analysis tools (Task 2)
3. → Systematically review Create/Add action modules (Task 3)
4. → Review query/search modules (Task 4)
5. → Review triggers and watcher modules (Task 5)
6. → Handle special operations and edge cases (Task 6)
7. → Build testing and validation suite (Task 7)
8. → Create documentation and user guide (Task 8)

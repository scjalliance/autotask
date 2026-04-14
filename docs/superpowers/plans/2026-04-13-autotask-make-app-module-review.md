# Autotask Make App Module Review and UX Enhancement

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Systematically review and enhance all modules in the Autotask Make custom app to provide excellent user experience with proper field types, validation, defaults, and documentation.

**Architecture:** Audit each module's expect/api/interface sections, identify UX issues (invalid picklist values, poor field types, missing defaults), fix systematically with proper dropdowns, validation, and helpful descriptions.

**Tech Stack:** Make.com Custom App API, Autotask REST API field definitions, bash/curl for automation

---

## Analysis Summary

From the Create Ticket Note fix, we identified these common UX anti-patterns:
- Number fields for picklist values → Should be select dropdowns
- Missing default values → Users confused about valid inputs
- Poor field labels/descriptions → Users don't understand purpose
- Optional fields that cause errors → Remove or make explicit

## Module Categories Expected

Based on the initial module scan, we expect these categories:
- **CRUD Operations:** Create/update entities (tickets, contacts, companies, etc.)
- **Query/Search:** Find entities with filters
- **Workflow Actions:** Add items to entities (notes, attachments, checklist items)
- **Special Operations:** Adjust services, markup operations, etc.
- **Triggers/Watchers:** Monitor entity changes

---

### Task 1: Complete Module Inventory and Categorization

**Files:**
- Create: `docs/make-app-modules-inventory.md`

- [ ] **Step 1: Get complete module list with details**

```bash
set -a; source .secrets/.env
curl -s -H "Authorization: Token $MAKE_API_KEY" \
  "$MAKE_API_URL/sdk/apps/scj-autotask-nn8loi/1/modules" | \
  jq '.appModules[] | {name, label, typeId, description, public}' > modules-raw.json
```

Expected: JSON array with all module metadata

- [ ] **Step 2: Categorize modules by purpose**

```bash
# Create organized inventory
cat modules-raw.json | jq -r '"\(.name),\(.label),\(.typeId),\(.public)"' | \
  sort > modules-inventory.csv
```

- [ ] **Step 3: Document module categories in inventory file**

Create `docs/make-app-modules-inventory.md` with:
- Total count of modules
- Breakdown by typeId (1=trigger, 4=action, 6=feeder, 9=search, 10=instant, 11=responder, 12=universal)
- Public vs private modules
- Modules by entity type (Ticket*, Company*, Contact*, etc.)

- [ ] **Step 4: Identify high-priority modules for review**

Focus criteria:
- Public modules (user-facing)
- typeId=4 (actions) - most likely to have UX issues
- Modules with "Create" or "Add" in label - likely to have picklist issues

- [ ] **Step 5: Commit module inventory**

```bash
git add docs/make-app-modules-inventory.md modules-raw.json modules-inventory.csv
git commit -m "docs: add complete Make app module inventory and categorization"
```

---

### Task 2: Audit Framework - Field Type Analysis Tools

**Files:**
- Create: `scripts/audit-make-module.sh`
- Create: `scripts/get-autotask-field-info.sh`

- [ ] **Step 1: Create module auditing script**

```bash
#!/bin/bash
# scripts/audit-make-module.sh
# Usage: ./audit-make-module.sh moduleName

MODULE_NAME="$1"
if [[ -z "$MODULE_NAME" ]]; then
  echo "Usage: $0 <moduleName>"
  exit 1
fi

set -a; source .secrets/.env

echo "=== Module: $MODULE_NAME ==="
echo "--- Basic Info ---"
curl -s -H "Authorization: Token $MAKE_API_KEY" \
  "$MAKE_API_URL/sdk/apps/scj-autotask-nn8loi/1/modules/$MODULE_NAME" | \
  jq '{name, label, typeId, description, public}'

echo "--- API Configuration ---"
curl -s -H "Authorization: Token $MAKE_API_KEY" \
  "$MAKE_API_URL/sdk/apps/scj-autotask-nn8loi/1/modules/$MODULE_NAME/api" | \
  jq '.'

echo "--- Expect (Input Fields) ---"
curl -s -H "Authorization: Token $MAKE_API_KEY" \
  "$MAKE_API_URL/sdk/apps/scj-autotask-nn8loi/1/modules/$MODULE_NAME/expect" | \
  jq '.'

echo "--- Interface (Output Fields) ---"
curl -s -H "Authorization: Token $MAKE_API_KEY" \
  "$MAKE_API_URL/sdk/apps/scj-autotask-nn8loi/1/modules/$MODULE_NAME/interface" | \
  jq '.'
```

- [ ] **Step 2: Create Autotask field definition lookup script**

```bash
#!/bin/bash
# scripts/get-autotask-field-info.sh
# Usage: ./get-autotask-field-info.sh EntityName fieldName

ENTITY="$1"
FIELD="$2"
if [[ -z "$ENTITY" || -z "$FIELD" ]]; then
  echo "Usage: $0 <EntityName> <fieldName>"
  echo "Example: $0 TicketNote noteType"
  exit 1
fi

set -a; source .secrets/.env

curl -s \
  -H "ApiIntegrationCode: $AUTOTASK_INTEGRATION_CODE" \
  -H "UserName: $AUTOTASK_USERNAME" \
  -H "Secret: $AUTOTASK_SECRET" \
  "https://webservices5.autotask.net/ATServicesRest/V1.0/${ENTITY}s/entityInformation/fields" | \
  jq ".fields[] | select(.name == \"$FIELD\")"
```

- [ ] **Step 3: Make scripts executable**

```bash
chmod +x scripts/audit-make-module.sh scripts/get-autotask-field-info.sh
```

- [ ] **Step 4: Test audit framework with known module**

```bash
./scripts/audit-make-module.sh ticketNoteCreate
```

Expected: Complete module configuration output

- [ ] **Step 5: Commit audit tools**

```bash
git add scripts/audit-make-module.sh scripts/get-autotask-field-info.sh
git commit -m "feat: add Make module audit and Autotask field lookup tools"
```

---

### Task 3: Systematic Module Review - Create/Add Actions

**Files:**
- Create: `docs/module-review-create-actions.md`
- Modify: Multiple module expect sections via API

- [ ] **Step 1: Identify all Create/Add action modules**

```bash
# Find modules with Create/Add in label
cat modules-raw.json | jq -r 'select(.label | test("Create|Add"; "i")) | .name' > create-modules.txt
cat create-modules.txt
```

- [ ] **Step 2: Audit first batch of Create modules (5 modules)**

```bash
# Audit each Create module
head -5 create-modules.txt | while read module; do
  echo "=== $module ===" >> audit-results.txt
  ./scripts/audit-make-module.sh "$module" >> audit-results.txt
  echo "" >> audit-results.txt
done
```

- [ ] **Step 3: Identify picklist fields needing dropdown conversion**

Review audit-results.txt for:
- Fields with `type: "number"` that should be picklists
- Fields with unhelpful labels
- Missing default values
- Missing help text

Document findings in `docs/module-review-create-actions.md`

- [ ] **Step 4: Fix first high-impact module (based on usage)**

Example module fix pattern:
```bash
# Get entity field definitions
./scripts/get-autotask-field-info.sh Ticket status
./scripts/get-autotask-field-info.sh Ticket priority

# Update module expect section
curl -s -X PUT \
  -H "Authorization: Token $MAKE_API_KEY" \
  -H "Content-Type: application/jsonc" \
  "$MAKE_API_URL/sdk/apps/scj-autotask-nn8loi/1/modules/MODULENAME/expect" \
  -d '[updated field definitions]'
```

- [ ] **Step 5: Test fixed module in Make scenario**

Create test scenario with fixed module, verify:
- Dropdown options load correctly
- Default values are sensible
- Help text is helpful
- Submission works without errors

- [ ] **Step 6: Document fix pattern and commit**

```bash
git add docs/module-review-create-actions.md audit-results.txt
git commit -m "docs: audit Create/Add action modules, fix [specific module]"
```

---

### Task 4: Review Query/Search Modules

**Files:**
- Create: `docs/module-review-query-actions.md`

- [ ] **Step 1: Identify query/search modules**

```bash
# Find modules with Query/Search in label or typeId=9
cat modules-raw.json | jq -r 'select(.label | test("Query|Search"; "i") or .typeId == 9) | .name' > query-modules.txt
```

- [ ] **Step 2: Audit query module patterns**

Focus on:
- Filter field usability
- Pagination settings
- Output field selection
- Performance considerations

- [ ] **Step 3: Identify common query improvements**

- Better filter field labels
- Default pagination limits
- Helpful examples in field descriptions
- RPC-driven dynamic dropdowns for entity selection

- [ ] **Step 4: Fix top 3 query modules**

Apply improvements systematically

- [ ] **Step 5: Commit query module improvements**

```bash
git add docs/module-review-query-actions.md
git commit -m "feat: improve query module UX with better filters and pagination"
```

---

### Task 5: Review Trigger and Watcher Modules

**Files:**
- Create: `docs/module-review-triggers.md`

- [ ] **Step 1: Identify trigger modules (typeId=1,10)**

```bash
cat modules-raw.json | jq -r 'select(.typeId == 1 or .typeId == 10) | .name' > trigger-modules.txt
```

- [ ] **Step 2: Audit webhook configuration**

Focus on:
- Parameter clarity for webhook setup
- Default polling intervals
- Filter configuration for relevant events

- [ ] **Step 3: Improve webhook parameter UX**

- Clear labels for webhook fields
- Helpful descriptions for filtering options
- Sensible defaults for polling frequency

- [ ] **Step 4: Test webhook triggers**

Verify:
- Webhook registration works smoothly
- Event filtering behaves as expected
- Output fields are comprehensive

- [ ] **Step 5: Commit trigger improvements**

```bash
git add docs/module-review-triggers.md
git commit -m "feat: enhance trigger module configuration and webhook setup UX"
```

---

### Task 6: Special Operations and Edge Cases

**Files:**
- Create: `docs/module-review-special-ops.md`

- [ ] **Step 1: Identify unique operation modules**

Modules that don't fit standard CRUD patterns:
- Adjust operations
- Markup operations
- File operations
- Workflow-specific actions

- [ ] **Step 2: Audit special operation requirements**

Each may need custom field validation, specific parameter formats

- [ ] **Step 3: Create tailored improvements**

Focus on making complex operations more intuitive

- [ ] **Step 4: Commit special operations improvements**

```bash
git add docs/module-review-special-ops.md
git commit -m "feat: improve special operations modules with tailored UX enhancements"
```

---

### Task 7: Module Testing and Validation Suite

**Files:**
- Create: `scripts/test-module.sh`
- Create: `docs/module-testing-checklist.md`

- [ ] **Step 1: Create module testing script**

```bash
#!/bin/bash
# scripts/test-module.sh
# Tests a module by creating a minimal test scenario

MODULE_NAME="$1"
if [[ -z "$MODULE_NAME" ]]; then
  echo "Usage: $0 <moduleName>"
  exit 1
fi

echo "Testing module: $MODULE_NAME"
# Logic to create test scenario, execute with test data, verify output
```

- [ ] **Step 2: Define testing checklist**

For each fixed module:
- [ ] Dropdown options load correctly
- [ ] Default values are appropriate
- [ ] Help text is clear and helpful
- [ ] Required fields are obvious
- [ ] Optional fields don't cause errors
- [ ] API responses are properly handled
- [ ] Error messages are user-friendly

- [ ] **Step 3: Test all improved modules**

Run through testing checklist for each module that was modified

- [ ] **Step 4: Document testing results**

Create testing report with any remaining issues

- [ ] **Step 5: Commit testing framework**

```bash
git add scripts/test-module.sh docs/module-testing-checklist.md
git commit -m "feat: add comprehensive module testing framework and validation"
```

---

### Task 8: Documentation and User Guide

**Files:**
- Create: `docs/make-app-user-guide.md`
- Update: App description and module documentation

- [ ] **Step 1: Create comprehensive user guide**

Document:
- Common module patterns
- Field value guidelines
- Troubleshooting common issues
- Best practices for each module type

- [ ] **Step 2: Update module descriptions**

Improve in-app descriptions for better discoverability

- [ ] **Step 3: Create quick reference**

Field value cheat sheet for common picklists

- [ ] **Step 4: Commit documentation**

```bash
git add docs/make-app-user-guide.md
git commit -m "docs: add comprehensive Make app user guide and module reference"
```

---

## Success Criteria

- [ ] All public modules reviewed and improved
- [ ] Number fields converted to dropdowns where appropriate
- [ ] All modules have helpful field descriptions
- [ ] Default values set for common use cases
- [ ] Testing framework validates all improvements
- [ ] User guide provides clear module usage instructions
- [ ] Zero 500 errors from validation issues

## Post-Implementation

1. Monitor module usage analytics for adoption improvements
2. Gather user feedback on enhanced UX
3. Create automated testing for regression prevention
4. Consider implementing dynamic field validation via RPCs
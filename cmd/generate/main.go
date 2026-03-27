// Command generate reads a Swagger 2.0 JSON spec and produces Go source files
// for all Autotask entity models and services.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// ---------------------------------------------------------------------------
// Swagger types (only the parts we need)
// ---------------------------------------------------------------------------

type swaggerSpec struct {
	Definitions map[string]swaggerDefinition           `json:"definitions"`
	Paths       map[string]map[string]swaggerOperation `json:"paths"`
}

type swaggerDefinition struct {
	Type       string                     `json:"type"`
	Properties map[string]swaggerProperty `json:"properties"`
}

type swaggerProperty struct {
	Type     string           `json:"type"`
	Format   string           `json:"format"`
	Ref      string           `json:"$ref"`
	ReadOnly bool             `json:"readOnly"`
	Items    *swaggerProperty `json:"items"`
}

type swaggerOperation struct {
	OperationID string   `json:"operationId"`
	Tags        []string `json:"tags"`
}

// ---------------------------------------------------------------------------
// Parsed entity metadata
// ---------------------------------------------------------------------------

type entityInfo struct {
	Tag        string
	TypeName   string
	EntityPath string
	CanGet     bool
	CanQuery   bool
	CanCreate  bool
	CanUpdate  bool
	CanPatch   bool
	CanDelete  bool
}

type childInfo struct {
	Tag          string
	TypeName     string
	ParentPath   string
	ChildSegment string
	MethodName   string
	CanGet       bool
	CanQuery     bool
	CanCreate    bool
	CanUpdate    bool
	CanPatch     bool
	CanDelete    bool
}

// tagOps tracks operations and paths for a swagger tag.
type tagOps struct {
	ops   map[string]bool
	paths map[string]bool
}

// ---------------------------------------------------------------------------
// Suffix replacements for Go naming
// ---------------------------------------------------------------------------

var suffixReplacements = []struct {
	lower, upper string
}{
	{"Id", "ID"},
	{"Url", "URL"},
	{"Api", "API"},
	{"Udf", "UDF"},
	{"Ip", "IP"},
	{"Dns", "DNS"},
	{"Ssl", "SSL"},
	{"Rma", "RMA"},
	{"Sla", "SLA"},
}

// exportedName converts a camelCase JSON name to an exported Go name.
func exportedName(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	result := string(runes)

	for _, sr := range suffixReplacements {
		result = fixSuffix(result, sr.lower, sr.upper)
	}
	return result
}

// fixSuffix replaces occurrences of a title-cased abbreviation with its
// all-caps form at word boundaries.
func fixSuffix(s, lower, upper string) string {
	for {
		idx := strings.Index(s, lower)
		if idx < 0 {
			return s
		}
		after := idx + len(lower)
		atEnd := after == len(s)
		followedByUpper := !atEnd && (unicode.IsUpper(rune(s[after])) || unicode.IsDigit(rune(s[after])))
		if atEnd || followedByUpper {
			s = s[:idx] + upper + s[after:]
		} else {
			break
		}
	}
	return s
}

// goType converts a swagger property to its Go type string.
func goType(prop swaggerProperty) string {
	if prop.Ref != "" {
		return refToGoType(prop.Ref)
	}
	if prop.Type == "array" && prop.Items != nil {
		elem := goType(*prop.Items)
		elem = strings.TrimPrefix(elem, "*")
		return "[]" + elem
	}
	switch prop.Type {
	case "integer":
		return "*int64"
	case "number":
		return "*float64"
	case "string":
		return "*string"
	case "boolean":
		return "*bool"
	default:
		return "any"
	}
}

// refToGoType converts a $ref string to a Go type.
func refToGoType(ref string) string {
	name := strings.TrimPrefix(ref, "#/definitions/")
	if strings.HasPrefix(name, "Expression[") {
		return "*int64"
	}
	if name == "UserDefinedField" {
		return "UDF"
	}
	if name == "Object" {
		return "any"
	}
	if name == "Byte" {
		return "byte"
	}
	// Handle names with special characters (e.g. "KeyValuePair`2") by falling
	// back to any, since these aren't valid Go identifiers.
	name = strings.TrimSuffix(name, "Model")
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return "any"
		}
	}
	return "*" + name
}

// ---------------------------------------------------------------------------
// Exclusion filters
// ---------------------------------------------------------------------------

var excludedDefinitions = map[string]bool{
	"Object":         true,
	"Byte":           true,
	"CollectionItem": true,
}

func isExcludedDef(name string) bool {
	if excludedDefinitions[name] {
		return true
	}
	if strings.HasPrefix(name, "QueryActionResult") {
		return true
	}
	if strings.HasPrefix(name, "ItemQueryResultModel") {
		return true
	}
	if name == "OperationResultModel" {
		return true
	}
	return false
}

func isApiIntegrationTag(tag string) bool {
	return strings.HasSuffix(tag, "ApiIntegration")
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	specPath := flag.String("spec", "", "path to swagger JSON (default: glob from cache)")
	outDir := flag.String("out", ".", "output directory")
	flag.Parse()

	if *specPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		matches, err := filepath.Glob(filepath.Join(home, ".cache/api-explorer/apis/autotask/raw/*/swagger-apisguru.json"))
		if err != nil || len(matches) == 0 {
			log.Fatal("no swagger spec found; use -spec flag")
		}
		sort.Strings(matches)
		*specPath = matches[len(matches)-1]
	}

	data, err := os.ReadFile(*specPath)
	if err != nil {
		log.Fatalf("reading spec: %v", err)
	}

	var spec swaggerSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		log.Fatalf("parsing spec: %v", err)
	}

	// Collect tags and their capabilities
	tagDataMap := map[string]*tagOps{}
	for path, methods := range spec.Paths {
		for _, op := range methods {
			for _, tag := range op.Tags {
				if isApiIntegrationTag(tag) {
					continue
				}
				td, ok := tagDataMap[tag]
				if !ok {
					td = &tagOps{ops: map[string]bool{}, paths: map[string]bool{}}
					tagDataMap[tag] = td
				}
				if parts := strings.SplitN(op.OperationID, "_", 2); len(parts) == 2 {
					td.ops[parts[1]] = true
				}
				td.paths[path] = true
			}
		}
	}

	// Build model name set
	modelNames := map[string]bool{}
	for defName := range spec.Definitions {
		if !strings.HasSuffix(defName, "Model") || isExcludedDef(defName) {
			continue
		}
		typeName := strings.TrimSuffix(defName, "Model")
		modelNames[typeName] = true
	}

	// Determine which tags have UDF endpoints
	tagsWithUDF := map[string]bool{}
	for _, methods := range spec.Paths {
		for _, op := range methods {
			if strings.HasSuffix(op.OperationID, "_QueryUserDefinedFieldDefinitions") {
				for _, tag := range op.Tags {
					tagsWithUDF[tag] = true
				}
			}
		}
	}

	// Regex for child entity path pattern
	childPathRe := regexp.MustCompile(`^/V1\.0/(\w+)/\{parentId\}/(\w+)$`)

	// Classify entities
	var topLevel []entityInfo
	var children []childInfo

	for tag, td := range tagDataMap {
		if strings.HasSuffix(tag, "Child") {
			continue
		}

		// Determine entity path by deriving from any known path.
		// First try: a bare path (for entities with POST/PUT/PATCH).
		// Fallback: strip known suffixes from any path we have.
		entityPath := ""
		for p := range td.paths {
			if !strings.Contains(p, "/query") &&
				!strings.Contains(p, "/entityInformation") &&
				!strings.Contains(p, "/{id}") &&
				!strings.Contains(p, "{parentId}") {
				entityPath = p
				break
			}
		}
		if entityPath == "" {
			// Derive from suffixed paths (read-only entities have no bare path).
			for p := range td.paths {
				if strings.Contains(p, "{parentId}") {
					continue
				}
				base := p
				for _, suffix := range []string{"/query/count", "/query", "/entityInformation/userDefinedFields", "/entityInformation/fields", "/entityInformation"} {
					if strings.HasSuffix(base, suffix) {
						base = strings.TrimSuffix(base, suffix)
						break
					}
				}
				if idx := strings.Index(base, "/{id}"); idx >= 0 {
					base = base[:idx]
				}
				if base != "" && strings.HasPrefix(base, "/V1.0/") {
					entityPath = base
					break
				}
			}
		}
		if entityPath == "" {
			continue
		}

		typeName := singularize(tag)
		if !modelNames[typeName] {
			continue
		}

		topLevel = append(topLevel, entityInfo{
			Tag:        tag,
			TypeName:   typeName,
			EntityPath: entityPath,
			CanGet:     td.ops["QueryItem"],
			CanQuery:   td.ops["Query"] || td.ops["UrlParameterQuery"],
			CanCreate:  td.ops["CreateEntity"],
			CanUpdate:  td.ops["UpdateEntity"],
			CanPatch:   td.ops["PatchEntity"],
			CanDelete:  td.ops["DeleteEntity"],
		})
	}
	sort.Slice(topLevel, func(i, j int) bool {
		return topLevel[i].Tag < topLevel[j].Tag
	})

	// Process child entities
	for tag, td := range tagDataMap {
		if !strings.HasSuffix(tag, "Child") {
			continue
		}

		var parentSegment, childSegment string
		for p := range td.paths {
			m := childPathRe.FindStringSubmatch(p)
			if m != nil {
				parentSegment = m[1]
				childSegment = m[2]
				break
			}
		}
		if parentSegment == "" || childSegment == "" {
			continue
		}

		childTag := strings.TrimSuffix(tag, "Child")
		typeName := singularize(childTag)
		if !modelNames[typeName] {
			continue
		}

		parentSingular := singularize(parentSegment)
		methodName := parentSingular + childSegment

		children = append(children, childInfo{
			Tag:          tag,
			TypeName:     typeName,
			ParentPath:   "/V1.0/" + parentSegment,
			ChildSegment: childSegment,
			MethodName:   methodName,
			CanGet:       td.ops["QueryItem"],
			CanQuery:     td.ops["Query"] || td.ops["UrlParameterQuery"],
			CanCreate:    td.ops["CreateEntity"],
			CanUpdate:    td.ops["UpdateEntity"],
			CanPatch:     td.ops["PatchEntity"],
			CanDelete:    td.ops["DeleteEntity"],
		})
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].Tag < children[j].Tag
	})

	// Collect all entity type names we need models for (top-level + children)
	neededTypes := map[string]bool{}
	for _, e := range topLevel {
		neededTypes[e.TypeName] = true
	}
	for _, c := range children {
		neededTypes[c.TypeName] = true
	}

	// Generate models
	modelsCode := generateModels(spec, neededTypes, tagsWithUDF, tagDataMap)
	if err := writeFormatted(filepath.Join(*outDir, "gen_models.go"), modelsCode); err != nil {
		log.Fatalf("writing gen_models.go: %v", err)
	}

	// Generate services
	servicesCode := generateServices(topLevel, children, neededTypes)
	if err := writeFormatted(filepath.Join(*outDir, "gen_services.go"), servicesCode); err != nil {
		log.Fatalf("writing gen_services.go: %v", err)
	}

	fmt.Println("Generated gen_models.go and gen_services.go")
}

// ---------------------------------------------------------------------------
// Model generation
// ---------------------------------------------------------------------------

func generateModels(spec swaggerSpec, neededTypes map[string]bool, tagsWithUDF map[string]bool, tagDataMap map[string]*tagOps) string {
	var b strings.Builder
	b.WriteString("// Code generated by cmd/generate; DO NOT EDIT.\n\n")
	b.WriteString("package autotask\n\n")

	// Collect and sort definition names, only those whose type name is needed
	var defNames []string
	for defName := range spec.Definitions {
		if !strings.HasSuffix(defName, "Model") || isExcludedDef(defName) {
			continue
		}
		typeName := strings.TrimSuffix(defName, "Model")
		if !neededTypes[typeName] {
			continue
		}
		defNames = append(defNames, defName)
	}
	sort.Strings(defNames)

	// Build a map from type name to tag (for UDF check)
	typeToTag := map[string]string{}
	for tag := range tagDataMap {
		typeName := singularize(tag)
		if neededTypes[typeName] {
			typeToTag[typeName] = tag
		}
	}

	for _, defName := range defNames {
		def := spec.Definitions[defName]
		typeName := strings.TrimSuffix(defName, "Model")

		b.WriteString(fmt.Sprintf("// %s represents an Autotask %s entity.\n", typeName, typeName))
		b.WriteString(fmt.Sprintf("type %s struct {\n", typeName))

		// Sort property names for deterministic output
		var propNames []string
		for pn := range def.Properties {
			propNames = append(propNames, pn)
		}
		sort.Strings(propNames)

		hasExplicitUDF := false
		for _, pn := range propNames {
			prop := def.Properties[pn]

			// Skip userDefinedFields — we handle it separately
			if pn == "userDefinedFields" {
				hasExplicitUDF = true
				continue
			}

			goName := exportedName(pn)
			goTypeName := goType(prop)

			comment := ""
			if prop.ReadOnly {
				comment = " // READ-ONLY"
			}

			b.WriteString(fmt.Sprintf("\t%s %s `json:\"%s,omitempty\"`%s\n", goName, goTypeName, pn, comment))
		}

		// Add UserDefinedFields if the definition had the field
		if hasExplicitUDF {
			b.WriteString("\tUserDefinedFields []UDF `json:\"userDefinedFields,omitempty\"`\n")
		}

		b.WriteString("}\n\n")
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// Service generation
// ---------------------------------------------------------------------------

func generateServices(topLevel []entityInfo, children []childInfo, neededTypes map[string]bool) string {
	var b strings.Builder
	b.WriteString("// Code generated by cmd/generate; DO NOT EDIT.\n\n")
	b.WriteString("package autotask\n\n")
	b.WriteString("import \"fmt\"\n\n")

	// Build a function to get the service type name, avoiding collisions with model types.
	svcTypeName := func(typeName string) string {
		name := typeName + "Service"
		// If the service name collides with a model type, disambiguate.
		if neededTypes[name] {
			return typeName + "EntityService"
		}
		return name
	}

	// Service types for top-level entities
	for _, e := range topLevel {
		stn := svcTypeName(e.TypeName)
		b.WriteString(fmt.Sprintf("// %s provides operations for %s entities.\n", stn, e.TypeName))
		b.WriteString(fmt.Sprintf("type %s struct {\n", stn))
		writeTraitFields(&b, e.TypeName, e.CanGet || e.CanQuery, e.CanCreate, e.CanUpdate, e.CanPatch, e.CanDelete)
		b.WriteString("}\n\n")
	}

	// Service types for child entities
	childTypesSeen := map[string]bool{}
	for _, c := range children {
		svcName := c.TypeName + "ChildService"
		if childTypesSeen[svcName] {
			continue
		}
		childTypesSeen[svcName] = true

		b.WriteString(fmt.Sprintf("// %s provides operations for %s child entities.\n", svcName, c.TypeName))
		b.WriteString(fmt.Sprintf("type %s struct {\n", svcName))
		writeTraitFields(&b, c.TypeName, c.CanGet || c.CanQuery, c.CanCreate, c.CanUpdate, c.CanPatch, c.CanDelete)
		b.WriteString("}\n\n")
	}

	// entityServiceFields struct
	b.WriteString("// entityServiceFields holds all top-level entity services.\n")
	b.WriteString("type entityServiceFields struct {\n")
	for _, e := range topLevel {
		b.WriteString(fmt.Sprintf("\t%s %s\n", e.Tag, svcTypeName(e.TypeName)))
	}
	b.WriteString("}\n\n")

	// initServices method
	b.WriteString("// initServices initializes all entity services on the client.\n")
	b.WriteString("func (c *Client) initServices() {\n")
	for _, e := range topLevel {
		hasAny := (e.CanGet || e.CanQuery || e.CanCreate || e.CanUpdate || e.CanPatch || e.CanDelete)
		if !hasAny {
			continue
		}
		baseName := "b" + e.TypeName
		stn := svcTypeName(e.TypeName)
		b.WriteString(fmt.Sprintf("\t%s := baseService{client: c, entityPath: %q, entityName: %q}\n", baseName, e.EntityPath, e.TypeName))
		b.WriteString(fmt.Sprintf("\tc.%s = %s{\n", e.Tag, stn))
		writeTraitInit(&b, baseName, e.TypeName, e.CanGet || e.CanQuery, e.CanCreate, e.CanUpdate, e.CanPatch, e.CanDelete)
		b.WriteString("\t}\n")
	}
	b.WriteString("}\n\n")

	// Build set of top-level tag names (which are also field names on entityServiceFields)
	topLevelTagSet := map[string]bool{}
	for _, e := range topLevel {
		topLevelTagSet[e.Tag] = true
	}

	// Child entity accessor methods — skip when the method name would
	// shadow a top-level entity field on the embedded entityServiceFields.
	methodsSeen := map[string]bool{}
	for _, c := range children {
		if methodsSeen[c.MethodName] {
			continue
		}
		if topLevelTagSet[c.MethodName] {
			continue
		}
		methodsSeen[c.MethodName] = true

		svcName := c.TypeName + "ChildService"
		b.WriteString(fmt.Sprintf("// %s returns a service for %s entities under a parent.\n", c.MethodName, c.TypeName))
		b.WriteString(fmt.Sprintf("func (c *Client) %s(parentID int64) %s {\n", c.MethodName, svcName))
		b.WriteString("\tbase := baseService{\n")
		b.WriteString("\t\tclient:     c,\n")
		b.WriteString(fmt.Sprintf("\t\tentityPath: fmt.Sprintf(\"%s/%%d/%s\", parentID),\n", c.ParentPath, c.ChildSegment))
		b.WriteString(fmt.Sprintf("\t\tentityName: %q,\n", c.TypeName))
		b.WriteString("\t}\n")
		b.WriteString(fmt.Sprintf("\treturn %s{\n", svcName))
		writeTraitInit(&b, "base", c.TypeName, c.CanGet || c.CanQuery, c.CanCreate, c.CanUpdate, c.CanPatch, c.CanDelete)
		b.WriteString("\t}\n")
		b.WriteString("}\n\n")
	}

	return b.String()
}

func writeTraitFields(b *strings.Builder, typeName string, canRead, canCreate, canUpdate, canPatch, canDelete bool) {
	if canRead {
		b.WriteString(fmt.Sprintf("\tReader[%s]\n", typeName))
	}
	if canCreate {
		b.WriteString(fmt.Sprintf("\tCreator[%s]\n", typeName))
	}
	if canUpdate {
		b.WriteString(fmt.Sprintf("\tUpdater[%s]\n", typeName))
	}
	if canPatch {
		b.WriteString(fmt.Sprintf("\tPatcher[%s]\n", typeName))
	}
	if canDelete {
		b.WriteString(fmt.Sprintf("\tDeleter[%s]\n", typeName))
	}
}

func writeTraitInit(b *strings.Builder, baseName, typeName string, canRead, canCreate, canUpdate, canPatch, canDelete bool) {
	// In Go, embedded generic types like Reader[T] have the field name "Reader".
	if canRead {
		b.WriteString(fmt.Sprintf("\t\tReader: Reader[%s]{baseService: %s},\n", typeName, baseName))
	}
	if canCreate {
		b.WriteString(fmt.Sprintf("\t\tCreator: Creator[%s]{baseService: %s},\n", typeName, baseName))
	}
	if canUpdate {
		b.WriteString(fmt.Sprintf("\t\tUpdater: Updater[%s]{baseService: %s},\n", typeName, baseName))
	}
	if canPatch {
		b.WriteString(fmt.Sprintf("\t\tPatcher: Patcher[%s]{baseService: %s},\n", typeName, baseName))
	}
	if canDelete {
		b.WriteString(fmt.Sprintf("\t\tDeleter: Deleter[%s]{baseService: %s},\n", typeName, baseName))
	}
}

// ---------------------------------------------------------------------------
// Singularize
// ---------------------------------------------------------------------------

func singularize(s string) string {
	// Handle "Child" suffix first
	s = strings.TrimSuffix(s, "Child")

	switch s {
	case "Companies":
		return "Company"
	case "Currencies":
		return "Currency"
	case "Taxes":
		return "Tax"
	case "Statuses":
		return "Status"
	case "AttachmentInfo":
		return "AttachmentInfo"
	case "NotificationHistory":
		return "NotificationHistory"
	case "TicketHistory":
		return "TicketHistory"
	case "TagAliases":
		return "TagAlias"
	case "InternalLocationWithBusinessHours":
		return "InternalLocationWithBusinessHours"
	case "InventoryStockedItemsAdd":
		return "InventoryStockedItemAdd"
	case "InventoryStockedItemsRemove":
		return "InventoryStockedItemRemove"
	case "InventoryStockedItemsTransfer":
		return "InventoryStockedItemTransfer"
	case "ServiceLevelAgreementResults":
		return "ServiceLevelAgreementResults"
	case "SurveyResults":
		return "SurveyResults"
	case "TicketCategoryFieldDefaults":
		return "TicketCategoryFieldDefaults"
	}

	// *ies -> *y
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	// *sses -> *ss (e.g. "Addresses" won't exist but just in case)
	if strings.HasSuffix(s, "sses") {
		return s[:len(s)-2]
	}
	// *ses -> *s (e.g. TagAliases → TagAlias, if not caught by special cases)
	if strings.HasSuffix(s, "ses") && !strings.HasSuffix(s, "sses") {
		return s[:len(s)-1]
	}
	// *xes -> *x
	if strings.HasSuffix(s, "xes") {
		return s[:len(s)-2]
	}
	// *s -> remove trailing s (but not "ss")
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		return s[:len(s)-1]
	}
	return s
}

func writeFormatted(path string, code string) error {
	formatted, err := format.Source([]byte(code))
	if err != nil {
		// Write unformatted for debugging
		_ = os.WriteFile(path+".broken", []byte(code), 0o644)
		return fmt.Errorf("formatting %s: %w\n\nRaw output written to %s.broken", path, err, path)
	}
	return os.WriteFile(path, formatted, 0o644)
}

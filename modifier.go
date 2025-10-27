// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	// MaxModifierChainDepth limits the number of chained modifiers to prevent
	// stack overflow or excessive processing time.
	MaxModifierChainDepth = 20
)

// Modifier transforms a Result value.
// Modifiers are applied after path resolution completes.
//
// Example: Get(xml, "items.item.@price|@sort|@last")
//  1. Path "items.item.@price" resolves to Result (array of prices)
//  2. @sort modifier sorts the array
//  3. @last modifier returns the last element
//
// Thread Safety: Modifier implementations must be safe for concurrent use.
// Each modifier receives a copy of the Result and returns a new Result.
type Modifier interface {
	// Apply transforms the input Result and returns a new Result.
	// Input Result must not be modified; return a new Result instead.
	//
	// If the modifier cannot be applied (e.g., @sort on non-array),
	// return the input Result unchanged or return an error Result.
	Apply(r Result) Result

	// Name returns the modifier name (e.g., "reverse", "sort").
	// Used for registration and path parsing.
	Name() string
}

// ModifierFunc is a function adapter for simple modifiers.
// Allows registering function-based modifiers without defining a struct.
type ModifierFunc struct {
	name string
	fn   func(Result) Result
}

// NewModifierFunc creates a new function-based modifier with the given name.
// This is the correct way to create ModifierFunc instances.
//
// Example:
//
//	uppercase := NewModifierFunc("uppercase", func(r Result) Result {
//	    return Result{Type: r.Type, Str: strings.ToUpper(r.Str), Raw: r.Raw}
//	})
//	xmldot.RegisterModifier("uppercase", uppercase)
func NewModifierFunc(name string, fn func(Result) Result) Modifier {
	return &ModifierFunc{name: name, fn: fn}
}

// Apply applies the modifier function to the result.
func (m *ModifierFunc) Apply(r Result) Result {
	return m.fn(r)
}

// Name returns the name of the modifier.
func (m *ModifierFunc) Name() string {
	return m.name
}

// modifierRegistry is a global registry for built-in and custom modifiers.
// Thread-safe for concurrent registration and lookup.
var (
	modifierRegistry = make(map[string]Modifier)
	modifierMu       sync.RWMutex
)

// RegisterModifier registers a custom modifier globally.
// Can be called during init() for package-level modifiers.
//
// Returns error if a modifier with the same name already exists.
//
// Example:
//
//	func init() {
//	    uppercase := xmldot.NewModifierFunc("uppercase", func(r Result) Result {
//	        return Result{Type: r.Type, Str: strings.ToUpper(r.Str), Raw: r.Raw}
//	    })
//	    xmldot.RegisterModifier("uppercase", uppercase)
//	}
func RegisterModifier(name string, m Modifier) error {
	if name == "" {
		return fmt.Errorf("modifier name cannot be empty")
	}

	modifierMu.Lock()
	defer modifierMu.Unlock()

	if _, exists := modifierRegistry[name]; exists {
		return fmt.Errorf("modifier %q already registered", name)
	}

	modifierRegistry[name] = m
	return nil
}

// GetModifier retrieves a registered modifier by name.
// Returns nil if not found.
func GetModifier(name string) Modifier {
	modifierMu.RLock()
	defer modifierMu.RUnlock()
	return modifierRegistry[name]
}

// UnregisterModifier removes a custom modifier from the registry.
// Built-in modifiers cannot be unregistered (returns error).
func UnregisterModifier(name string) error {
	modifierMu.Lock()
	defer modifierMu.Unlock()

	// Protect built-in modifiers from removal
	if isBuiltinModifier(name) {
		return fmt.Errorf("cannot unregister built-in modifier %q", name)
	}

	delete(modifierRegistry, name)
	return nil
}

// isBuiltinModifier checks if a modifier name is built-in (cannot be unregistered)
func isBuiltinModifier(name string) bool {
	builtins := []string{"reverse", "sort", "first", "last", "flatten", "pretty", "ugly"}
	for _, b := range builtins {
		if name == b {
			return true
		}
	}
	return false
}

// applyModifiers applies a chain of modifiers to a Result.
// Modifiers execute left-to-right (pipeline order).
//
// Error Handling:
// When a modifier fails or is not found, the result is Result{Type: Null}.
// This silent failure approach matches the library's philosophy of returning
// empty/null results for query errors rather than panicking.
//
// To debug modifier failures:
//  1. Check Result.Type == Null after Get with modifiers
//  2. Test each modifier individually to isolate the failing one
//  3. Verify custom modifiers are registered before use
//  4. Check that modifier chain depth doesn't exceed MaxModifierChainDepth (20)
//
// Common failure causes:
//   - Unknown modifier name (typo or not registered)
//   - Modifier chain exceeds MaxModifierChainDepth (20 modifiers)
//   - Modifier Apply() returns Null (internal failure)
//   - Modifier expects different input type (e.g., @sort on non-array)
//
// Example debugging:
//
//	result := Get(xml, "items.*|@custom|@sort")
//	if result.Type == Null {
//	    // Check if @custom is registered
//	    _, ok := GetModifier("custom")
//	    if !ok {
//	        log.Println("Modifier 'custom' not found")
//	    }
//	    // Test without @custom to isolate issue
//	    result2 := Get(xml, "items.*|@sort")
//	    if result2.Type != Null {
//	        log.Println("@custom is causing the failure")
//	    }
//	}
//
// Future Enhancement: Consider returning Result with error information
// instead of silent Null to improve debuggability.
func applyModifiers(r Result, modifierNames []string) Result {
	// Security check: limit modifier chain depth
	if len(modifierNames) > MaxModifierChainDepth {
		return Result{Type: Null} // Return error for excessive chaining
	}

	current := r

	for _, name := range modifierNames {
		mod := GetModifier(name)
		if mod == nil {
			// Unknown modifier - return Null to indicate failure
			// Future enhancement: return error type with modifier name
			return Result{Type: Null}
		}

		current = mod.Apply(current)

		// Stop if modifier returned Null - propagate failure
		// Future enhancement: track which modifier failed
		if current.Type == Null {
			break
		}
	}

	return current
}

// parseModifiers extracts modifiers from a path segment.
// Example: "element|@reverse|@first" → element="element", modifiers=["reverse", "first"]
func parseModifiers(pathPart string) (elementPath string, modifiers []string) {
	parts := strings.Split(pathPart, "|")
	elementPath = parts[0]

	for i := 1; i < len(parts); i++ {
		mod := strings.TrimSpace(parts[i])
		if strings.HasPrefix(mod, "@") {
			modifiers = append(modifiers, mod[1:]) // Strip @ prefix
		}
	}

	return elementPath, modifiers
}

// Core Modifiers Implementation (P6.2)

// reverseModifier reverses array order
type reverseModifier struct{}

func (m *reverseModifier) Name() string { return "reverse" }

func (m *reverseModifier) Apply(r Result) Result {
	if r.Type != Array || len(r.Results) == 0 {
		return r // No-op for non-arrays or empty arrays
	}

	reversed := make([]Result, len(r.Results))
	for i, res := range r.Results {
		reversed[len(r.Results)-1-i] = res
	}

	return Result{
		Type:    Array,
		Results: reversed,
	}
}

// sortModifier sorts array elements (numeric or string)
type sortModifier struct{}

func (m *sortModifier) Name() string { return "sort" }

func (m *sortModifier) Apply(r Result) Result {
	if r.Type != Array || len(r.Results) <= 1 {
		return r
	}

	// Create copy to avoid modifying input
	sorted := make([]Result, len(r.Results))
	copy(sorted, r.Results)

	// Attempt numeric sort if all elements are numeric
	allNumeric := true
	for _, res := range sorted {
		if res.Type != Number {
			// Check if string can be parsed as number
			if res.Type == String || res.Type == Element || res.Type == Attribute {
				if _, err := parseFloat64(res.Str); err != nil {
					allNumeric = false
					break
				}
			} else {
				allNumeric = false
				break
			}
		}
	}

	if allNumeric {
		// Numeric sort
		sort.Slice(sorted, func(i, j int) bool {
			var numI, numJ float64
			if sorted[i].Type == Number {
				numI = sorted[i].Num
			} else {
				numI, _ = parseFloat64(sorted[i].Str)
			}
			if sorted[j].Type == Number {
				numJ = sorted[j].Num
			} else {
				numJ, _ = parseFloat64(sorted[j].Str)
			}
			return numI < numJ
		})
	} else {
		// String sort
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].String() < sorted[j].String()
		})
	}

	return Result{Type: Array, Results: sorted}
}

// firstModifier returns first element of array
type firstModifier struct{}

func (m *firstModifier) Name() string { return "first" }

func (m *firstModifier) Apply(r Result) Result {
	if r.Type == Array && len(r.Results) > 0 {
		return r.Results[0]
	}
	return r // Single elements return as-is
}

// lastModifier returns last element of array
type lastModifier struct{}

func (m *lastModifier) Name() string { return "last" }

func (m *lastModifier) Apply(r Result) Result {
	if r.Type == Array && len(r.Results) > 0 {
		return r.Results[len(r.Results)-1]
	}
	return r
}

// flattenModifier flattens nested arrays one level
type flattenModifier struct{}

func (m *flattenModifier) Name() string { return "flatten" }

func (m *flattenModifier) Apply(r Result) Result {
	if r.Type != Array {
		return r
	}

	var flattened []Result
	for _, res := range r.Results {
		if res.Type == Array {
			flattened = append(flattened, res.Results...)
		} else {
			flattened = append(flattened, res)
		}
	}

	if len(flattened) == 0 {
		return Result{Type: Null}
	}

	return Result{Type: Array, Results: flattened}
}

// prettyModifier formats XML with indentation
type prettyModifier struct{}

func (m *prettyModifier) Name() string { return "pretty" }

func (m *prettyModifier) Apply(r Result) Result {
	if r.Raw == "" {
		return r
	}

	// Parse and re-encode with indentation
	// Use xml.Decoder to parse and xml.Encoder to format
	decoder := xml.NewDecoder(strings.NewReader(r.Raw))
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	// Copy tokens from decoder to encoder
	for {
		token, err := decoder.Token()
		if err != nil {
			// End of document or parse error
			if err.Error() == "EOF" {
				break
			}
			// Return unchanged on error
			return r
		}

		// Fix duplicate xmlns by clearing Name.Space on StartElement and EndElement tokens
		// xml.Decoder sets both Name.Space (to the namespace URI) and adds xmlns attributes
		// xml.Encoder sees Name.Space is set and adds xmlns again, causing duplicates
		// Solution: Clear namespace from both start and end elements since xmlns attrs preserve it
		if startElem, ok := token.(xml.StartElement); ok {
			// Deduplicate xmlns attributes first
			dedupedAttrs := deduplicateXmlnsAttrs(startElem.Attr)

			// Create a completely new StartElement with cleared namespace
			newStartElem := xml.StartElement{
				Name: xml.Name{
					Space: "", // Clear namespace to prevent encoder from adding xmlns
					Local: startElem.Name.Local,
				},
				Attr: dedupedAttrs,
			}

			token = newStartElem
		} else if endElem, ok := token.(xml.EndElement); ok {
			// Also clear namespace from end elements to match the cleared start elements
			token = xml.EndElement{
				Name: xml.Name{
					Space: "",
					Local: endElem.Name.Local,
				},
			}
		}

		if err := encoder.EncodeToken(token); err != nil {
			return r
		}
	}

	// Flush the encoder
	if err := encoder.Flush(); err != nil {
		return r
	}

	return Result{
		Type: r.Type,
		Str:  r.Str,
		Raw:  buf.String(),
		Num:  r.Num,
	}
}

// deduplicateXmlnsAttrs removes duplicate xmlns namespace declarations from attributes.
// Keeps the first occurrence of each unique xmlns declaration.
func deduplicateXmlnsAttrs(attrs []xml.Attr) []xml.Attr {
	if len(attrs) <= 1 {
		return attrs
	}

	seen := make(map[string]bool)
	result := make([]xml.Attr, 0, len(attrs))

	for _, attr := range attrs {
		// Create a unique key for xmlns attributes
		// xmlns attributes have either:
		// - Name.Local = "xmlns" (default namespace)
		// - Name.Space = "xmlns" and Name.Local = prefix (prefixed namespace)
		var key string
		if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
			// Default namespace: xmlns="..."
			key = "xmlns=" + attr.Value
		} else if attr.Name.Space == "xmlns" {
			// Prefixed namespace: xmlns:prefix="..."
			key = "xmlns:" + attr.Name.Local + "=" + attr.Value
		} else {
			// Regular attribute - use name only
			if attr.Name.Space != "" {
				key = attr.Name.Space + ":" + attr.Name.Local
			} else {
				key = attr.Name.Local
			}
		}

		if !seen[key] {
			seen[key] = true
			result = append(result, attr)
		}
	}

	return result
}

// uglyModifier compacts XML (removes whitespace)
type uglyModifier struct{}

func (m *uglyModifier) Name() string { return "ugly" }

func (m *uglyModifier) Apply(r Result) Result {
	if r.Raw == "" {
		return r
	}

	// Remove whitespace between tags
	compacted := compactXML(r.Raw)

	return Result{
		Type: r.Type,
		Str:  r.Str,
		Raw:  compacted,
		Num:  r.Num,
	}
}

// compactXML removes unnecessary whitespace from XML while preserving CDATA sections.
// CDATA sections are preserved verbatim including all whitespace, as they may contain
// pre-formatted text, code snippets, or other content where whitespace is significant.
func compactXML(xmlStr string) string {
	var buf strings.Builder
	buf.Grow(len(xmlStr))

	inTag := false
	inCDATA := false

	for i := 0; i < len(xmlStr); i++ {
		// Check for CDATA start
		if !inTag && i+9 <= len(xmlStr) && xmlStr[i:i+9] == "<![CDATA[" {
			inCDATA = true
			buf.WriteString(xmlStr[i : i+9])
			i += 8
			continue
		}

		// Check for CDATA end
		if inCDATA && i+3 <= len(xmlStr) && xmlStr[i:i+3] == "]]>" {
			inCDATA = false
			buf.WriteString("]]>")
			i += 2
			continue
		}

		c := xmlStr[i]

		// Preserve ALL whitespace in CDATA
		if inCDATA {
			buf.WriteByte(c)
			continue
		}

		// Rest of existing logic for non-CDATA content
		if c == '<' {
			inTag = true
			buf.WriteByte(c)
		} else if c == '>' {
			buf.WriteByte(c)
			inTag = false
		} else if inTag {
			// Keep all whitespace inside tags (attributes)
			buf.WriteByte(c)
		} else if !isWhitespace(c) {
			// Outside tags and not whitespace: keep the character
			buf.WriteByte(c)
		}
		// Skip whitespace outside tags (and not in CDATA)
	}

	return buf.String()
}

// init registers all built-in modifiers
func init() {
	// Register all built-in modifiers
	// Using struct pointers to allow method calls
	modifierRegistry["reverse"] = &reverseModifier{}
	modifierRegistry["sort"] = &sortModifier{}
	modifierRegistry["first"] = &firstModifier{}
	modifierRegistry["last"] = &lastModifier{}
	modifierRegistry["flatten"] = &flattenModifier{}
	modifierRegistry["pretty"] = &prettyModifier{}
	modifierRegistry["ugly"] = &uglyModifier{}
}

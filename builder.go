// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// MaxPositionOffset prevents position arithmetic overflow in deeply nested documents
	// Set to ~1GB max to prevent overflow while allowing large documents
	MaxPositionOffset = 1<<30 - 1
)

// xmlBuilder is an internal XML builder for modifications
type xmlBuilder struct {
	data   []byte
	result strings.Builder
	pos    int
	opts   *Options
}

// newXMLBuilder creates a new XML builder with default options
func newXMLBuilder(data []byte) *xmlBuilder {
	return &xmlBuilder{
		data: data,
		pos:  0,
		opts: DefaultOptions(),
	}
}

// newXMLBuilderWithOptions creates a new XML builder with custom options
func newXMLBuilderWithOptions(data []byte, opts *Options) *xmlBuilder {
	// Treat nil options as default options to prevent nil pointer dereferences
	if opts == nil {
		opts = DefaultOptions()
	}
	return &xmlBuilder{
		data: data,
		pos:  0,
		opts: opts,
	}
}

// valueToXML converts various value types to XML string representation
// Returns the XML string and whether the value should be treated as raw XML
func valueToXML(value interface{}) (string, bool, error) {
	if value == nil {
		return "", false, nil
	}

	switch v := value.(type) {
	case string:
		// Regular string - needs escaping
		return escapeXML(v), false, nil
	case int:
		return fmt.Sprintf("%d", v), false, nil
	case int64:
		return fmt.Sprintf("%d", v), false, nil
	case float64:
		return fmt.Sprintf("%g", v), false, nil
	case float32:
		return fmt.Sprintf("%g", v), false, nil
	case bool:
		if v {
			return "true", false, nil
		}
		return "false", false, nil
	case []byte:
		// Byte slice - treat as raw XML (no escaping)
		return string(v), true, nil
	default:
		// Unsupported type
		return "", false, fmt.Errorf("%w: unsupported type %T", ErrInvalidValue, value)
	}
}

// setElement replaces or creates an element at the specified path
func (b *xmlBuilder) setElement(path []PathSegment, value interface{}) error {
	if len(path) == 0 {
		return ErrInvalidPath
	}

	// Convert value to XML string
	xmlValue, isRaw, err := valueToXML(value)
	if err != nil {
		return err
	}

	// Security check: reject values that are too large
	if len(xmlValue) > MaxValueSize {
		return fmt.Errorf("%w: value exceeds maximum size of %d bytes", ErrInvalidValue, MaxValueSize)
	}

	// Security check: reject documents that are too large
	if len(b.data) > MaxDocumentSize {
		return ErrMalformedXML
	}

	// Check if this is actually an attribute operation
	if len(path) > 0 && path[len(path)-1].Type == SegmentAttribute {
		// This is an attribute set, not element set
		attrName := path[len(path)-1].Value
		elementPath := path[:len(path)-1]

		// Find the parent element
		parser := newXMLParser(b.data)
		location, found := b.findElementLocation(parser, elementPath, 0, 0)
		if !found {
			// Parent element doesn't exist - need to create it
			// For now, return error (Phase 3 can handle creation)
			return fmt.Errorf("%w: parent element not found for attribute %s", ErrInvalidPath, attrName)
		}

		// Modify the attribute
		return b.replaceAttribute(location, attrName, xmlValue)
	}

	// Find the target element
	parser := newXMLParser(b.data)
	location, found := b.findElementLocation(parser, path, 0, 0)

	if found {
		// Element exists - replace it
		return b.replaceElement(location, path[len(path)-1], xmlValue)
	}

	// Element doesn't exist - create it
	return b.createElement(path, xmlValue, isRaw)
}

// elementLocation tracks the position of an element in the source XML
type elementLocation struct {
	startPos      int    // Position of '<' in opening tag
	endTagPos     int    // Position of '<' in closing tag
	contentStart  int    // Position after '>' of opening tag
	contentEnd    int    // Position of '<' in closing tag
	elementName   string // Name of the element
	attrs         map[string]string
	isSelfClosing bool
}

// findElementLocation locates an element in the XML based on the path
// baseOffset tracks the position offset in the original document when recursing into nested content
func (b *xmlBuilder) findElementLocation(parser *xmlParser, segments []PathSegment, segIndex int, baseOffset int) (*elementLocation, bool) {
	// Check for position overflow at function entry
	if baseOffset > MaxPositionOffset {
		return nil, false
	}

	if segIndex >= len(segments) {
		return nil, false
	}

	currentSeg := segments[segIndex]
	isLastSegment := segIndex == len(segments)-1

	// Track matches for array indexing
	matchCount := 0
	needsIndex := !isLastSegment && segments[segIndex+1].Type == SegmentIndex

	for parser.skipToNextElement() {
		elemStartPos := parser.pos
		parser.next() // skip '<'

		elemName, attrs, isSelfClosing := parser.parseElementName()

		// Check if element matches current segment (with options support)
		if !currentSeg.matchesWithOptions(elemName, b.opts) {
			// Skip this element
			if !isSelfClosing {
				parser.parseElementContent(elemName)
			}
			continue
		}

		// We have a match
		if needsIndex {
			// Need to count matches for array index
			if matchCount == segments[segIndex+1].Index {
				// This is the indexed element we want
				if segIndex+2 == len(segments) {
					// This is the target element
					contentStart := parser.pos
					var contentEnd, endTagPos int
					if isSelfClosing {
						contentEnd = parser.pos
						endTagPos = parser.pos
					} else {
						_ = parser.parseElementContent(elemName)
						// parser.pos is now just after the '>' of </name>
						endTagPos = parser.pos - len(elemName) - 3 // at '<' of </name>
						contentEnd = endTagPos
					}
					return &elementLocation{
						startPos:      elemStartPos + baseOffset,
						endTagPos:     endTagPos + baseOffset,
						contentStart:  contentStart + baseOffset,
						contentEnd:    contentEnd + baseOffset,
						elementName:   elemName,
						attrs:         attrs,
						isSelfClosing: isSelfClosing,
					}, true
				}
				// Continue searching within this element
				var content string
				contentStartPos := parser.pos
				if isSelfClosing {
					content = ""
				} else {
					content = parser.parseElementContent(elemName)
				}
				contentParser := newXMLParser([]byte(content))
				// Pass the current content start position as the new base offset
				// Check for overflow before recursing
				newOffset := baseOffset + contentStartPos
				if newOffset < baseOffset || newOffset > MaxPositionOffset {
					// Overflow detected
					return nil, false
				}
				return b.findElementLocation(contentParser, segments, segIndex+2, newOffset)
			}
			matchCount++
			// Skip this element and continue looking
			if !isSelfClosing {
				parser.parseElementContent(elemName)
			}
			continue
		}

		// Not an array operation
		if isLastSegment {
			// Found the target element
			contentStart := parser.pos
			var contentEnd, endTagPos int
			if isSelfClosing {
				contentEnd = parser.pos
				endTagPos = parser.pos
			} else {
				startOfContent := parser.pos
				_ = parser.parseElementContent(elemName)
				// parser.pos is now just after the '>' of </name>
				// We want endTagPos to be at the '<' of </name>
				endTagPos = parser.pos - len(elemName) - 3 // at '<' of </name>
				contentEnd = endTagPos
				_ = startOfContent
			}
			return &elementLocation{
				startPos:      elemStartPos + baseOffset,
				endTagPos:     endTagPos + baseOffset,
				contentStart:  contentStart + baseOffset,
				contentEnd:    contentEnd + baseOffset,
				elementName:   elemName,
				attrs:         attrs,
				isSelfClosing: isSelfClosing,
			}, true
		}

		// Continue searching within this element
		var content string
		contentStartPos := parser.pos
		if isSelfClosing {
			content = ""
		} else {
			content = parser.parseElementContent(elemName)
		}
		contentParser := newXMLParser([]byte(content))
		// Pass the current content start position as the new base offset
		// Check for overflow before recursing
		newOffset := baseOffset + contentStartPos
		if newOffset < baseOffset || newOffset > MaxPositionOffset {
			// Overflow detected
			return nil, false
		}
		return b.findElementLocation(contentParser, segments, segIndex+1, newOffset)
	}

	return nil, false
}

// replaceElement replaces an element's content in the XML
func (b *xmlBuilder) replaceElement(location *elementLocation, segment PathSegment, xmlValue string) error {
	// Check if this is an attribute operation
	if segment.Type == SegmentAttribute {
		return b.replaceAttribute(location, segment.Value, xmlValue)
	}

	// Build the result XML
	b.result.Reset()

	// Write everything up to and including the opening tag (up to contentStart)
	b.result.Write(b.data[:location.contentStart])

	// Write the new value
	b.result.WriteString(xmlValue)

	// Write everything after the content (includes closing tag and rest of document)
	// contentEnd points to the start of the closing tag '<'
	b.result.Write(b.data[location.contentEnd:])

	return nil
}

// replaceAttribute replaces or adds an attribute to an element
func (b *xmlBuilder) replaceAttribute(location *elementLocation, attrName string, attrValue string) error {
	// Build new opening tag with updated attribute
	b.result.Reset()

	// Write everything before the opening tag
	b.result.Write(b.data[:location.startPos])

	// Build new opening tag
	b.result.WriteString("<")
	b.result.WriteString(location.elementName)

	// Sort attribute names for deterministic output
	attrNames := make([]string, 0, len(location.attrs)+1)
	for name := range location.attrs {
		attrNames = append(attrNames, name)
	}

	// Add new/modified attribute to list if not already present
	// For case-insensitive matching, check existing attributes case-insensitively
	attrExists := false
	if !b.opts.CaseSensitive {
		lowerAttrName := toLowerASCII(attrName)
		for _, name := range attrNames {
			if toLowerASCII(name) == lowerAttrName {
				attrExists = true
				break
			}
		}
	} else {
		for _, name := range attrNames {
			if name == attrName {
				attrExists = true
				break
			}
		}
	}
	if !attrExists {
		attrNames = append(attrNames, attrName)
	}

	sort.Strings(attrNames)

	// Write attributes in sorted order
	for _, name := range attrNames {
		b.result.WriteString(" ")
		b.result.WriteString(name)
		b.result.WriteString("=\"")

		// Check if this is the attribute being modified (case-sensitive or insensitive)
		isModifiedAttr := false
		if b.opts.CaseSensitive {
			isModifiedAttr = (name == attrName)
		} else {
			isModifiedAttr = (toLowerASCII(name) == toLowerASCII(attrName))
		}

		if isModifiedAttr {
			b.result.WriteString(attrValue)
		} else {
			b.result.WriteString(escapeXML(location.attrs[name]))
		}
		b.result.WriteString("\"")
	}

	// Close the opening tag
	if location.isSelfClosing {
		b.result.WriteString("/>")
		// Write everything after this tag
		b.result.Write(b.data[location.contentStart:])
	} else {
		b.result.WriteString(">")
		// Write everything after the opening tag
		b.result.Write(b.data[location.contentStart:])
	}

	return nil
}

// createElement creates a new element at the specified path
func (b *xmlBuilder) createElement(path []PathSegment, xmlValue string, isRaw bool) error {
	// Security check
	if len(b.data) == 0 {
		return ErrMalformedXML
	}

	// Find the deepest existing parent in the path
	// We go from the full path backwards, checking if each progressively shorter path exists
	parentDepth := -1
	var parentLocation *elementLocation

	// Try to find the deepest existing parent by checking path[0:i] for i from len-1 down to 1
	for i := len(path) - 1; i >= 1; i-- {
		parser := newXMLParser(b.data)
		partialPath := path[0:i]
		loc, found := b.findElementLocation(parser, partialPath, 0, 0)
		if found {
			parentLocation = loc
			parentDepth = i - 1 // Index of the last found segment
			break
		}
	}

	if parentDepth == -1 || parentLocation == nil {
		// No parent found in path - need to find root element and append
		// Skip the first segment if it matches root
		return b.createInRoot(path, xmlValue, isRaw)
	}

	// Create missing elements from parentDepth+1 to end
	return b.createInParent(parentLocation, path[parentDepth+1:], xmlValue, isRaw)
}

// createInRoot creates element path starting from root
func (b *xmlBuilder) createInRoot(path []PathSegment, xmlValue string, isRaw bool) error {
	// Find the root element to insert into
	parser := newXMLParser(b.data)
	if !parser.skipToNextElement() {
		return ErrMalformedXML
	}

	elemStartPos := parser.pos
	parser.next() // skip '<'
	elemName, attrs, isSelfClosing := parser.parseElementName()

	// Check if the first path segment matches the root element name
	// If yes, skip it and create the rest inside root
	pathToCreate := path
	if len(path) > 0 && path[0].Type == SegmentElement && path[0].Value == elemName {
		pathToCreate = path[1:]
	}

	if len(pathToCreate) == 0 {
		// Nothing to create
		return nil
	}

	if isSelfClosing {
		// Root is self-closing - convert to full element with content
		b.result.Reset()
		// Write everything BEFORE the self-closing tag
		b.result.Write(b.data[:elemStartPos])
		// Build new opening tag
		b.result.WriteString("<")
		b.result.WriteString(elemName)

		// Write attributes if any
		if len(attrs) > 0 {
			for k, v := range attrs {
				b.result.WriteString(" ")
				b.result.WriteString(k)
				b.result.WriteString(`="`)
				b.result.WriteString(escapeXML(v))
				b.result.WriteString(`"`)
			}
		}

		b.result.WriteString(">")

		// Build the path
		b.buildElementPath(pathToCreate, xmlValue, isRaw)

		b.result.WriteString("</")
		b.result.WriteString(elemName)
		b.result.WriteString(">")

		// Write everything AFTER the original self-closing tag
		// parser.pos is now positioned after the '>' of the self-closing tag
		b.result.Write(b.data[parser.pos:])
		return nil
	}

	// Find where root element ends
	_ = parser.parseElementContent(elemName)
	contentEnd := parser.pos - len(elemName) - 3

	b.result.Reset()
	b.result.Write(b.data[:contentEnd])

	// Build the path
	b.buildElementPath(pathToCreate, xmlValue, isRaw)

	b.result.Write(b.data[contentEnd:])
	return nil
}

// createInParent creates missing elements within a parent element
func (b *xmlBuilder) createInParent(parentLocation *elementLocation, remainingPath []PathSegment, xmlValue string, isRaw bool) error {
	b.result.Reset()

	if parentLocation.isSelfClosing {
		// Parent is self-closing - need to convert to full element
		// Write everything before the self-closing tag
		b.result.Write(b.data[:parentLocation.startPos])

		// Rebuild opening tag
		b.result.WriteString("<")
		b.result.WriteString(parentLocation.elementName)

		// Write attributes if any
		if len(parentLocation.attrs) > 0 {
			for k, v := range parentLocation.attrs {
				b.result.WriteString(" ")
				b.result.WriteString(k)
				b.result.WriteString(`="`)
				b.result.WriteString(escapeXML(v))
				b.result.WriteString(`"`)
			}
		}

		b.result.WriteString(">")

		// Build the missing path
		b.buildElementPath(remainingPath, xmlValue, isRaw)

		// Close the parent element
		b.result.WriteString("</")
		b.result.WriteString(parentLocation.elementName)
		b.result.WriteString(">")

		// Write everything after the original self-closing tag
		b.result.Write(b.data[parentLocation.contentEnd:])
	} else {
		// Parent is a regular element - insert before closing tag
		b.result.Write(b.data[:parentLocation.contentEnd])

		// Build the missing path
		b.buildElementPath(remainingPath, xmlValue, isRaw)

		b.result.Write(b.data[parentLocation.contentEnd:])
	}

	return nil
}

// buildElementPath builds a chain of nested elements
func (b *xmlBuilder) buildElementPath(path []PathSegment, xmlValue string, isRaw bool) {
	b.buildElementPathWithDepth(path, xmlValue, isRaw, 0)
}

// buildElementPathWithDepth builds a chain of nested elements with optional indentation
func (b *xmlBuilder) buildElementPathWithDepth(path []PathSegment, xmlValue string, isRaw bool, baseDepth int) {
	indent := b.opts.Indent
	useIndent := indent != ""

	for i, seg := range path {
		if seg.Type == SegmentAttribute {
			// Can't create just an attribute without an element
			continue
		}

		// Add indentation if enabled
		if useIndent {
			b.result.WriteString("\n")
			for j := 0; j < baseDepth+i; j++ {
				b.result.WriteString(indent)
			}
		}

		b.result.WriteString("<")
		b.result.WriteString(seg.Value)
		b.result.WriteString(">")

		if i == len(path)-1 {
			// Last element - add the value
			if isRaw {
				b.result.WriteString(xmlValue)
			} else {
				b.result.WriteString(xmlValue)
			}
		}
	}

	// Close all elements in reverse order
	for i := len(path) - 1; i >= 0; i-- {
		if path[i].Type == SegmentAttribute {
			continue
		}
		b.result.WriteString("</")
		b.result.WriteString(path[i].Value)
		b.result.WriteString(">")
	}

	// Add final newline if indenting
	if useIndent && baseDepth > 0 {
		b.result.WriteString("\n")
		for j := 0; j < baseDepth-1; j++ {
			b.result.WriteString(indent)
		}
	}
}

// deleteElement removes an element or attribute at the specified path
func (b *xmlBuilder) deleteElement(path []PathSegment) error {
	if len(path) == 0 {
		return ErrInvalidPath
	}

	// Security check
	if len(b.data) > MaxDocumentSize {
		return ErrMalformedXML
	}

	// Check if we're deleting an attribute
	lastSeg := path[len(path)-1]
	if lastSeg.Type == SegmentAttribute {
		// Find the parent element
		if len(path) < 2 {
			return ErrInvalidPath
		}
		parentPath := path[:len(path)-1]
		parser := newXMLParser(b.data)
		location, found := b.findElementLocation(parser, parentPath, 0, 0)
		if !found {
			// Parent doesn't exist - return original XML unchanged
			b.result.Reset()
			b.result.Write(b.data)
			return nil
		}
		return b.deleteAttribute(location, lastSeg.Value)
	}

	// Deleting an element
	parser := newXMLParser(b.data)
	location, found := b.findElementLocation(parser, path, 0, 0)
	if !found {
		// Element doesn't exist - return original XML unchanged
		b.result.Reset()
		b.result.Write(b.data)
		return nil
	}

	// Delete the element
	b.result.Reset()
	b.result.Write(b.data[:location.startPos])

	// Skip the element and its closing tag
	if location.isSelfClosing {
		// Self-closing tag: <tag attr="val"/>
		// location.contentStart points to position after '/>', so we can write from there
		b.result.Write(b.data[location.contentStart:])
	} else {
		// Regular element: <tag>content</tag>
		// Skip past the closing tag </name>
		endPos := location.endTagPos + len(location.elementName) + 3 // </name>
		b.result.Write(b.data[endPos:])
	}

	return nil
}

// deleteAttribute removes an attribute from an element
func (b *xmlBuilder) deleteAttribute(location *elementLocation, attrName string) error {
	// Check if attribute exists (case-sensitive or insensitive)
	attrFound := false
	if b.opts.CaseSensitive {
		if _, exists := location.attrs[attrName]; exists {
			attrFound = true
		}
	} else {
		lowerAttrName := toLowerASCII(attrName)
		for name := range location.attrs {
			if toLowerASCII(name) == lowerAttrName {
				attrFound = true
				break
			}
		}
	}

	if !attrFound {
		// Attribute doesn't exist - return original XML unchanged
		b.result.Reset()
		b.result.Write(b.data)
		return nil
	}

	// Build new opening tag without the deleted attribute
	b.result.Reset()
	b.result.Write(b.data[:location.startPos])

	b.result.WriteString("<")
	b.result.WriteString(location.elementName)

	// Sort attribute names for deterministic output
	attrNames := make([]string, 0, len(location.attrs))
	for name := range location.attrs {
		// Skip the attribute to be deleted (case-sensitive or insensitive)
		shouldSkip := false
		if b.opts.CaseSensitive {
			shouldSkip = (name == attrName)
		} else {
			shouldSkip = (toLowerASCII(name) == toLowerASCII(attrName))
		}
		if !shouldSkip {
			attrNames = append(attrNames, name)
		}
	}

	sort.Strings(attrNames)

	// Copy all attributes except the one being deleted
	for _, name := range attrNames {
		b.result.WriteString(" ")
		b.result.WriteString(name)
		b.result.WriteString("=\"")
		b.result.WriteString(escapeXML(location.attrs[name]))
		b.result.WriteString("\"")
	}

	// Close the opening tag
	if location.isSelfClosing {
		b.result.WriteString("/>")
		b.result.Write(b.data[location.contentStart:])
	} else {
		b.result.WriteString(">")
		b.result.Write(b.data[location.contentStart:])
	}

	return nil
}

// getResult returns the built XML string
func (b *xmlBuilder) getResult() string {
	if b.result.Len() == 0 {
		return string(b.data)
	}
	return b.result.String()
}

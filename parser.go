// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"strings"
)

// Security limits to protect against malicious XML
const (
	// MaxNestingDepth is the maximum allowed nesting depth for XML elements.
	// This prevents stack overflow attacks with deeply nested structures.
	MaxNestingDepth = 100

	// MaxDocumentSize is the maximum allowed size for an XML document in bytes.
	// This prevents memory exhaustion attacks with extremely large documents.
	// Set to 10MB by default.
	MaxDocumentSize = 10 * 1024 * 1024

	// MaxValueSize is the maximum allowed size for a single value in Set operations.
	// This prevents DoS attacks via memory exhaustion with extremely large values.
	// Set to 5MB by default.
	MaxValueSize = 5 * 1024 * 1024

	// MaxAttributes is the maximum number of attributes allowed per element.
	// This prevents DoS attacks with attribute flooding.
	MaxAttributes = 100

	// MaxTokenSize is the maximum size for any single token (element name, attribute value, etc).
	// This prevents buffer overrun attacks with extremely large tokens.
	MaxTokenSize = 1024 * 1024 // 1MB

	// MaxNamespacePrefixLength is the maximum allowed length for namespace prefixes.
	// This prevents memory exhaustion attacks with excessively long namespace prefixes.
	MaxNamespacePrefixLength = 256
)

// xmlParser is an incremental XML parser that parses only what's needed.
//
// CONCURRENCY: xmlParser instances are NOT safe for concurrent use.
// Each goroutine must create its own xmlParser instance via newXMLParser().
// The parser maintains internal state (position, depth) that would be corrupted
// by concurrent access.
//
// Example concurrent usage:
//
//	func processXML(xml string) {
//	    var wg sync.WaitGroup
//	    wg.Add(2)
//
//	    // Each goroutine gets its own parser
//	    go func() {
//	        defer wg.Done()
//	        parser := newXMLParser([]byte(xml))
//	        // Use parser...
//	    }()
//
//	    go func() {
//	        defer wg.Done()
//	        parser := newXMLParser([]byte(xml))
//	        // Use parser...
//	    }()
//
//	    wg.Wait()
//	}
type xmlParser struct {
	data        []byte
	pos         int
	depth       int
	filterDepth int
	dataLen     int // Cache data length to avoid repeated len() calls
}

// newXMLParser creates a new XML parser
func newXMLParser(data []byte) *xmlParser {
	return &xmlParser{
		data:    data,
		pos:     0,
		depth:   0,
		dataLen: len(data), // Cache length for performance
	}
}

// skipWhitespace advances the position past any whitespace characters
// Optimized: Use cached dataLen and inline isWhitespace check
func (p *xmlParser) skipWhitespace() {
	for p.pos < p.dataLen && isWhitespace(p.data[p.pos]) {
		p.pos++
	}
}

// isWhitespace checks if a byte is a whitespace character
// Inlined for performance in hot path
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// peek returns the next character without advancing position
// Optimized: Use cached dataLen
func (p *xmlParser) peek() byte {
	if p.pos < p.dataLen {
		return p.data[p.pos]
	}
	return 0
}

// next advances position and returns the current character
// Optimized: Use cached dataLen
func (p *xmlParser) next() byte {
	if p.pos < p.dataLen {
		c := p.data[p.pos]
		p.pos++
		return c
	}
	return 0
}

// readUntil reads until the specified byte is found
// Security: Limits token size to prevent buffer overrun attacks
// Returns empty string if token exceeds MaxTokenSize to signal error condition
// Optimized: Use cached dataLen
func (p *xmlParser) readUntil(delim byte) string {
	start := p.pos
	for p.pos < p.dataLen {
		// Security check: enforce maximum token size
		if p.pos-start > MaxTokenSize {
			// Token too large - return empty string to signal error
			return ""
		}
		if p.data[p.pos] == delim {
			result := string(p.data[start:p.pos])
			return result
		}
		p.pos++
	}
	return string(p.data[start:])
}

// readUntilAny reads until any of the specified bytes is found
// Security: Limits token size to prevent buffer overrun attacks
// Returns empty string if token exceeds MaxTokenSize to signal error condition
// Optimized: Use cached dataLen
func (p *xmlParser) readUntilAny(delims string) string {
	start := p.pos
	delimsLen := len(delims)
	for p.pos < p.dataLen {
		// Security check: enforce maximum token size
		if p.pos-start > MaxTokenSize {
			// Token too large - return empty string to signal error
			return ""
		}
		c := p.data[p.pos]
		for i := 0; i < delimsLen; i++ {
			if c == delims[i] {
				result := string(p.data[start:p.pos])
				return result
			}
		}
		p.pos++
	}
	return string(p.data[start:])
}

// parseAttributes extracts attributes from an element opening tag
// Returns a map of attribute names to values
// Optimized: Pre-allocate map with capacity hint to reduce allocations
func (p *xmlParser) parseAttributes() map[string]string {
	attrs := make(map[string]string, 4) // Most elements have 0-4 attributes
	attrCount := 0

	for {
		p.skipWhitespace()

		// Security check: enforce maximum attribute count to prevent DoS attacks
		if attrCount >= MaxAttributes {
			break
		}

		// Check if we've reached the end of the tag
		if p.pos >= p.dataLen {
			break
		}

		c := p.peek()
		if c == '>' || c == '/' {
			break
		}

		// Read attribute name
		name := p.readUntilAny("= \t\n\r/>")
		if name == "" {
			break
		}
		attrCount++

		p.skipWhitespace()

		// Expect '='
		if p.peek() != '=' {
			break
		}
		p.next()

		p.skipWhitespace()

		// Read attribute value (can be quoted with " or ')
		quote := p.peek()
		if quote == '"' || quote == '\'' {
			p.next() // skip opening quote
			value := p.readUntil(quote)
			p.next() // skip closing quote
			attrs[name] = unescapeXML(value)
		} else {
			// Unquoted attribute value (read until whitespace or >)
			value := p.readUntilAny(" \t\n\r/>")
			attrs[name] = unescapeXML(value)
		}
	}

	return attrs
}

// parseElementName extracts the element name and attributes from an opening tag
// Assumes the parser is positioned after the '<' character
// Returns: elementName, attributes, isSelfClosing, error
func (p *xmlParser) parseElementName() (string, map[string]string, bool) {
	// Read element name (until whitespace, '>', or '/')
	name := p.readUntilAny(" \t\n\r/>")

	// Parse attributes
	attrs := p.parseAttributes()

	// Check for self-closing tag
	isSelfClosing := false
	if p.peek() == '/' {
		p.next()
		isSelfClosing = true
	}

	// Skip closing '>'
	if p.peek() == '>' {
		p.next()
	}

	return name, attrs, isSelfClosing
}

// parseElementContent extracts the content between opening and closing tags
// Returns the text content, handling nested elements and escaping
func (p *xmlParser) parseElementContent(elementName string) string {
	// Track nesting depth to prevent stack overflow attacks
	p.depth++
	if p.depth > MaxNestingDepth {
		// Exceeded maximum nesting depth - stop parsing and return empty
		p.depth--
		return ""
	}
	defer func() { p.depth-- }()

	var content strings.Builder
	// Track element-specific depth for matching tags
	elementDepth := 1

	for p.pos < p.dataLen && elementDepth > 0 {
		c := p.peek()

		if c == '<' {
			// Check if this is a closing tag, opening tag, or comment/CDATA
			if p.pos+1 < p.dataLen {
				next := p.data[p.pos+1]

				if next == '/' {
					// Closing tag
					p.next() // skip '<'
					p.next() // skip '/'
					closeName := p.readUntil('>')
					p.next() // skip '>'

					// Extract just the element name (before any whitespace)
					closeName = strings.TrimSpace(closeName)
					if idx := strings.IndexAny(closeName, " \t\n\r"); idx >= 0 {
						closeName = closeName[:idx]
					}

					if closeName == elementName {
						elementDepth--
						if elementDepth == 0 {
							break
						}
					}
					content.WriteString("</")
					content.WriteString(closeName)
					content.WriteString(">")
				} else if next == '!' {
					// Comment or CDATA - include in content for now
					content.WriteByte(c)
					p.next()
				} else {
					// Opening tag of nested element
					p.next() // skip '<'
					nestedName := p.readUntilAny(" \t\n\r/>")

					// Write the opening tag
					content.WriteString("<")
					content.WriteString(nestedName)

					// Parse attributes and check for self-closing
					attrs := p.parseAttributes()
					for attrName, attrValue := range attrs {
						content.WriteString(" ")
						content.WriteString(attrName)
						content.WriteString("=\"")
						content.WriteString(escapeXML(attrValue))
						content.WriteString("\"")
					}

					isSelfClosing := false
					if p.peek() == '/' {
						p.next()
						isSelfClosing = true
						content.WriteString("/")
					}

					if p.peek() == '>' {
						p.next()
						content.WriteString(">")
					}

					// Only increment elementDepth if this is the same element type we're tracking
					if !isSelfClosing && nestedName == elementName {
						elementDepth++
					}
				}
			} else {
				content.WriteByte(c)
				p.next()
			}
		} else {
			content.WriteByte(c)
			p.next()
		}
	}

	return content.String()
}

// extractTextContent extracts only text content, stripping out all XML tags
func extractTextContent(content string) string {
	var result strings.Builder
	inTag := false

	for i := 0; i < len(content); i++ {
		c := content[i]
		if c == '<' {
			inTag = true
		} else if c == '>' {
			inTag = false
		} else if !inTag {
			result.WriteByte(c)
		}
	}

	return strings.TrimSpace(result.String())
}

// extractDirectTextOnly extracts only direct text content, excluding text from nested elements
// This is used for the % operator
func extractDirectTextOnly(content string) string {
	var result strings.Builder
	inTag := false
	depth := 0

	for i := 0; i < len(content); i++ {
		c := content[i]
		if c == '<' {
			inTag = true
			// Check if it's an opening or closing tag
			if i+1 < len(content) && content[i+1] == '/' {
				// Closing tag
				depth--
			} else if i+1 < len(content) && content[i+1] != '!' && content[i+1] != '?' {
				// Opening tag (not comment or PI)
				depth++
			}
		} else if c == '>' {
			// Check for self-closing tag
			if i > 0 && content[i-1] == '/' {
				depth--
			}
			inTag = false
		} else if !inTag && depth == 0 {
			// Only collect text when not inside a nested element
			result.WriteByte(c)
		}
	}

	return strings.TrimSpace(result.String())
}

// skipToNextElement advances the parser to the next element opening tag
func (p *xmlParser) skipToNextElement() bool {
	for p.pos < p.dataLen {
		if p.peek() == '<' {
			if p.pos+1 < p.dataLen {
				next := p.data[p.pos+1]
				if next != '!' && next != '?' && next != '/' {
					// This is an opening tag
					return true
				}
				// Skip comments, processing instructions, closing tags
				if next == '!' {
					// Check for DOCTYPE declaration (XXE protection)
					// Case-insensitive check to prevent bypass with <!doctype> or mixed case
					if p.pos+8 < p.dataLen {
						doctypeEnd := p.pos + 9
						if doctypeEnd > p.dataLen {
							doctypeEnd = p.dataLen
						}
						doctype := string(p.data[p.pos:doctypeEnd])
						if strings.EqualFold(doctype, "<!DOCTYPE") {
							// DOCTYPE detected - skip it to prevent XXE attacks
							p.pos += len(doctype)
							// Skip to end of DOCTYPE declaration
							depth := 0
							for p.pos < p.dataLen {
								c := p.data[p.pos]
								if c == '[' {
									depth++
								} else if c == ']' {
									depth--
								} else if c == '>' && depth == 0 {
									p.pos++
									break
								}
								p.pos++
							}
							continue
						}
					}
					// Skip comment or CDATA
					p.next() // skip '<'
					p.next() // skip '!'
					if p.pos+1 < p.dataLen && p.data[p.pos] == '-' && p.data[p.pos+1] == '-' {
						// Comment: skip to -->
						for p.pos < p.dataLen-2 {
							if p.data[p.pos] == '-' && p.data[p.pos+1] == '-' && p.data[p.pos+2] == '>' {
								p.pos += 3
								break
							}
							p.pos++
						}
					}
				} else if next == '?' {
					// Skip processing instruction
					p.readUntil('>')
					p.next()
				} else {
					// Skip closing tag
					p.readUntil('>')
					p.next()
				}
			} else {
				return false
			}
		} else {
			p.next()
		}
	}
	return false
}

// XML escaping functions

// xmlEscaper is a pre-compiled replacer for efficient XML escaping
var xmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"\"", "&quot;",
	"'", "&apos;",
)

// xmlUnescaper is a pre-compiled replacer for efficient XML unescaping
// Note: Must unescape &amp; last to avoid double-unescaping
var xmlUnescaper = strings.NewReplacer(
	"&lt;", "<",
	"&gt;", ">",
	"&quot;", "\"",
	"&apos;", "'",
	"&amp;", "&",
)

// escapeXML escapes special XML characters using a pre-compiled replacer
// for better performance than multiple ReplaceAll calls
func escapeXML(s string) string {
	return xmlEscaper.Replace(s)
}

// unescapeXML unescapes XML entity references using a pre-compiled replacer
// for better performance than multiple ReplaceAll calls
func unescapeXML(s string) string {
	return xmlUnescaper.Replace(s)
}

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"bytes"
	"strings"
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
	data    []byte
	pos     int
	dataLen int // Cache data length to avoid repeated len() calls
}

// newXMLParser creates a new XML parser
func newXMLParser(data []byte) *xmlParser {
	return &xmlParser{
		data:    data,
		pos:     0,
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
// Aborts the query with a *LimitError if the token exceeds MaxTokenSize
// Optimized: Use cached dataLen
func (p *xmlParser) readUntil(delim byte) string {
	start := p.pos
	for p.pos < p.dataLen {
		// Security check: enforce maximum token size
		if p.pos-start > MaxTokenSize {
			exceedLimit("MaxTokenSize", p.pos-start, MaxTokenSize)
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
// Aborts the query with a *LimitError if the token exceeds MaxTokenSize
// Optimized: Use cached dataLen
func (p *xmlParser) readUntilAny(delims string) string {
	start := p.pos
	delimsLen := len(delims)
	for p.pos < p.dataLen {
		// Security check: enforce maximum token size
		if p.pos-start > MaxTokenSize {
			exceedLimit("MaxTokenSize", p.pos-start, MaxTokenSize)
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

		// Security check: enforce maximum attribute count to prevent DoS attacks
		if attrCount > MaxAttributes {
			exceedLimit("MaxAttributes", attrCount, MaxAttributes)
		}

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
// Security: Content nested more than MaxNestingDepth levels below elementName (counting
// elementName itself) aborts the query with a *LimitError. Queries scan documents from
// the root, so this enforces the same absolute depth limit as validation.
func (p *xmlParser) parseElementContent(elementName string) string {
	var content strings.Builder
	// Track element-specific depth for matching tags
	elementDepth := 1
	// Track overall nesting depth below and including elementName
	nesting := 1

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
					nesting--

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
					// Comments and CDATA sections are copied verbatim, so that
					// markup inside them is not parsed as elements
					if end := p.skipCommentOrCDATA(); end > p.pos {
						content.Write(p.data[p.pos:end])
						p.pos = end
					} else {
						content.WriteByte(c)
						p.next()
					}
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

					if !isSelfClosing {
						nesting++
						if nesting > MaxNestingDepth {
							exceedLimit("MaxNestingDepth", nesting, MaxNestingDepth)
						}
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
					// Skip processing instruction. Like the validator, count
					// MaxTokenSize from after "<?".
					p.next() // skip '<'
					p.next() // skip '?'
					p.readUntil('>')
					p.next()
				} else {
					// Skip closing tag, counting MaxTokenSize from after "</"
					p.next() // skip '<'
					p.next() // skip '/'
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

// skipCommentOrCDATA returns the position just after the comment or CDATA section
// starting at p.pos, the end of the data if it is unterminated, or p.pos if no
// comment or CDATA section starts there.
func (p *xmlParser) skipCommentOrCDATA() int {
	var start, end string
	switch rest := p.data[p.pos:]; {
	case bytes.HasPrefix(rest, []byte("<!--")):
		start, end = "<!--", "-->"
	case bytes.HasPrefix(rest, []byte("<![CDATA[")):
		start, end = "<![CDATA[", "]]>"
	default:
		return p.pos
	}
	from := p.pos + len(start)
	if idx := bytes.Index(p.data[from:], []byte(end)); idx >= 0 {
		return from + idx + len(end)
	}
	return p.dataLen
}

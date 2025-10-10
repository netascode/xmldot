// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import "fmt"

// ValidateError represents an XML validation error with location information
type ValidateError struct {
	Line    int
	Column  int
	Message string
}

func (e *ValidateError) Error() string {
	return fmt.Sprintf("XML validation error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// validatingParser extends xmlParser with line/column tracking for validation
type validatingParser struct {
	xmlParser
	line       int
	column     int
	tagStack   []tagInfo // Stack of open tags for nesting validation
	rootFound  bool      // Track if we've found a root element
	rootClosed bool      // Track if the root element has been closed
}

// tagInfo tracks information about an open tag
type tagInfo struct {
	name   string
	line   int
	column int
}

// newValidatingParser creates a new validating parser
func newValidatingParser(data []byte) *validatingParser {
	return &validatingParser{
		xmlParser: xmlParser{
			data:    data,
			pos:     0,
			depth:   0,
			dataLen: len(data),
		},
		line:     1,
		column:   0,
		tagStack: make([]tagInfo, 0, 16), // Pre-allocate for typical nesting depth
	}
}

// advance moves the parser forward and updates line/column tracking
func (p *validatingParser) advance() byte {
	if p.pos >= p.dataLen {
		return 0
	}

	c := p.data[p.pos]
	p.pos++

	if c == '\n' {
		p.line++
		p.column = 0
	} else {
		p.column++
	}

	return c
}

// peekChar returns the current character without advancing
func (p *validatingParser) peekChar() byte {
	if p.pos >= p.dataLen {
		return 0
	}
	return p.data[p.pos]
}

// validate performs full document validation
func (p *validatingParser) validate() *ValidateError {
	// Check document size
	if p.dataLen > MaxDocumentSize {
		return &ValidateError{
			Line:    1,
			Column:  0,
			Message: fmt.Sprintf("document exceeds maximum size of %d bytes", MaxDocumentSize),
		}
	}

	// Empty document is invalid
	if p.dataLen == 0 {
		return &ValidateError{
			Line:    1,
			Column:  0,
			Message: "empty document",
		}
	}

	// Parse the entire document
	for p.pos < p.dataLen {
		p.skipWhitespaceTracked()

		if p.pos >= p.dataLen {
			break
		}

		c := p.peekChar()

		if c == '<' {
			if err := p.parseTag(); err != nil {
				return err
			}
		} else {
			// Text content outside root element is invalid
			if !p.rootFound || p.rootClosed {
				// Check if this is non-whitespace content
				if !isWhitespace(c) {
					return &ValidateError{
						Line:    p.line,
						Column:  p.column,
						Message: "content not allowed outside root element",
					}
				}
				// Whitespace is okay outside root - just advance
				p.advance()
			} else {
				// Content inside root element - just advance
				p.advance()
			}
		}
	}

	// Check if we have unclosed tags
	if len(p.tagStack) > 0 {
		unclosed := p.tagStack[len(p.tagStack)-1]
		return &ValidateError{
			Line:    unclosed.line,
			Column:  unclosed.column,
			Message: fmt.Sprintf("unclosed tag '%s'", unclosed.name),
		}
	}

	// Check if we found a root element
	if !p.rootFound {
		return &ValidateError{
			Line:    1,
			Column:  0,
			Message: "no root element found",
		}
	}

	return nil
}

// skipWhitespaceTracked skips whitespace while tracking line/column
func (p *validatingParser) skipWhitespaceTracked() {
	for p.pos < p.dataLen && isWhitespace(p.data[p.pos]) {
		p.advance()
	}
}

// parseTag parses an XML tag (opening, closing, comment, etc.)
func (p *validatingParser) parseTag() *ValidateError {
	tagLine := p.line
	tagColumn := p.column

	p.advance() // skip '<'

	if p.pos >= p.dataLen {
		return &ValidateError{
			Line:    tagLine,
			Column:  tagColumn,
			Message: "unexpected end of document in tag",
		}
	}

	next := p.peekChar()

	// Processing instruction
	if next == '?' {
		p.advance()
		return p.skipUntilTracked('>')
	}

	// Comment or CDATA
	if next == '!' {
		p.advance()

		// Check for DOCTYPE (forbidden for security)
		if p.pos+7 <= p.dataLen {
			doctype := string(p.data[p.pos : p.pos+7])
			if doctype == "DOCTYPE" || doctype == "doctype" {
				// Skip DOCTYPE for security but don't treat as error
				return p.skipDOCTYPE()
			}
		}

		// Comment
		if p.pos+1 < p.dataLen && p.data[p.pos] == '-' && p.data[p.pos+1] == '-' {
			p.advance()
			p.advance()
			return p.skipComment()
		}

		// CDATA
		if p.pos+7 <= p.dataLen && string(p.data[p.pos:p.pos+7]) == "[CDATA[" {
			for i := 0; i < 7; i++ {
				p.advance()
			}
			return p.skipCDATA()
		}

		// Unknown declaration - skip to >
		return p.skipUntilTracked('>')
	}

	// Closing tag
	if next == '/' {
		p.advance()
		return p.parseClosingTag(tagLine, tagColumn)
	}

	// Opening tag
	return p.parseOpeningTag(tagLine, tagColumn)
}

// parseOpeningTag parses an opening tag
func (p *validatingParser) parseOpeningTag(tagLine, tagColumn int) *ValidateError {
	// Read element name
	nameStart := p.pos
	nameLine := p.line
	nameColumn := p.column

	name := p.readNameTracked()
	if name == "" {
		if p.pos > nameStart {
			return &ValidateError{
				Line:    nameLine,
				Column:  nameColumn,
				Message: "element name exceeds maximum token size",
			}
		}
		return &ValidateError{
			Line:    nameLine,
			Column:  nameColumn,
			Message: "empty element name",
		}
	}

	// Validate element name
	if err := validateName(name); err != nil {
		return &ValidateError{
			Line:    nameLine,
			Column:  nameColumn,
			Message: fmt.Sprintf("invalid element name '%s': %s", name, err),
		}
	}

	// Check nesting depth
	if len(p.tagStack) >= MaxNestingDepth {
		return &ValidateError{
			Line:    tagLine,
			Column:  tagColumn,
			Message: fmt.Sprintf("nesting depth exceeds maximum of %d", MaxNestingDepth),
		}
	}

	// Parse attributes
	if err := p.parseAttributesValidating(); err != nil {
		return err
	}

	// Check for self-closing tag
	p.skipWhitespaceTracked()
	isSelfClosing := false
	if p.peekChar() == '/' {
		p.advance()
		isSelfClosing = true
	}

	// Expect closing '>'
	if p.peekChar() != '>' {
		return &ValidateError{
			Line:    p.line,
			Column:  p.column,
			Message: "expected '>' to close tag",
		}
	}
	p.advance()

	// Track root element
	if !p.rootFound {
		p.rootFound = true
	} else if len(p.tagStack) == 0 {
		// Multiple root elements
		return &ValidateError{
			Line:    tagLine,
			Column:  tagColumn,
			Message: "multiple root elements",
		}
	}

	// Push to tag stack if not self-closing
	if !isSelfClosing {
		p.tagStack = append(p.tagStack, tagInfo{
			name:   name,
			line:   tagLine,
			column: tagColumn,
		})
	} else if len(p.tagStack) == 0 {
		// Self-closing root element
		p.rootClosed = true
	}

	return nil
}

// parseClosingTag parses a closing tag
func (p *validatingParser) parseClosingTag(tagLine, tagColumn int) *ValidateError {
	nameLine := p.line
	nameColumn := p.column

	name := p.readNameTracked()
	if name == "" {
		return &ValidateError{
			Line:    nameLine,
			Column:  nameColumn,
			Message: "empty closing tag name",
		}
	}

	p.skipWhitespaceTracked()

	// Expect closing '>'
	if p.peekChar() != '>' {
		return &ValidateError{
			Line:    p.line,
			Column:  p.column,
			Message: "expected '>' to close tag",
		}
	}
	p.advance()

	// Check if we have an opening tag to match
	if len(p.tagStack) == 0 {
		return &ValidateError{
			Line:    tagLine,
			Column:  tagColumn,
			Message: fmt.Sprintf("closing tag '%s' without matching opening tag", name),
		}
	}

	// Check if tag names match
	openTag := p.tagStack[len(p.tagStack)-1]
	if openTag.name != name {
		return &ValidateError{
			Line:    tagLine,
			Column:  tagColumn,
			Message: fmt.Sprintf("mismatched closing tag '%s' (expected '%s' opened at line %d, column %d)", name, openTag.name, openTag.line, openTag.column),
		}
	}

	// Pop from stack
	p.tagStack = p.tagStack[:len(p.tagStack)-1]

	// Mark root as closed if this was the root element
	if len(p.tagStack) == 0 {
		p.rootClosed = true
	}

	return nil
}

// parseAttributesValidating parses and validates attributes
func (p *validatingParser) parseAttributesValidating() *ValidateError {
	attrCount := 0

	for {
		p.skipWhitespaceTracked()

		if p.pos >= p.dataLen {
			return &ValidateError{
				Line:    p.line,
				Column:  p.column,
				Message: "unexpected end of document in tag",
			}
		}

		c := p.peekChar()
		if c == '>' || c == '/' {
			break
		}

		// Check attribute limit
		if attrCount >= MaxAttributes {
			return &ValidateError{
				Line:    p.line,
				Column:  p.column,
				Message: fmt.Sprintf("too many attributes (maximum %d)", MaxAttributes),
			}
		}

		// Read attribute name
		attrLine := p.line
		attrColumn := p.column
		attrName := p.readNameTracked()
		if attrName == "" {
			return &ValidateError{
				Line:    attrLine,
				Column:  attrColumn,
				Message: "invalid attribute name",
			}
		}

		// Validate attribute name
		if err := validateName(attrName); err != nil {
			return &ValidateError{
				Line:    attrLine,
				Column:  attrColumn,
				Message: fmt.Sprintf("invalid attribute name '%s': %s", attrName, err),
			}
		}

		attrCount++

		p.skipWhitespaceTracked()

		// Expect '='
		if p.peekChar() != '=' {
			return &ValidateError{
				Line:    p.line,
				Column:  p.column,
				Message: "expected '=' after attribute name",
			}
		}
		p.advance()

		p.skipWhitespaceTracked()

		// Read attribute value
		quote := p.peekChar()
		if quote != '"' && quote != '\'' {
			return &ValidateError{
				Line:    p.line,
				Column:  p.column,
				Message: "attribute value must be quoted",
			}
		}
		p.advance()

		// Skip to closing quote
		if err := p.skipUntilTracked(quote); err != nil {
			return err
		}
	}

	return nil
}

// readNameTracked reads an XML name while tracking position
func (p *validatingParser) readNameTracked() string {
	start := p.pos

	for p.pos < p.dataLen {
		// Check token size limit
		if p.pos-start > MaxTokenSize {
			return ""
		}

		c := p.data[p.pos]

		// Valid name characters: letter, digit, ., -, _, :
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '.' || c == '-' ||
			c == '_' || c == ':' {
			p.advance()
		} else {
			break
		}
	}

	return string(p.data[start:p.pos])
}

// skipUntilTracked skips until the specified character
func (p *validatingParser) skipUntilTracked(delim byte) *ValidateError {
	start := p.pos

	for p.pos < p.dataLen {
		if p.pos-start > MaxTokenSize {
			return &ValidateError{
				Line:    p.line,
				Column:  p.column,
				Message: "token exceeds maximum size",
			}
		}

		if p.peekChar() == delim {
			p.advance()
			return nil
		}
		p.advance()
	}

	return &ValidateError{
		Line:    p.line,
		Column:  p.column,
		Message: fmt.Sprintf("unexpected end of document (expected '%c')", delim),
	}
}

// skipComment skips a comment (assumes <!-- already consumed)
func (p *validatingParser) skipComment() *ValidateError {
	for p.pos < p.dataLen-2 {
		if p.data[p.pos] == '-' && p.data[p.pos+1] == '-' {
			p.advance()
			p.advance()
			if p.peekChar() == '>' {
				p.advance()
				return nil
			}
			return &ValidateError{
				Line:    p.line,
				Column:  p.column,
				Message: "'--' not allowed in comment",
			}
		}
		p.advance()
	}

	return &ValidateError{
		Line:    p.line,
		Column:  p.column,
		Message: "unclosed comment",
	}
}

// skipCDATA skips a CDATA section
func (p *validatingParser) skipCDATA() *ValidateError {
	for p.pos < p.dataLen-2 {
		if p.data[p.pos] == ']' && p.data[p.pos+1] == ']' && p.data[p.pos+2] == '>' {
			p.advance()
			p.advance()
			p.advance()
			return nil
		}
		p.advance()
	}

	return &ValidateError{
		Line:    p.line,
		Column:  p.column,
		Message: "unclosed CDATA section",
	}
}

// skipDOCTYPE skips a DOCTYPE declaration
func (p *validatingParser) skipDOCTYPE() *ValidateError {
	depth := 0
	for p.pos < p.dataLen {
		c := p.peekChar()
		if c == '[' {
			depth++
		} else if c == ']' {
			depth--
		} else if c == '>' && depth == 0 {
			p.advance()
			return nil
		}
		p.advance()
	}

	return &ValidateError{
		Line:    p.line,
		Column:  p.column,
		Message: "unclosed DOCTYPE declaration",
	}
}

// validateName validates an XML name according to XML 1.0 spec
func validateName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("empty name")
	}

	// First character must be a letter, underscore, or colon
	first := name[0]
	if (first < 'a' || first > 'z') && (first < 'A' || first > 'Z') && first != '_' && first != ':' {
		if first >= '0' && first <= '9' {
			return fmt.Errorf("name cannot start with a digit")
		}
		return fmt.Errorf("invalid first character '%c'", first)
	}

	// Subsequent characters can be letters, digits, ., -, _, or :
	for i := 1; i < len(name); i++ {
		c := name[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') &&
			(c < '0' || c > '9') && c != '.' && c != '-' &&
			c != '_' && c != ':' {
			return fmt.Errorf("invalid character '%c' at position %d", c, i)
		}
	}

	return nil
}

// Valid checks if XML is well-formed
// Returns true if valid, false otherwise
func Valid(xml string) bool {
	return ValidBytes([]byte(xml))
}

// ValidBytes checks if XML is well-formed (byte slice variant)
func ValidBytes(xml []byte) bool {
	parser := newValidatingParser(xml)
	return parser.validate() == nil
}

// ValidateWithError checks XML and returns detailed error on failure
// Returns nil if valid, *ValidateError otherwise
func ValidateWithError(xml string) *ValidateError {
	return ValidateBytesWithError([]byte(xml))
}

// ValidateBytesWithError checks XML and returns detailed error (byte slice variant)
func ValidateBytesWithError(xml []byte) *ValidateError {
	parser := newValidatingParser(xml)
	return parser.validate()
}

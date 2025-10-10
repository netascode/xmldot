// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"testing"
)

func TestXMLParser_SkipWhitespace(t *testing.T) {
	xml := []byte("   \t\n\r<root>")
	parser := newXMLParser(xml)

	parser.skipWhitespace()
	if parser.pos != 6 {
		t.Errorf("skipWhitespace() pos = %d, want 6", parser.pos)
	}
	if parser.peek() != '<' {
		t.Errorf("peek() = %c, want '<'", parser.peek())
	}
}

func TestXMLParser_ParseAttributes(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want map[string]string
	}{
		{
			name: "Single attribute",
			xml:  `id="123">`,
			want: map[string]string{"id": "123"},
		},
		{
			name: "Multiple attributes",
			xml:  `id="123" name="test" active="true">`,
			want: map[string]string{
				"id":     "123",
				"name":   "test",
				"active": "true",
			},
		},
		{
			name: "No attributes",
			xml:  `>`,
			want: map[string]string{},
		},
		{
			name: "Attribute with single quotes",
			xml:  `id='456'>`,
			want: map[string]string{"id": "456"},
		},
		{
			name: "Attribute with spaces",
			xml:  `  id = "789"  >`,
			want: map[string]string{"id": "789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newXMLParser([]byte(tt.xml))
			got := parser.parseAttributes()

			if len(got) != len(tt.want) {
				t.Errorf("parseAttributes() returned %d attrs, want %d", len(got), len(tt.want))
			}

			for key, wantVal := range tt.want {
				gotVal, ok := got[key]
				if !ok {
					t.Errorf("parseAttributes() missing key %q", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("parseAttributes()[%q] = %q, want %q", key, gotVal, wantVal)
				}
			}
		})
	}
}

func TestXMLParser_ParseElementName(t *testing.T) {
	tests := []struct {
		name            string
		xml             string
		wantName        string
		wantSelfClosing bool
		wantAttrCount   int
	}{
		{
			name:            "Simple element",
			xml:             "user>",
			wantName:        "user",
			wantSelfClosing: false,
			wantAttrCount:   0,
		},
		{
			name:            "Element with attribute",
			xml:             `user id="123">`,
			wantName:        "user",
			wantSelfClosing: false,
			wantAttrCount:   1,
		},
		{
			name:            "Self-closing element",
			xml:             "br/>",
			wantName:        "br",
			wantSelfClosing: true,
			wantAttrCount:   0,
		},
		{
			name:            "Self-closing with attributes",
			xml:             `img src="test.jpg"/>`,
			wantName:        "img",
			wantSelfClosing: true,
			wantAttrCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newXMLParser([]byte(tt.xml))
			name, attrs, isSelfClosing := parser.parseElementName()

			if name != tt.wantName {
				t.Errorf("parseElementName() name = %q, want %q", name, tt.wantName)
			}
			if isSelfClosing != tt.wantSelfClosing {
				t.Errorf("parseElementName() selfClosing = %v, want %v", isSelfClosing, tt.wantSelfClosing)
			}
			if len(attrs) != tt.wantAttrCount {
				t.Errorf("parseElementName() attrs count = %d, want %d", len(attrs), tt.wantAttrCount)
			}
		})
	}
}

func TestXMLParser_ParseElementContent(t *testing.T) {
	tests := []struct {
		name        string
		xml         string
		elementName string
		want        string
	}{
		{
			name:        "Simple text",
			xml:         "Hello World</user>",
			elementName: "user",
			want:        "Hello World",
		},
		{
			name:        "Nested elements",
			xml:         "<name>John</name><age>30</age></user>",
			elementName: "user",
			want:        "<name>John</name><age>30</age>",
		},
		{
			name:        "Mixed content",
			xml:         "Text <b>bold</b> more text</p>",
			elementName: "p",
			want:        "Text <b>bold</b> more text",
		},
		{
			name:        "Empty content",
			xml:         "</empty>",
			elementName: "empty",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newXMLParser([]byte(tt.xml))
			got := parser.parseElementContent(tt.elementName)

			if got != tt.want {
				t.Errorf("parseElementContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractTextContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "Simple text",
			content: "Hello World",
			want:    "Hello World",
		},
		{
			name:    "Text with tags",
			content: "Text <b>bold</b> more text",
			want:    "Text bold more text",
		},
		{
			name:    "Nested tags",
			content: "<name>John</name><age>30</age>",
			want:    "John30",
		},
		{
			name:    "Empty content",
			content: "",
			want:    "",
		},
		{
			name:    "Only tags",
			content: "<tag></tag>",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTextContent(tt.content)
			if got != tt.want {
				t.Errorf("extractTextContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Less than",
			input: "<tag>",
			want:  "&lt;tag&gt;",
		},
		{
			name:  "Ampersand",
			input: "Tom & Jerry",
			want:  "Tom &amp; Jerry",
		},
		{
			name:  "Quotes",
			input: `He said "Hello"`,
			want:  `He said &quot;Hello&quot;`,
		},
		{
			name:  "Single quotes",
			input: "It's working",
			want:  "It&apos;s working",
		},
		{
			name:  "Multiple entities",
			input: `<tag attr="value">Tom & Jerry</tag>`,
			want:  `&lt;tag attr=&quot;value&quot;&gt;Tom &amp; Jerry&lt;/tag&gt;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeXML(tt.input)
			if got != tt.want {
				t.Errorf("escapeXML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnescapeXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Less than",
			input: "&lt;tag&gt;",
			want:  "<tag>",
		},
		{
			name:  "Ampersand",
			input: "Tom &amp; Jerry",
			want:  "Tom & Jerry",
		},
		{
			name:  "Quotes",
			input: `He said &quot;Hello&quot;`,
			want:  `He said "Hello"`,
		},
		{
			name:  "Single quotes",
			input: "It&apos;s working",
			want:  "It's working",
		},
		{
			name:  "Multiple entities",
			input: `&lt;tag attr=&quot;value&quot;&gt;Tom &amp; Jerry&lt;/tag&gt;`,
			want:  `<tag attr="value">Tom & Jerry</tag>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unescapeXML(tt.input)
			if got != tt.want {
				t.Errorf("unescapeXML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestXMLParser_SkipToNextElement(t *testing.T) {
	tests := []struct {
		name        string
		xml         string
		wantFound   bool
		wantElement string
	}{
		{
			name:        "Simple element",
			xml:         "<user>",
			wantFound:   true,
			wantElement: "user",
		},
		{
			name:        "With whitespace",
			xml:         "   \n\t<user>",
			wantFound:   true,
			wantElement: "user",
		},
		{
			name:        "Skip comment",
			xml:         "<!-- comment --><user>",
			wantFound:   true,
			wantElement: "user",
		},
		{
			name:        "No elements",
			xml:         "just text",
			wantFound:   false,
			wantElement: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := newXMLParser([]byte(tt.xml))
			found := parser.skipToNextElement()

			if found != tt.wantFound {
				t.Errorf("skipToNextElement() = %v, want %v", found, tt.wantFound)
			}

			if found && tt.wantElement != "" {
				parser.next() // skip '<'
				name, _, _ := parser.parseElementName()
				if name != tt.wantElement {
					t.Errorf("element name = %q, want %q", name, tt.wantElement)
				}
			}
		})
	}
}

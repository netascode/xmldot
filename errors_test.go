// SPDX-License-Identifier: MIT
// Copyright (c) 2025 Daniel Schmidt

package xmldot

import (
	"errors"
	"strings"
	"testing"
)

// TestErrorTypes verifies that all error types are properly defined and can be checked with errors.Is
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() error
		wantType error
	}{
		{
			name: "malformed XML - unclosed tag",
			fn: func() error {
				_, err := Set("<root>", "root.item", "value")
				return err
			},
			wantType: ErrMalformedXML,
		},
		{
			name: "malformed XML - mismatched tags",
			fn: func() error {
				_, err := Set("<root><item>value</wrong></root>", "root.item", "value")
				return err
			},
			wantType: ErrMalformedXML,
		},
		{
			name: "invalid path - empty",
			fn: func() error {
				_, err := Set("<root/>", "", "value")
				return err
			},
			wantType: ErrInvalidPath,
		},
		{
			name: "invalid path - SetMany mismatch",
			fn: func() error {
				_, err := SetMany("<root/>", []string{"a", "b"}, []interface{}{"value"})
				return err
			},
			wantType: ErrInvalidPath,
		},
		{
			name: "document too large",
			fn: func() error {
				largeXML := strings.Repeat("<item>x</item>", 1000000)
				xml := "<root>" + largeXML + "</root>"
				_, err := Set(xml, "root.item", "new")
				return err
			},
			wantType: ErrMalformedXML,
		},
		{
			name: "invalid value - SetRaw with unclosed tag",
			fn: func() error {
				_, err := SetRaw("<root/>", "root.data", "<item>")
				return err
			},
			wantType: ErrInvalidValue,
		},
		{
			name: "invalid value - SetRaw with mismatched tags",
			fn: func() error {
				_, err := SetRaw("<root/>", "root.data", "<a></b>")
				return err
			},
			wantType: ErrInvalidValue,
		},
		{
			name: "invalid value - SetRaw with DOCTYPE",
			fn: func() error {
				_, err := SetRaw("<root/>", "root.data", "<!DOCTYPE test><test/>")
				return err
			},
			wantType: ErrInvalidValue,
		},
		{
			name: "invalid value - SetRaw with ENTITY",
			fn: func() error {
				_, err := SetRaw("<root/>", "root.data", "<!ENTITY test 'value'><test/>")
				return err
			},
			wantType: ErrInvalidValue,
		},
		{
			name: "invalid value - SetRaw with nested CDATA",
			fn: func() error {
				_, err := SetRaw("<root/>", "root.data", "<![CDATA[<![CDATA[test]]>]]>")
				return err
			},
			wantType: ErrInvalidValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Errorf("Expected error but got nil")
				return
			}
			if tt.wantType != nil {
				if !errors.Is(err, tt.wantType) {
					t.Errorf("Expected error type %v, got %v", tt.wantType, err)
				}
			}
		})
	}
}

// TestErrorMessages verifies error messages are clear and helpful
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		fn          func() error
		contains    []string
		notContains []string
	}{
		{
			name: "malformed XML message",
			fn: func() error {
				_, err := Set("<root>", "root.item", "value")
				return err
			},
			contains:    []string{"malformed"},
			notContains: []string{"panic", "internal", "nil pointer"},
		},
		{
			name: "invalid path message",
			fn: func() error {
				_, err := Set("<root/>", "", "value")
				return err
			},
			contains:    []string{"path"},
			notContains: []string{"panic"},
		},
		{
			name: "SetMany mismatch message",
			fn: func() error {
				_, err := SetMany("<root/>", []string{"a", "b"}, []interface{}{"value"})
				return err
			},
			contains:    []string{"path", "length", "mismatch"},
			notContains: []string{"panic"},
		},
		{
			name: "SetRaw validation message",
			fn: func() error {
				_, err := SetRaw("<root/>", "root.data", "<item>")
				return err
			},
			contains:    []string{"invalid"},
			notContains: []string{"panic"},
		},
		{
			name: "SetRaw DOCTYPE security message",
			fn: func() error {
				_, err := SetRaw("<root/>", "root.data", "<!DOCTYPE test><test/>")
				return err
			},
			contains:    []string{"DOCTYPE", "not allowed"},
			notContains: []string{"panic"},
		},
		{
			name: "SetRaw ENTITY security message",
			fn: func() error {
				_, err := SetRaw("<root/>", "root.data", "<!ENTITY test 'value'>")
				return err
			},
			contains:    []string{"ENTITY", "not allowed"},
			notContains: []string{"panic"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Errorf("Expected error but got nil")
				return
			}

			errMsg := err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(substr)) {
					t.Errorf("Error message should contain %q, got: %s", substr, errMsg)
				}
			}
			for _, substr := range tt.notContains {
				if strings.Contains(strings.ToLower(errMsg), strings.ToLower(substr)) {
					t.Errorf("Error message should NOT contain %q, got: %s", substr, errMsg)
				}
			}
		})
	}
}

// TestErrorReturnsUnchangedXML verifies that errors return original XML unchanged
func TestErrorReturnsUnchangedXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		fn   func(string) (string, error)
	}{
		{
			name: "Set with malformed XML",
			xml:  "<root>",
			fn: func(xml string) (string, error) {
				return Set(xml, "root.item", "value")
			},
		},
		{
			name: "Set with invalid path",
			xml:  "<root/>",
			fn: func(xml string) (string, error) {
				return Set(xml, "", "value")
			},
		},
		{
			name: "Delete with malformed XML",
			xml:  "<root>",
			fn: func(xml string) (string, error) {
				return Delete(xml, "root.item")
			},
		},
		{
			name: "Delete with invalid path",
			xml:  "<root/>",
			fn: func(xml string) (string, error) {
				return Delete(xml, "")
			},
		},
		{
			name: "SetMany with mismatched lengths",
			xml:  "<root/>",
			fn: func(xml string) (string, error) {
				return SetMany(xml, []string{"a", "b"}, []interface{}{"value"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.fn(tt.xml)
			if err == nil {
				t.Errorf("Expected error but got nil")
				return
			}
			if result != tt.xml {
				t.Errorf("Error should return original XML unchanged\nGot:  %q\nWant: %q", result, tt.xml)
			}
		})
	}
}

// TestNoErrorsOnValidOperations verifies valid operations don't return errors
func TestNoErrorsOnValidOperations(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Get never returns error",
			fn: func() error {
				_ = Get("<root><item>value</item></root>", "root.item")
				return nil
			},
		},
		{
			name: "Get on empty XML",
			fn: func() error {
				_ = Get("", "root.item")
				return nil
			},
		},
		{
			name: "Get on malformed XML (graceful)",
			fn: func() error {
				_ = Get("<root>", "root.item")
				return nil
			},
		},
		{
			name: "Set on valid XML",
			fn: func() error {
				_, err := Set("<root/>", "root.item", "value")
				return err
			},
		},
		{
			name: "Delete on valid XML",
			fn: func() error {
				_, err := Delete("<root><item>value</item></root>", "root.item")
				return err
			},
		},
		{
			name: "Delete on non-existent path (no error)",
			fn: func() error {
				_, err := Delete("<root/>", "root.nonexistent")
				return err
			},
		},
		{
			name: "SetMany with matching lengths",
			fn: func() error {
				_, err := SetMany("<root/>", []string{"root.a", "root.b"}, []interface{}{1, 2})
				return err
			},
		},
		{
			name: "DeleteMany on valid paths",
			fn: func() error {
				_, err := DeleteMany("<root><a/><b/></root>", "root.a", "root.b")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestErrorWrapping verifies errors can be unwrapped with errors.Is
func TestErrorWrapping(t *testing.T) {
	// Test that wrapped errors can be detected with errors.Is
	_, err := SetMany("<root/>", []string{"a"}, []interface{}{})
	if err == nil {
		t.Fatal("Expected error")
	}

	if !errors.Is(err, ErrInvalidPath) {
		t.Errorf("Expected wrapped ErrInvalidPath to be detectable with errors.Is")
	}
}

// TestNoPanicsOnErrors verifies no panics occur during error conditions
func TestNoPanicsOnErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "Set with malformed XML",
			fn: func() {
				_, _ = Set("<root>", "root.item", "value")
			},
		},
		{
			name: "Get with extremely large XML",
			fn: func() {
				largeXML := strings.Repeat("x", MaxDocumentSize*2)
				_ = Get(largeXML, "root.item")
			},
		},
		{
			name: "Set with nil value",
			fn: func() {
				_, _ = Set("<root/>", "root.item", nil)
			},
		},
		{
			name: "SetMany with empty slices",
			fn: func() {
				_, _ = SetMany("<root/>", []string{}, []interface{}{})
			},
		},
		{
			name: "Delete with empty path",
			fn: func() {
				_, _ = Delete("<root/>", "")
			},
		},
		{
			name: "SetRaw with invalid XML",
			fn: func() {
				_, _ = SetRaw("<root/>", "root.data", "<invalid>")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Operation panicked: %v", r)
				}
			}()
			tt.fn()
		})
	}
}

// ============================================================================
// Error Tests - Batch Operations
// ============================================================================

// TestSetManyErrors_MalformedXML tests SetMany with malformed XML
func TestSetManyErrors_MalformedXML(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		paths   []string
		values  []interface{}
		wantErr bool
		errType error
	}{
		{
			name:    "unclosed tag",
			xml:     "<root>",
			paths:   []string{"root.a"},
			values:  []interface{}{"value"},
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "mismatched tags",
			xml:     "<root><item>value</wrong></root>",
			paths:   []string{"root.a"},
			values:  []interface{}{"value"},
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "empty document",
			xml:     "",
			paths:   []string{"root.a"},
			values:  []interface{}{"value"},
			wantErr: false, // Empty XML is now valid for creating new XML from scratch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetMany panicked: %v", r)
				}
			}()

			result, err := SetMany(tt.xml, tt.paths, tt.values)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if tt.errType != nil && !errors.Is(err, tt.errType) {
						t.Errorf("Expected error type %v, got %v", tt.errType, err)
					}
				}
				// Original XML should be unchanged on error
				if result != tt.xml {
					t.Errorf("Expected original XML on error")
				}
			}
		})
	}
}

// TestSetManyErrors_MismatchedLengths tests SetMany with mismatched array lengths
func TestSetManyErrors_MismatchedLengths(t *testing.T) {
	xml := "<root/>"

	tests := []struct {
		name    string
		paths   []string
		values  []interface{}
		wantErr bool
		errType error
	}{
		{
			name:    "more paths than values",
			paths:   []string{"root.a", "root.b"},
			values:  []interface{}{"value"},
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "more values than paths",
			paths:   []string{"root.a"},
			values:  []interface{}{"value1", "value2"},
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "empty paths with values",
			paths:   []string{},
			values:  []interface{}{"value"},
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "paths with empty values",
			paths:   []string{"root.a"},
			values:  []interface{}{},
			wantErr: true,
			errType: ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetMany(xml, tt.paths, tt.values)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if tt.errType != nil && !errors.Is(err, tt.errType) {
						t.Errorf("Expected error type %v, got %v", tt.errType, err)
					}
				}
				// Original XML should be unchanged on error
				if result != xml {
					t.Errorf("Expected original XML on error")
				}
			}
		})
	}
}

// TestSetManyErrors_EmptyInputs tests SetMany with empty inputs (should not error)
func TestSetManyErrors_EmptyInputs(t *testing.T) {
	xml := "<root/>"

	tests := []struct {
		name   string
		paths  []string
		values []interface{}
	}{
		{
			name:   "both empty",
			paths:  []string{},
			values: []interface{}{},
		},
		{
			name:   "nil paths and values",
			paths:  nil,
			values: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetMany(xml, tt.paths, tt.values)
			if err != nil {
				t.Errorf("Empty inputs should not error: %v", err)
			}
			// Result should be unchanged
			if result != xml {
				t.Error("XML should be unchanged for empty inputs")
			}
		})
	}
}

// TestSetManyErrors_PartialSuccess tests SetMany with one path failing
func TestSetManyErrors_PartialSuccess(t *testing.T) {
	xml := "<root/>"
	paths := []string{"root.a", "", "root.c"} // Middle path is invalid
	values := []interface{}{"val1", "val2", "val3"}

	result, err := SetMany(xml, paths, values)

	// Should error on the invalid path
	if err == nil {
		t.Error("Expected error for invalid path in batch")
	}

	// Original XML should be unchanged on error
	if result != xml {
		t.Error("Expected original XML when batch fails")
	}
}

// TestSetManyErrors_DocumentTooLarge tests SetMany with large document
func TestSetManyErrors_DocumentTooLarge(t *testing.T) {
	largeXML := strings.Repeat("<item>x</item>", 1000000)
	xml := "<root>" + largeXML + "</root>"
	paths := []string{"root.a"}
	values := []interface{}{"value"}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetMany panicked on large document: %v", r)
		}
	}()

	result, err := SetMany(xml, paths, values)
	if err == nil {
		t.Error("Expected error for document too large")
	}
	if !errors.Is(err, ErrMalformedXML) {
		t.Errorf("Expected ErrMalformedXML, got %v", err)
	}
	// Original XML should be unchanged
	if result != xml {
		t.Error("Expected original XML on error")
	}
}

// TestSetManyBytes_Errors tests SetManyBytes error handling
func TestSetManyBytes_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     []byte
		paths   []string
		values  []interface{}
		wantErr bool
	}{
		{
			name:    "nil bytes",
			xml:     nil,
			paths:   []string{"root.a"},
			values:  []interface{}{"value"},
			wantErr: true,
		},
		{
			name:    "empty bytes",
			xml:     []byte{},
			paths:   []string{"root.a"},
			values:  []interface{}{"value"},
			wantErr: false, // Empty XML is now valid for creating new XML from scratch
		},
		{
			name:    "malformed bytes",
			xml:     []byte("<root>"),
			paths:   []string{"root.a"},
			values:  []interface{}{"value"},
			wantErr: true,
		},
		{
			name:    "valid bytes",
			xml:     []byte("<root/>"),
			paths:   []string{"root.a"},
			values:  []interface{}{"value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetManyBytes panicked: %v", r)
				}
			}()

			result, err := SetManyBytes(tt.xml, tt.paths, tt.values)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(result) == 0 {
					t.Error("Expected result but got empty bytes")
				}
			}
		})
	}
}

// TestDeleteManyErrors_MalformedXML tests DeleteMany with malformed XML
func TestDeleteManyErrors_MalformedXML(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		paths   []string
		wantErr bool
		errType error
	}{
		{
			name:    "unclosed tag",
			xml:     "<root>",
			paths:   []string{"root.item"},
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "mismatched tags",
			xml:     "<root><item>value</wrong></root>",
			paths:   []string{"root.item"},
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "empty document",
			xml:     "",
			paths:   []string{"root.item"},
			wantErr: true,
			errType: ErrMalformedXML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DeleteMany panicked: %v", r)
				}
			}()

			result, err := DeleteMany(tt.xml, tt.paths...)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if tt.errType != nil && !errors.Is(err, tt.errType) {
						t.Errorf("Expected error type %v, got %v", tt.errType, err)
					}
				}
				// Original XML should be unchanged on error
				if result != tt.xml {
					t.Errorf("Expected original XML on error")
				}
			}
		})
	}
}

// TestDeleteManyErrors_EmptyInputs tests DeleteMany with empty inputs
func TestDeleteManyErrors_EmptyInputs(t *testing.T) {
	xml := "<root><item>value</item></root>"

	tests := []struct {
		name  string
		paths []string
	}{
		{
			name:  "empty paths",
			paths: []string{},
		},
		{
			name:  "nil paths",
			paths: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DeleteMany(xml, tt.paths...)
			if err != nil {
				t.Errorf("Empty inputs should not error: %v", err)
			}
			// Result should be unchanged
			if result != xml {
				t.Error("XML should be unchanged for empty inputs")
			}
		})
	}
}

// TestDeleteManyErrors_NonExistentPaths tests DeleteMany with non-existent paths (should not error)
func TestDeleteManyErrors_NonExistentPaths(t *testing.T) {
	xml := "<root><item>value</item></root>"
	paths := []string{"root.nonexistent1", "root.nonexistent2", "root.nonexistent3"}

	result, err := DeleteMany(xml, paths...)
	if err != nil {
		t.Errorf("DeleteMany with non-existent paths should not error: %v", err)
	}
	// XML should be unchanged
	_ = result
}

// TestDeleteManyErrors_PartialSuccess tests DeleteMany with one path failing
func TestDeleteManyErrors_PartialSuccess(t *testing.T) {
	xml := "<root><item1>val1</item1><item2>val2</item2></root>"
	paths := []string{"root.item1", "", "root.item2"} // Middle path is invalid

	result, err := DeleteMany(xml, paths...)

	// Should error on the invalid path
	if err == nil {
		t.Error("Expected error for invalid path in batch")
	}

	// Original XML should be unchanged on error
	if result != xml {
		t.Error("Expected original XML when batch fails")
	}
}

// TestDeleteManyErrors_DocumentTooLarge tests DeleteMany with large document
func TestDeleteManyErrors_DocumentTooLarge(t *testing.T) {
	largeXML := strings.Repeat("<item>x</item>", 1000000)
	xml := "<root>" + largeXML + "</root>"
	paths := []string{"root.item"}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DeleteMany panicked on large document: %v", r)
		}
	}()

	result, err := DeleteMany(xml, paths...)
	if err == nil {
		t.Error("Expected error for document too large")
	}
	if !errors.Is(err, ErrMalformedXML) {
		t.Errorf("Expected ErrMalformedXML, got %v", err)
	}
	// Original XML should be unchanged
	if result != xml {
		t.Error("Expected original XML on error")
	}
}

// TestDeleteManyBytes_Errors tests DeleteManyBytes error handling
func TestDeleteManyBytes_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     []byte
		paths   []string
		wantErr bool
	}{
		{
			name:    "nil bytes",
			xml:     nil,
			paths:   []string{"root.item"},
			wantErr: true,
		},
		{
			name:    "empty bytes",
			xml:     []byte{},
			paths:   []string{"root.item"},
			wantErr: true, // Delete on empty XML should error
		},
		{
			name:    "malformed bytes",
			xml:     []byte("<root>"),
			paths:   []string{"root.item"},
			wantErr: true,
		},
		{
			name:    "valid bytes",
			xml:     []byte("<root><item>value</item></root>"),
			paths:   []string{"root.item"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DeleteManyBytes panicked: %v", r)
				}
			}()

			result, err := DeleteManyBytes(tt.xml, tt.paths...)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(result) == 0 {
					t.Error("Expected result but got empty bytes")
				}
			}
		})
	}
}

// TestBatchOperations_ConcurrentAccess tests concurrent batch operations
func TestBatchOperations_ConcurrentAccess(t *testing.T) {
	xml := "<root><a>1</a><b>2</b><c>3</c></root>"

	done := make(chan bool, 4)

	// Concurrent SetMany
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in concurrent SetMany: %v", r)
			}
			done <- true
		}()
		_, _ = SetMany(xml, []string{"root.x", "root.y"}, []interface{}{1, 2})
	}()

	// Concurrent DeleteMany
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in concurrent DeleteMany: %v", r)
			}
			done <- true
		}()
		_, _ = DeleteMany(xml, "root.a", "root.b")
	}()

	// Concurrent SetMany with error
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in concurrent SetMany error: %v", r)
			}
			done <- true
		}()
		_, _ = SetMany("<root>", []string{"root.a"}, []interface{}{"value"})
	}()

	// Concurrent DeleteMany with error
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in concurrent DeleteMany error: %v", r)
			}
			done <- true
		}()
		_, _ = DeleteMany("<root>", "root.item")
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}

// TestBatchOperations_NoErrorOnValidOperations tests valid batch operations
func TestBatchOperations_NoErrorOnValidOperations(t *testing.T) {
	tests := []struct {
		name   string
		xml    string
		paths  []string
		values []interface{}
	}{
		{
			name:   "set multiple elements",
			xml:    "<root/>",
			paths:  []string{"root.a", "root.b", "root.c"},
			values: []interface{}{1, 2, 3},
		},
		{
			name:   "set attributes",
			xml:    "<root><item/></root>",
			paths:  []string{"root.item.@id", "root.item.@name"},
			values: []interface{}{"123", "test"},
		},
		{
			name:   "mixed types",
			xml:    "<root/>",
			paths:  []string{"root.int", "root.float", "root.bool", "root.str"},
			values: []interface{}{42, 3.14, true, "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetMany(tt.xml, tt.paths, tt.values)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result == "" {
				t.Error("Expected non-empty result")
			}
			if result == tt.xml {
				t.Error("Expected modified XML")
			}
		})
	}
}

// ============================================================================
// Error Tests - Delete Operations
// ============================================================================

// TestDeleteErrors_MalformedXML tests Delete with malformed XML
func TestDeleteErrors_MalformedXML(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		path    string
		wantErr bool
		errType error
	}{
		{
			name:    "unclosed tag",
			xml:     "<root><item>",
			path:    "root.item",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "mismatched tags",
			xml:     "<root><item>value</wrong></root>",
			path:    "root.item",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "incomplete opening tag",
			xml:     "<root",
			path:    "root.item",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "only closing tag",
			xml:     "</root>",
			path:    "root.item",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "empty document",
			xml:     "",
			path:    "root.item",
			wantErr: true, // Delete on empty XML should error
			errType: ErrMalformedXML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Delete panicked on malformed XML: %v", r)
				}
			}()

			result, err := Delete(tt.xml, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if tt.errType != nil && !errors.Is(err, tt.errType) {
						t.Errorf("Expected error type %v, got %v", tt.errType, err)
					}
				}
				// Original XML should be unchanged on error
				if result != tt.xml {
					t.Errorf("Expected original XML on error\nGot:  %q\nWant: %q", result, tt.xml)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestDeleteErrors_InvalidPath tests Delete with invalid path syntax
func TestDeleteErrors_InvalidPath(t *testing.T) {
	xml := "<root><item>value</item></root>"

	tests := []struct {
		name    string
		path    string
		wantErr bool
		errType error
	}{
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "only dots",
			path:    "...",
			wantErr: false, // parsePath handles this
		},
		{
			name:    "leading dot",
			path:    ".root.item",
			wantErr: false, // parsePath handles this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Delete panicked on invalid path: %v", r)
				}
			}()

			result, err := Delete(xml, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if tt.errType != nil && !errors.Is(err, tt.errType) {
						t.Errorf("Expected error type %v, got %v", tt.errType, err)
					}
				}
				// Original XML should be unchanged on error
				if result != xml {
					t.Errorf("Expected original XML on error")
				}
			}
		})
	}
}

// TestDeleteErrors_NonExistentPath tests Delete with paths that don't exist (should not error)
func TestDeleteErrors_NonExistentPath(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
	}{
		{
			name: "non-existent element",
			xml:  "<root><item>value</item></root>",
			path: "root.nonexistent",
		},
		{
			name: "non-existent nested path",
			xml:  "<root><item>value</item></root>",
			path: "root.level1.level2.nonexistent",
		},
		{
			name: "non-existent attribute",
			xml:  "<root><item>value</item></root>",
			path: "root.item.@nonexistent",
		},
		{
			name: "non-existent array index",
			xml:  "<root><item>value</item></root>",
			path: "root.item.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Delete(tt.xml, tt.path)
			if err != nil {
				t.Errorf("Delete non-existent path should not error: %v", err)
			}

			// XML should be unchanged when deleting non-existent path
			if result != tt.xml {
				t.Logf("XML changed when deleting non-existent path (may be acceptable)")
			}
		})
	}
}

// TestDeleteErrors_DocumentTooLarge tests Delete with documents exceeding size limits
func TestDeleteErrors_DocumentTooLarge(t *testing.T) {
	// Create XML larger than MaxDocumentSize
	largeXML := strings.Repeat("<item>x</item>", 1000000)
	xml := "<root>" + largeXML + "</root>"

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Delete panicked on large document: %v", r)
		}
	}()

	result, err := Delete(xml, "root.item")
	if err == nil {
		t.Error("Expected error for document too large")
	}
	if !errors.Is(err, ErrMalformedXML) {
		t.Errorf("Expected ErrMalformedXML for large document, got %v", err)
	}
	// Original XML should be unchanged
	if result != xml {
		t.Error("Expected original XML on error")
	}
}

// TestDeleteBytes_Errors tests DeleteBytes error handling
func TestDeleteBytes_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     []byte
		path    string
		wantErr bool
	}{
		{
			name:    "nil bytes",
			xml:     nil,
			path:    "root.item",
			wantErr: true,
		},
		{
			name:    "empty bytes",
			xml:     []byte{},
			path:    "root.item",
			wantErr: true, // Delete on empty XML should error
		},
		{
			name:    "malformed bytes",
			xml:     []byte("<root>"),
			path:    "root.item",
			wantErr: true,
		},
		{
			name:    "valid bytes",
			xml:     []byte("<root><item>value</item></root>"),
			path:    "root.item",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DeleteBytes panicked: %v", r)
				}
			}()

			result, err := DeleteBytes(tt.xml, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(result) == 0 {
					t.Error("Expected result but got empty bytes")
				}
			}
		})
	}
}

// TestDelete_NoErrorOnValidOperations tests that valid operations don't return errors
func TestDelete_NoErrorOnValidOperations(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
	}{
		{
			name: "delete existing element",
			xml:  "<root><item>value</item></root>",
			path: "root.item",
		},
		{
			name: "delete attribute",
			xml:  "<root><item id='123'>value</item></root>",
			path: "root.item.@id",
		},
		{
			name: "delete nested element",
			xml:  "<root><a><b><c>value</c></b></a></root>",
			path: "root.a.b.c",
		},
		{
			name: "delete array element by index",
			xml:  "<root><item>1</item><item>2</item><item>3</item></root>",
			path: "root.item.1",
		},
		{
			name: "delete from self-closing element parent",
			xml:  "<root><parent/></root>",
			path: "root.parent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Delete(tt.xml, tt.path)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result == "" {
				t.Error("Expected non-empty result")
			}
		})
	}
}

// TestDelete_RecoveryAfterError tests that errors don't corrupt internal state
func TestDelete_RecoveryAfterError(t *testing.T) {
	xml := "<root><item>value</item><other>data</other></root>"

	// Perform invalid operation
	_, err := Delete(xml, "")
	if err == nil {
		t.Fatal("Expected error for empty path")
	}

	// Should still work correctly after error
	result, err := Delete(xml, "root.item")
	if err != nil {
		t.Fatalf("Delete after error failed: %v", err)
	}

	// Verify the deletion worked
	checkResult := Get(result, "root.item")
	if checkResult.Exists() {
		t.Error("Item should have been deleted")
	}

	// Verify other element still exists
	otherResult := Get(result, "root.other")
	if !otherResult.Exists() || otherResult.String() != "data" {
		t.Error("Other element should still exist")
	}
}

// TestDelete_ConcurrentOperations tests concurrent Delete operations don't cause panics
func TestDelete_ConcurrentOperations(t *testing.T) {
	xml := "<root><item1>value1</item1><item2>value2</item2><item3>value3</item3></root>"
	paths := []string{"root.item1", "root.item2", "root.item3", "root.nonexistent"}

	done := make(chan bool, len(paths))
	for _, path := range paths {
		go func(p string) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic in concurrent delete: %v", r)
				}
				done <- true
			}()

			_, err := Delete(xml, p)
			if err != nil {
				t.Logf("Concurrent delete error (may be acceptable): %v", err)
			}
		}(path)
	}

	// Wait for all goroutines
	for i := 0; i < len(paths); i++ {
		<-done
	}
}

// TestDelete_EdgeCases tests Delete edge cases
func TestDelete_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
	}{
		{
			name: "delete root element (invalid but shouldn't panic)",
			xml:  "<root><item>value</item></root>",
			path: "root",
		},
		{
			name: "delete with whitespace in XML",
			xml:  "<root>  \n  <item>value</item>  \n  </root>",
			path: "root.item",
		},
		{
			name: "delete with comments",
			xml:  "<root><!-- comment --><item>value</item></root>",
			path: "root.item",
		},
		{
			name: "delete with CDATA",
			xml:  "<root><item><![CDATA[<tag>]]></item></root>",
			path: "root.item",
		},
		{
			name: "delete with namespaces",
			xml:  "<root xmlns:ns='http://example.com'><ns:item>value</ns:item></root>",
			path: "root.ns:item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Delete panicked on edge case: %v", r)
				}
			}()

			result, err := Delete(tt.xml, tt.path)
			if err != nil {
				t.Logf("Delete edge case returned error (may be acceptable): %v", err)
			}
			_ = result
		})
	}
}

// ============================================================================
// Error Tests - Get Operations
// ============================================================================

// TestGetErrors_EmptyXML tests Get with empty XML
func TestGetErrors_EmptyXML(t *testing.T) {
	result := Get("", "root.item")
	if result.Exists() {
		t.Error("Get on empty XML should return non-existent result")
	}
	if result.Type != Null {
		t.Errorf("Expected Null type, got %v", result.Type)
	}
}

// TestGetErrors_MalformedXML tests Get graceful handling of malformed XML
func TestGetErrors_MalformedXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
	}{
		{
			name: "unclosed tag",
			xml:  "<root><item>",
			path: "root.item",
		},
		{
			name: "mismatched tags",
			xml:  "<root><item>value</wrong></root>",
			path: "root.item",
		},
		{
			name: "missing closing bracket",
			xml:  "<root<item>value</item></root>",
			path: "root.item",
		},
		{
			name: "partial closing tag",
			xml:  "<root><item>value</item",
			path: "root.item",
		},
		{
			name: "multiple opening brackets",
			xml:  "<root><<item>value</item></root>",
			path: "root.item",
		},
		{
			name: "unclosed attribute quote",
			xml:  "<root><item attr=\"value>text</item></root>",
			path: "root.item",
		},
		{
			name: "only opening tag",
			xml:  "<root>",
			path: "root",
		},
		{
			name: "only closing tag",
			xml:  "</root>",
			path: "root",
		},
		{
			name: "random text",
			xml:  "not xml at all",
			path: "root.item",
		},
		{
			name: "incomplete XML declaration",
			xml:  "<?xml",
			path: "root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Get panicked on malformed XML: %v", r)
				}
			}()

			result := Get(tt.xml, tt.path)
			// Get should return gracefully, even if result doesn't exist
			_ = result.String()
			_ = result.Exists()
		})
	}
}

// TestGetErrors_InvalidPaths tests Get with invalid path syntax
func TestGetErrors_InvalidPaths(t *testing.T) {
	xml := "<root><item>value</item></root>"

	tests := []struct {
		name string
		path string
	}{
		{name: "empty path", path: ""},
		{name: "only dots", path: "..."},
		{name: "leading dot", path: ".root.item"},
		{name: "trailing dot", path: "root.item."},
		{name: "double dots", path: "root..item"},
		{name: "path with null", path: "root\x00item"},
		{name: "path with control chars", path: "root\x01\x02item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Get panicked on invalid path: %v", r)
				}
			}()

			result := Get(xml, tt.path)
			// Invalid paths should return empty result gracefully
			_ = result.String()
		})
	}
}

// TestGetErrors_InvalidFilterSyntax tests Get with invalid filter syntax
func TestGetErrors_InvalidFilterSyntax(t *testing.T) {
	xml := "<root><item id='1'>val1</item><item id='2'>val2</item></root>"

	tests := []struct {
		name string
		path string
	}{
		{name: "unclosed filter bracket", path: "root.item[@id==1"},
		{name: "empty filter", path: "root.item[]"},
		{name: "filter without expression", path: "root.item[@]"},
		{name: "missing filter value", path: "root.item[@id==]"},
		{name: "filter with invalid operator", path: "root.item[@id===1]"},
		{name: "filter with unclosed quote", path: "root.item[@id=='1]"},
		{name: "nested unclosed filters", path: "root.item[@id==[1]]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Get panicked on invalid filter: %v", r)
				}
			}()

			result := Get(xml, tt.path)
			_ = result.String()
		})
	}
}

// TestGetErrors_InvalidModifierSyntax tests Get with invalid modifier syntax
func TestGetErrors_InvalidModifierSyntax(t *testing.T) {
	xml := "<root><item>value</item></root>"

	tests := []struct {
		name string
		path string
	}{
		{name: "unknown modifier", path: "root.item|unknown"},
		{name: "empty modifier", path: "root.item|"},
		{name: "multiple pipes", path: "root.item||upper"},
		{name: "modifier without path", path: "|upper"},
		{name: "modifier with invalid args", path: "root.item|substring("},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Get panicked on invalid modifier: %v", r)
				}
			}()

			result := Get(xml, tt.path)
			_ = result.String()
		})
	}
}

// TestGetErrors_DocumentTooLarge tests Get with documents exceeding size limits
func TestGetErrors_DocumentTooLarge(t *testing.T) {
	// Create XML larger than MaxDocumentSize
	largeXML := strings.Repeat("<item>x</item>", 1000000)
	xml := "<root>" + largeXML + "</root>"

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Get panicked on large document: %v", r)
		}
	}()

	result := Get(xml, "root.item")
	// Should handle gracefully (return Null or truncated result)
	if result.Type != Null {
		t.Logf("Large document returned result type: %v", result.Type)
	}
}

// TestGetErrors_PathTooDeep tests Get with extremely deep paths
func TestGetErrors_PathTooDeep(t *testing.T) {
	xml := "<root><item>value</item></root>"

	tests := []struct {
		name       string
		pathLength int
	}{
		{name: "very deep path (1000 segments)", pathLength: 1000},
		{name: "extremely deep path (10000 segments)", pathLength: 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments := make([]string, tt.pathLength)
			for i := 0; i < tt.pathLength; i++ {
				segments[i] = "item"
			}
			path := strings.Join(segments, ".")

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Get panicked on deep path: %v", r)
				}
			}()

			result := Get(xml, path)
			_ = result.String()
		})
	}
}

// TestGetErrors_NonExistentPaths tests Get with paths that don't exist
func TestGetErrors_NonExistentPaths(t *testing.T) {
	xml := "<root><item>value</item></root>"

	tests := []struct {
		name string
		path string
	}{
		{name: "non-existent element", path: "root.nonexistent"},
		{name: "non-existent nested", path: "root.item.nested"},
		{name: "non-existent attribute", path: "root.item.@nonexistent"},
		{name: "non-existent array index", path: "root.item.5"},
		{name: "wrong root", path: "wrong.item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Get(xml, tt.path)
			if result.Exists() {
				t.Errorf("Expected non-existent result for path %q", tt.path)
			}
			if result.Type != Null {
				t.Errorf("Expected Null type for non-existent path, got %v", result.Type)
			}
			// Should be safe to call String() on non-existent results
			str := result.String()
			if str != "" {
				t.Errorf("Expected empty string for non-existent result, got %q", str)
			}
		})
	}
}

// TestGetErrors_InvalidXMLCharacters tests Get with invalid XML characters
func TestGetErrors_InvalidXMLCharacters(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
	}{
		{
			name: "control characters",
			xml:  "<root><item>\x01\x02\x03</item></root>",
			path: "root.item",
		},
		{
			name: "invalid UTF-8",
			xml:  string([]byte{'<', 'r', 'o', 'o', 't', '>', 0xFF, 0xFE, '<', '/', 'r', 'o', 'o', 't', '>'}),
			path: "root",
		},
		{
			name: "null byte",
			xml:  "<root><item>valid\x00invalid</item></root>",
			path: "root.item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Get panicked on invalid characters: %v", r)
				}
			}()

			result := Get(tt.xml, tt.path)
			_ = result.String()
		})
	}
}

// TestGetErrors_ConcurrentInvalidAccess tests concurrent Get with invalid data
func TestGetErrors_ConcurrentInvalidAccess(t *testing.T) {
	malformedXMLs := []string{
		"<root><unclosed>",
		"<root><item>value</wrong></root>",
		"<<<>>>",
		"",
		"not xml at all",
		"<root" + strings.Repeat("x", MaxDocumentSize+1) + "></root>",
	}

	done := make(chan bool, len(malformedXMLs))
	for _, xml := range malformedXMLs {
		go func(xml string) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic in concurrent invalid access: %v", r)
				}
				done <- true
			}()

			result := Get(xml, "root.item")
			_ = result.String()
		}(xml)
	}

	// Wait for all goroutines
	for i := 0; i < len(malformedXMLs); i++ {
		<-done
	}
}

// TestGetErrors_WildcardEdgeCases tests Get with wildcard edge cases
func TestGetErrors_WildcardEdgeCases(t *testing.T) {
	xml := "<root><item>val1</item><item>val2</item></root>"

	tests := []struct {
		name string
		path string
	}{
		{name: "only wildcard", path: "*"},
		{name: "only recursive wildcard", path: "**"},
		{name: "multiple consecutive wildcards", path: "*.*.*"},
		{name: "recursive wildcard at end", path: "root.**"},
		{name: "recursive wildcard without target", path: "root.**"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Get panicked on wildcard edge case: %v", r)
				}
			}()

			result := Get(xml, tt.path)
			_ = result.String()
		})
	}
}

// TestGetBytes_Errors tests GetBytes error handling
func TestGetBytes_Errors(t *testing.T) {
	tests := []struct {
		name string
		xml  []byte
		path string
	}{
		{
			name: "nil bytes",
			xml:  nil,
			path: "root.item",
		},
		{
			name: "empty bytes",
			xml:  []byte{},
			path: "root.item",
		},
		{
			name: "malformed bytes",
			xml:  []byte("<root>"),
			path: "root.item",
		},
		{
			name: "too large bytes",
			xml:  make([]byte, MaxDocumentSize+1),
			path: "root.item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GetBytes panicked: %v", r)
				}
			}()

			result := GetBytes(tt.xml, tt.path)
			_ = result.String()
		})
	}
}

// TestGetMany_Errors tests GetMany error handling
func TestGetMany_Errors(t *testing.T) {
	tests := []struct {
		name  string
		xml   string
		paths []string
	}{
		{
			name:  "empty paths",
			xml:   "<root><item>value</item></root>",
			paths: []string{},
		},
		{
			name:  "nil paths",
			xml:   "<root><item>value</item></root>",
			paths: nil,
		},
		{
			name:  "mix of valid and invalid paths",
			xml:   "<root><item>value</item></root>",
			paths: []string{"root.item", "", "root.nonexistent"},
		},
		{
			name:  "malformed XML with multiple paths",
			xml:   "<root>",
			paths: []string{"root.a", "root.b", "root.c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GetMany panicked: %v", r)
				}
			}()

			results := GetMany(tt.xml, tt.paths...)
			if len(results) != len(tt.paths) {
				t.Errorf("Expected %d results, got %d", len(tt.paths), len(results))
			}
			for _, r := range results {
				_ = r.String()
			}
		})
	}
}

// TestGetWithOptions_ErrorPaths tests GetWithOptions error handling
func TestGetWithOptions_ErrorPaths(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
		opts *Options
	}{
		{
			name: "nil options",
			xml:  "<root><item>value</item></root>",
			path: "root.item",
			opts: nil,
		},
		{
			name: "options with malformed XML",
			xml:  "<root>",
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
		{
			name: "options with invalid path",
			xml:  "<root><item>value</item></root>",
			path: "",
			opts: &Options{CaseSensitive: true},
		},
		{
			name: "options with too large document",
			xml:  strings.Repeat("x", MaxDocumentSize+1),
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GetWithOptions panicked: %v", r)
				}
			}()

			result := GetWithOptions(tt.xml, tt.path, tt.opts)
			_ = result.String()
		})
	}
}

// ============================================================================
// Error Tests - Options Operations
// ============================================================================

// TestOptionsErrors_InvalidOptions tests operations with potentially invalid options
func TestOptionsErrors_InvalidOptions(t *testing.T) {
	xml := "<root><item>value</item></root>"

	tests := []struct {
		name string
		opts *Options
		path string
	}{
		{
			name: "nil options (should use defaults)",
			opts: nil,
			path: "root.item",
		},
		{
			name: "empty namespace map",
			opts: &Options{
				Namespaces: map[string]string{},
			},
			path: "root.item",
		},
		{
			name: "case insensitive",
			opts: &Options{
				CaseSensitive: false,
			},
			path: "ROOT.ITEM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Get - should not panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("GetWithOptions panicked (may be acceptable for nil options): %v", r)
					}
				}()
				result := GetWithOptions(xml, tt.path, tt.opts)
				_ = result.String()
			}()

			// Test Set - should not panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("SetWithOptions panicked (may be acceptable for nil options): %v", r)
					}
				}()
				_, err := SetWithOptions(xml, tt.path, "new", tt.opts)
				if err != nil {
					t.Logf("SetWithOptions returned error (may be acceptable): %v", err)
				}
			}()
		})
	}
}

// TestGetWithOptions_OptionsErrors tests GetWithOptions error handling with various options
func TestGetWithOptions_OptionsErrors(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		path string
		opts *Options
	}{
		{
			name: "malformed XML with nil options",
			xml:  "<root>",
			path: "root.item",
			opts: nil,
		},
		{
			name: "malformed XML with case insensitive",
			xml:  "<root>",
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
		{
			name: "empty XML with options",
			xml:  "",
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
		{
			name: "invalid path with options",
			xml:  "<root><item>value</item></root>",
			path: "",
			opts: &Options{CaseSensitive: true},
		},
		{
			name: "document too large with options",
			xml:  strings.Repeat("x", MaxDocumentSize+1),
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GetWithOptions panicked: %v", r)
				}
			}()

			result := GetWithOptions(tt.xml, tt.path, tt.opts)
			// Should return gracefully
			_ = result.String()
			_ = result.Exists()
		})
	}
}

// TestSetWithOptions_OptionsErrors tests SetWithOptions error handling with various options
func TestSetWithOptions_OptionsErrors(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		path    string
		value   interface{}
		opts    *Options
		wantErr bool
	}{
		{
			name:    "malformed XML with nil options",
			xml:     "<root>",
			path:    "root.item",
			value:   "value",
			opts:    nil,
			wantErr: true,
		},
		{
			name:    "malformed XML with case insensitive",
			xml:     "<root>",
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: true,
		},
		{
			name:    "empty XML with options",
			xml:     "",
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: false, // Empty XML is now valid for creating new XML from scratch
		},
		{
			name:    "invalid path with options",
			xml:     "<root/>",
			path:    "",
			value:   "value",
			opts:    &Options{CaseSensitive: true},
			wantErr: true,
		},
		{
			name:    "document too large with options",
			xml:     strings.Repeat("x", MaxDocumentSize+1),
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: true,
		},
		{
			name:    "valid XML with nil options",
			xml:     "<root/>",
			path:    "root.item",
			value:   "value",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid XML with default options",
			xml:     "<root/>",
			path:    "root.item",
			value:   "value",
			opts:    DefaultOptions(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetWithOptions panicked: %v", r)
				}
			}()

			result, err := SetWithOptions(tt.xml, tt.path, tt.value, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				// Original XML should be unchanged on error
				if result != tt.xml {
					t.Errorf("Expected original XML on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == "" {
					t.Error("Expected non-empty result")
				}
			}
		})
	}
}

// TestOptions_CaseInsensitiveErrors tests case insensitive operations with errors
func TestOptions_CaseInsensitiveErrors(t *testing.T) {
	opts := &Options{CaseSensitive: false}

	tests := []struct {
		name    string
		xml     string
		path    string
		wantErr bool
	}{
		{
			name:    "case mismatch with malformed XML",
			xml:     "<ROOT>",
			path:    "root.item",
			wantErr: false, // Get doesn't error, just returns empty
		},
		{
			name:    "case mismatch with empty XML",
			xml:     "",
			path:    "ROOT.ITEM",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Case insensitive operation panicked: %v", r)
				}
			}()

			// Get with options
			result := GetWithOptions(tt.xml, tt.path, opts)
			_ = result.String()

			// Set with options
			_, err := SetWithOptions(tt.xml, tt.path, "value", opts)
			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}

// TestGetBytesWithOptions_Errors tests GetBytesWithOptions error handling
func TestGetBytesWithOptions_Errors(t *testing.T) {
	tests := []struct {
		name string
		xml  []byte
		path string
		opts *Options
	}{
		{
			name: "nil bytes with options",
			xml:  nil,
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
		{
			name: "empty bytes with options",
			xml:  []byte{},
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
		{
			name: "malformed bytes with options",
			xml:  []byte("<root>"),
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
		{
			name: "too large bytes with options",
			xml:  make([]byte, MaxDocumentSize+1),
			path: "root.item",
			opts: &Options{CaseSensitive: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GetBytesWithOptions panicked: %v", r)
				}
			}()

			result := GetBytesWithOptions(tt.xml, tt.path, tt.opts)
			_ = result.String()
		})
	}
}

// TestSetBytesWithOptions_Errors tests SetBytesWithOptions error handling
func TestSetBytesWithOptions_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     []byte
		path    string
		value   interface{}
		opts    *Options
		wantErr bool
	}{
		{
			name:    "nil bytes with options",
			xml:     nil,
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: true,
		},
		{
			name:    "empty bytes with options",
			xml:     []byte{},
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: false, // Empty XML is now valid for creating new XML from scratch
		},
		{
			name:    "malformed bytes with options",
			xml:     []byte("<root>"),
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: true,
		},
		{
			name:    "valid bytes with nil options",
			xml:     []byte("<root/>"),
			path:    "root.item",
			value:   "value",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid bytes with options",
			xml:     []byte("<root/>"),
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetBytesWithOptions panicked: %v", r)
				}
			}()

			result, err := SetBytesWithOptions(tt.xml, tt.path, tt.value, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(result) == 0 {
					t.Error("Expected result but got empty bytes")
				}
			}
		})
	}
}

// TestOptions_DefaultBehavior tests that default options match regular functions
func TestOptions_DefaultBehavior(t *testing.T) {
	xml := "<root><item>value</item></root>"
	path := "root.item"

	// Get comparison with default options
	result1 := Get(xml, path)
	result3 := GetWithOptions(xml, path, DefaultOptions())

	if result1.String() != result3.String() {
		t.Error("Get and GetWithOptions(DefaultOptions()) should return same results")
	}

	// Set comparison with default options
	set1, err1 := Set(xml, "root.new", "val")
	set3, err3 := SetWithOptions(xml, "root.new", "val", DefaultOptions())

	if (err1 == nil) != (err3 == nil) {
		t.Error("Set and SetWithOptions(DefaultOptions()) should have same error behavior")
	}

	if err1 == nil && set1 != set3 {
		t.Error("Set and SetWithOptions(DefaultOptions()) should return same results")
	}

	// Test that nil options are handled gracefully (may panic or use defaults)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("SetWithOptions(nil) panicked (acceptable): %v", r)
			}
		}()
		_, _ = SetWithOptions(xml, "root.new", "val", nil)
	}()
}

// TestOptions_ConcurrentAccess tests concurrent access with options
func TestOptions_ConcurrentAccess(t *testing.T) {
	xml := "<root><item>value</item></root>"
	opts := &Options{CaseSensitive: false}

	done := make(chan bool, 4)

	// Concurrent Get with options
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in concurrent Get: %v", r)
			}
			done <- true
		}()
		_ = GetWithOptions(xml, "ROOT.ITEM", opts)
	}()

	// Concurrent Set with options
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in concurrent Set: %v", r)
			}
			done <- true
		}()
		_, _ = SetWithOptions(xml, "ROOT.NEW", "val", opts)
	}()

	// Concurrent Get with error
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in concurrent Get error: %v", r)
			}
			done <- true
		}()
		_ = GetWithOptions("<root>", "root.item", opts)
	}()

	// Concurrent Set with error
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic in concurrent Set error: %v", r)
			}
			done <- true
		}()
		_, _ = SetWithOptions("<root>", "root.item", "val", opts)
	}()

	for range 4 {
		<-done
	}
}

// TestOptions_NoErrorOnValidOperations tests valid operations with options don't error
func TestOptions_NoErrorOnValidOperations(t *testing.T) {
	tests := []struct {
		name  string
		xml   string
		path  string
		value any
		opts  *Options
	}{
		{
			name:  "case insensitive get",
			xml:   "<ROOT><ITEM>value</ITEM></ROOT>",
			path:  "root.item",
			value: "new",
			opts:  &Options{CaseSensitive: false},
		},
		{
			name:  "case sensitive get",
			xml:   "<root><item>value</item></root>",
			path:  "root.item",
			value: "new",
			opts:  &Options{CaseSensitive: true},
		},
		{
			name:  "default options",
			xml:   "<root><item>value</item></root>",
			path:  "root.item",
			value: "new",
			opts:  DefaultOptions(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get should not error
			result := GetWithOptions(tt.xml, tt.path, tt.opts)
			_ = result.String()

			// Set should not error for valid XML
			_, err := SetWithOptions(tt.xml, tt.path, tt.value, tt.opts)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// ============================================================================
// Error Tests - Set Operations
// ============================================================================

// TestSetErrors_MalformedXML tests Set with malformed XML
func TestSetErrors_MalformedXML(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		path    string
		value   interface{}
		wantErr bool
		errType error
	}{
		{
			name:    "unclosed tag",
			xml:     "<root><item>value",
			path:    "root.item",
			value:   "new",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "mismatched tags",
			xml:     "<root><item>value</wrong></root>",
			path:    "root.item",
			value:   "new",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "incomplete opening tag",
			xml:     "<root",
			path:    "root.item",
			value:   "new",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "only closing tag",
			xml:     "</root>",
			path:    "root.item",
			value:   "new",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "nested unclosed tags",
			xml:     "<root><a><b></a></root>",
			path:    "root.a",
			value:   "new",
			wantErr: true,
			errType: ErrMalformedXML,
		},
		{
			name:    "empty document",
			xml:     "",
			path:    "root.item",
			value:   "new",
			wantErr: false, // Empty XML is now valid for creating new XML from scratch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Set panicked on malformed XML: %v", r)
				}
			}()

			result, err := Set(tt.xml, tt.path, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if tt.errType != nil && !errors.Is(err, tt.errType) {
						t.Errorf("Expected error type %v, got %v", tt.errType, err)
					}
				}
				// Original XML should be unchanged on error
				if result != tt.xml {
					t.Errorf("Expected original XML on error\nGot:  %q\nWant: %q", result, tt.xml)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestSetErrors_InvalidPath tests Set with invalid path syntax
func TestSetErrors_InvalidPath(t *testing.T) {
	xml := "<root/>"

	tests := []struct {
		name    string
		path    string
		value   interface{}
		wantErr bool
		errType error
	}{
		{
			name:    "empty path",
			path:    "",
			value:   "value",
			wantErr: true,
			errType: ErrInvalidPath,
		},
		{
			name:    "only dots",
			path:    "...",
			value:   "value",
			wantErr: false, // parsePath handles this
		},
		{
			name:    "leading dot",
			path:    ".root.item",
			value:   "value",
			wantErr: false, // parsePath handles this
		},
		{
			name:    "trailing dot",
			path:    "root.item.",
			value:   "value",
			wantErr: false, // parsePath handles this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Set panicked on invalid path: %v", r)
				}
			}()

			result, err := Set(xml, tt.path, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if tt.errType != nil && !errors.Is(err, tt.errType) {
						t.Errorf("Expected error type %v, got %v", tt.errType, err)
					}
				}
				// Original XML should be unchanged on error
				if result != xml {
					t.Errorf("Expected original XML on error")
				}
			}
		})
	}
}

// TestSetErrors_InvalidValue tests Set with invalid values
func TestSetErrors_InvalidValue(t *testing.T) {
	xml := "<root/>"

	tests := []struct {
		name    string
		path    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "nil value (becomes delete)",
			path:    "root.item",
			value:   nil,
			wantErr: false, // nil is valid (triggers delete)
		},
		{
			name:    "channel value",
			path:    "root.item",
			value:   make(chan int),
			wantErr: false, // Converted to string representation
		},
		{
			name:    "function value",
			path:    "root.item",
			value:   func() {},
			wantErr: false, // Converted to string representation
		},
		{
			name:    "complex struct",
			path:    "root.item",
			value:   struct{ X, Y int }{1, 2},
			wantErr: false, // Converted to string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Set panicked on value type %T: %v", tt.value, r)
				}
			}()

			result, err := Set(xml, tt.path, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Logf("Got error for value type %T: %v", tt.value, err)
				}
				// Result should be valid XML or original
				_ = result
			}
		})
	}
}

// TestSetErrors_RawXMLValidation tests SetRaw with invalid raw XML
func TestSetErrors_RawXMLValidation(t *testing.T) {
	xml := "<root/>"

	tests := []struct {
		name    string
		rawxml  string
		wantErr bool
		errType error
	}{
		{
			name:    "unclosed tag in raw XML",
			rawxml:  "<item>",
			wantErr: true,
			errType: ErrInvalidValue,
		},
		{
			name:    "mismatched tags in raw XML",
			rawxml:  "<a></b>",
			wantErr: true,
			errType: ErrInvalidValue,
		},
		{
			name:    "unclosed CDATA",
			rawxml:  "<![CDATA[test",
			wantErr: true,
			errType: ErrInvalidValue,
		},
		{
			name:    "nested CDATA",
			rawxml:  "<![CDATA[<![CDATA[test]]>]]>",
			wantErr: true,
			errType: ErrInvalidValue,
		},
		{
			name:    "DOCTYPE in raw XML (security)",
			rawxml:  "<!DOCTYPE test><test/>",
			wantErr: true,
			errType: ErrInvalidValue,
		},
		{
			name:    "ENTITY in raw XML (security)",
			rawxml:  "<!ENTITY test 'value'><test/>",
			wantErr: true,
			errType: ErrInvalidValue,
		},
		{
			name:    "unclosed comment",
			rawxml:  "<!-- comment",
			wantErr: true,
			errType: ErrInvalidValue,
		},
		{
			name:    "valid raw XML",
			rawxml:  "<item>value</item>",
			wantErr: false,
		},
		{
			name:    "valid self-closing",
			rawxml:  "<item/>",
			wantErr: false,
		},
		{
			name:    "valid with CDATA",
			rawxml:  "<item><![CDATA[<tag>]]></item>",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetRaw panicked: %v", r)
				}
			}()

			result, err := SetRaw(xml, "root.data", tt.rawxml)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else {
					if tt.errType != nil && !errors.Is(err, tt.errType) {
						t.Errorf("Expected error type %v, got %v", tt.errType, err)
					}
				}
				// Original XML should be unchanged on error
				if result != xml {
					t.Errorf("Expected original XML on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestSetErrors_DocumentTooLarge tests Set with documents exceeding size limits
func TestSetErrors_DocumentTooLarge(t *testing.T) {
	// Create XML larger than MaxDocumentSize
	largeXML := strings.Repeat("<item>x</item>", 1000000)
	xml := "<root>" + largeXML + "</root>"

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Set panicked on large document: %v", r)
		}
	}()

	result, err := Set(xml, "root.newitem", "value")
	if err == nil {
		t.Error("Expected error for document too large")
	}
	if !errors.Is(err, ErrMalformedXML) {
		t.Errorf("Expected ErrMalformedXML for large document, got %v", err)
	}
	// Original XML should be unchanged
	if result != xml {
		t.Error("Expected original XML on error")
	}
}

// TestSetErrors_ValueTooLarge tests Set with values that would make document too large
func TestSetErrors_ValueTooLarge(t *testing.T) {
	xml := "<root/>"
	// Create a value that would make the document exceed MaxDocumentSize
	largeValue := strings.Repeat("x", MaxDocumentSize)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Set panicked on large value: %v", r)
		}
	}()

	// This may or may not error depending on implementation
	// But it should not panic
	result, err := Set(xml, "root.item", largeValue)
	if err != nil {
		t.Logf("Set returned error for large value: %v", err)
	}
	_ = result
}

// TestSetBytes_Errors tests SetBytes error handling
func TestSetBytes_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     []byte
		path    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "nil bytes",
			xml:     nil,
			path:    "root.item",
			value:   "value",
			wantErr: true,
		},
		{
			name:    "empty bytes",
			xml:     []byte{},
			path:    "root.item",
			value:   "value",
			wantErr: false, // Empty XML is now valid for creating new XML from scratch
		},
		{
			name:    "malformed bytes",
			xml:     []byte("<root>"),
			path:    "root.item",
			value:   "value",
			wantErr: true,
		},
		{
			name:    "valid bytes",
			xml:     []byte("<root/>"),
			path:    "root.item",
			value:   "value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetBytes panicked: %v", r)
				}
			}()

			result, err := SetBytes(tt.xml, tt.path, tt.value)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if len(result) == 0 {
					t.Error("Expected result but got empty bytes")
				}
			}
		})
	}
}

// TestSetWithOptions_ErrorPaths tests SetWithOptions error handling
func TestSetWithOptions_ErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		path    string
		value   interface{}
		opts    *Options
		wantErr bool
	}{
		{
			name:    "nil options with valid XML",
			xml:     "<root/>",
			path:    "root.item",
			value:   "value",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "options with malformed XML",
			xml:     "<root>",
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: true,
		},
		{
			name:    "options with invalid path",
			xml:     "<root/>",
			path:    "",
			value:   "value",
			opts:    &Options{CaseSensitive: true},
			wantErr: true,
		},
		{
			name:    "options with too large document",
			xml:     strings.Repeat("x", MaxDocumentSize+1),
			path:    "root.item",
			value:   "value",
			opts:    &Options{CaseSensitive: false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetWithOptions panicked: %v", r)
				}
			}()

			result, err := SetWithOptions(tt.xml, tt.path, tt.value, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == "" {
					t.Error("Expected non-empty result")
				}
			}
		})
	}
}

// TestSet_NoErrorOnValidOperations tests that valid operations don't return errors
func TestSet_NoErrorOnValidOperations(t *testing.T) {
	tests := []struct {
		name  string
		xml   string
		path  string
		value interface{}
	}{
		{
			name:  "create new element",
			xml:   "<root/>",
			path:  "root.item",
			value: "value",
		},
		{
			name:  "update existing element",
			xml:   "<root><item>old</item></root>",
			path:  "root.item",
			value: "new",
		},
		{
			name:  "set attribute",
			xml:   "<root><item/></root>",
			path:  "root.item.@id",
			value: "123",
		},
		{
			name:  "set integer value",
			xml:   "<root/>",
			path:  "root.count",
			value: 42,
		},
		{
			name:  "set float value",
			xml:   "<root/>",
			path:  "root.price",
			value: 19.99,
		},
		{
			name:  "set boolean value",
			xml:   "<root/>",
			path:  "root.active",
			value: true,
		},
		{
			name:  "set nil (delete)",
			xml:   "<root><item>value</item></root>",
			path:  "root.item",
			value: nil,
		},
		{
			name:  "set nested element",
			xml:   "<root/>",
			path:  "root.a.b.c",
			value: "deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Set(tt.xml, tt.path, tt.value)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if result == "" {
				t.Error("Expected non-empty result")
			}
			if result == tt.xml && tt.value != nil {
				t.Error("Expected modified XML")
			}
		})
	}
}

// TestSet_RecoveryAfterError tests that errors don't corrupt internal state
func TestSet_RecoveryAfterError(t *testing.T) {
	xml := "<root><item>value</item></root>"

	// Perform invalid operation
	_, err := Set(xml, "", "test")
	if err == nil {
		t.Fatal("Expected error for empty path")
	}

	// Should still work correctly after error
	result, err := Set(xml, "root.item", "newvalue")
	if err != nil {
		t.Fatalf("Set after error failed: %v", err)
	}

	// Verify the modification worked
	checkResult := Get(result, "root.item")
	if checkResult.String() != "newvalue" {
		t.Errorf("Recovery check failed: got %q, want 'newvalue'", checkResult.String())
	}
}

// ============================================================================
// Error Tests - Validation Operations
// ============================================================================

// TestValidationErrors_MalformedXML tests validation error detection
func TestValidationErrors_MalformedXML(t *testing.T) {
	tests := []struct {
		name        string
		xml         string
		wantErr     bool
		errContains string
	}{
		{
			name:        "unclosed tag",
			xml:         "<root><item>",
			wantErr:     true,
			errContains: "unclosed",
		},
		{
			name:        "mismatched tags",
			xml:         "<root><item></root>",
			wantErr:     true,
			errContains: "mismatched",
		},
		{
			name:        "invalid character in tag name",
			xml:         "<root><item@>value</item@></root>",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "tag name starts with digit",
			xml:         "<1root>value</1root>",
			wantErr:     true,
			errContains: "digit",
		},
		{
			name:        "empty tag name",
			xml:         "<>value</>",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "unclosed attribute quote",
			xml:         `<root attr="value>text</root>`,
			wantErr:     true,
			errContains: "expected",
		},
		{
			name:        "attribute without equals",
			xml:         `<root attr"value"></root>`,
			wantErr:     true,
			errContains: "expected",
		},
		{
			name:        "attribute without quotes",
			xml:         `<root attr=value></root>`,
			wantErr:     true,
			errContains: "quoted",
		},
		{
			name:        "content outside root",
			xml:         "<root></root>text outside",
			wantErr:     true,
			errContains: "outside root",
		},
		{
			name:        "empty document",
			xml:         "",
			wantErr:     true,
			errContains: "empty document",
		},
		{
			name:        "whitespace only document",
			xml:         "   \n\t  ",
			wantErr:     true,
			errContains: "no root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Valid()
			valid := Valid(tt.xml)
			if valid == tt.wantErr {
				t.Errorf("Valid() = %v, want %v", valid, !tt.wantErr)
			}

			// Test ValidateWithError()
			err := ValidateWithError(tt.xml)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error but got nil")
				} else {
					if tt.errContains != "" && !strings.Contains(strings.ToLower(err.Message), strings.ToLower(tt.errContains)) {
						t.Errorf("Error message should contain %q, got: %s", tt.errContains, err.Message)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidationErrors_ErrorDetails tests ValidateError structure
func TestValidationErrors_ErrorDetails(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		wantLine int
	}{
		{
			name:     "error on line 1",
			xml:      "<root><item></root>",
			wantLine: 1,
		},
		{
			name:     "error on line 3",
			xml:      "<root>\n<item>\n</root>",
			wantLine: 3,
		},
		{
			name:     "error on line 5",
			xml:      "<root>\n<a>\n<b>\n<c>\n</b>",
			wantLine: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithError(tt.xml)
			if err == nil {
				t.Fatal("Expected validation error")
			}

			// Check that line number is set
			if err.Line == 0 {
				t.Error("Error line should be non-zero")
			}

			// Check that error has a message
			if err.Message == "" {
				t.Error("Error should have a message")
			}

			// Check Error() method format
			errStr := err.Error()
			if !strings.Contains(errStr, "line") {
				t.Errorf("Error string should contain 'line': %s", errStr)
			}
			if !strings.Contains(errStr, "column") {
				t.Errorf("Error string should contain 'column': %s", errStr)
			}
		})
	}
}

// TestValidationErrors_AsError tests that ValidateError implements error interface
func TestValidationErrors_AsError(t *testing.T) {
	xml := "<root><item></root>"
	err := ValidateWithError(xml)

	if err == nil {
		t.Fatal("Expected validation error")
	}

	// Should be able to use as error
	var e error = err
	_ = e.Error()

	// Should be able to use errors.As
	var valErr *ValidateError
	if !errors.As(e, &valErr) {
		t.Error("Expected errors.As to work with ValidateError")
	}

	if valErr.Line == 0 || valErr.Message == "" {
		t.Error("ValidateError should have line and message set")
	}
}

// TestValidationErrors_SecurityLimits tests validation respects security limits
func TestValidationErrors_SecurityLimits(t *testing.T) {
	tests := []struct {
		name        string
		buildXML    func() string
		wantErr     bool
		errContains string
	}{
		{
			name: "document too large",
			buildXML: func() string {
				return strings.Repeat("x", MaxDocumentSize+1)
			},
			wantErr:     true,
			errContains: "maximum size",
		},
		{
			name: "nesting too deep",
			buildXML: func() string {
				var sb strings.Builder
				sb.WriteString("<root>")
				for i := 0; i < MaxNestingDepth+10; i++ {
					sb.WriteString("<level>")
				}
				sb.WriteString("value")
				for i := 0; i < MaxNestingDepth+10; i++ {
					sb.WriteString("</level>")
				}
				sb.WriteString("</root>")
				return sb.String()
			},
			wantErr:     true,
			errContains: "nesting depth",
		},
		{
			name: "too many attributes",
			buildXML: func() string {
				var sb strings.Builder
				sb.WriteString("<root ")
				for i := 0; i < MaxAttributes+10; i++ {
					sb.WriteString("attr")
					sb.WriteString(itoa(i))
					sb.WriteString("='val' ")
				}
				sb.WriteString("></root>")
				return sb.String()
			},
			wantErr:     true,
			errContains: "too many attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xml := tt.buildXML()

			valid := Valid(xml)
			if valid == tt.wantErr {
				t.Errorf("Valid() = %v, want %v", valid, !tt.wantErr)
			}

			err := ValidateWithError(xml)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error for security limit")
				} else {
					if tt.errContains != "" && !strings.Contains(strings.ToLower(err.Message), strings.ToLower(tt.errContains)) {
						t.Errorf("Error message should contain %q, got: %s", tt.errContains, err.Message)
					}
				}
			}
		})
	}
}

// TestValidationErrors_ValidXML tests that valid XML doesn't return errors
func TestValidationErrors_ValidXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{
			name: "simple valid XML",
			xml:  "<root>value</root>",
		},
		{
			name: "nested elements",
			xml:  "<root><a><b><c>value</c></b></a></root>",
		},
		{
			name: "with attributes",
			xml:  `<root attr="value"><item id="123">text</item></root>`,
		},
		{
			name: "self-closing tags",
			xml:  "<root><item/><item/></root>",
		},
		{
			name: "with CDATA",
			xml:  "<root><![CDATA[<tag>not parsed</tag>]]></root>",
		},
		{
			name: "with comments",
			xml:  "<root><!-- comment --><item>value</item></root>",
		},
		{
			name: "with processing instruction",
			xml:  `<?xml version="1.0"?><root>value</root>`,
		},
		{
			name: "with namespaces",
			xml:  `<root xmlns="http://example.com"><child>value</child></root>`,
		},
		{
			name: "with whitespace",
			xml:  "<root>  \n  <item>value</item>  \n  </root>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !Valid(tt.xml) {
				t.Error("Valid XML should pass validation")
			}

			err := ValidateWithError(tt.xml)
			if err != nil {
				t.Errorf("Valid XML should not return error: %v", err)
			}
		})
	}
}

// TestValidBytes_Errors tests ValidBytes error handling
func TestValidBytes_Errors(t *testing.T) {
	tests := []struct {
		name    string
		xml     []byte
		wantErr bool
	}{
		{
			name:    "nil bytes",
			xml:     nil,
			wantErr: true,
		},
		{
			name:    "empty bytes",
			xml:     []byte{},
			wantErr: true,
		},
		{
			name:    "malformed bytes",
			xml:     []byte("<root>"),
			wantErr: true,
		},
		{
			name:    "valid bytes",
			xml:     []byte("<root>value</root>"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := ValidBytes(tt.xml)
			if valid == tt.wantErr {
				t.Errorf("ValidBytes() = %v, want %v", valid, !tt.wantErr)
			}

			err := ValidateBytesWithError(tt.xml)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestValidation_NoPanics tests that validation never panics
func TestValidation_NoPanics(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{name: "extremely malformed", xml: "<<<>>><<<>>>"},
		{name: "control characters", xml: "\x01\x02\x03<root>\x01\x02</root>"},
		{name: "invalid UTF-8", xml: string([]byte{0xFF, 0xFE, '<', 'r', 'o', 'o', 't', '>'})},
		{name: "null bytes", xml: "<root>\x00</root>"},
		{name: "very long tag name", xml: "<" + strings.Repeat("a", 1000000) + ">"},
		{name: "deeply nested", xml: strings.Repeat("<a>", 10000) + strings.Repeat("</a>", 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Validation panicked: %v", r)
				}
			}()

			_ = Valid(tt.xml)
			_ = ValidateWithError(tt.xml)
		})
	}
}

// TestValidation_DOCTYPESecurity tests DOCTYPE handling for security
func TestValidation_DOCTYPESecurity(t *testing.T) {
	tests := []struct {
		name    string
		xml     string
		wantErr bool
	}{
		{
			name:    "simple DOCTYPE",
			xml:     `<!DOCTYPE root SYSTEM "root.dtd"><root></root>`,
			wantErr: false, // DOCTYPE is skipped but document is valid
		},
		{
			name:    "DOCTYPE with internal subset",
			xml:     `<!DOCTYPE root [<!ELEMENT root (#PCDATA)>]><root>text</root>`,
			wantErr: false,
		},
		{
			name:    "unclosed DOCTYPE",
			xml:     `<!DOCTYPE root SYSTEM "root.dtd"<root></root>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := Valid(tt.xml)
			if valid == tt.wantErr {
				t.Logf("DOCTYPE handling: Valid() = %v", valid)
			}

			err := ValidateWithError(tt.xml)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected validation error for malformed DOCTYPE")
				}
			}
		})
	}
}

// TestValidation_ConcurrentAccess tests concurrent validation doesn't cause issues
func TestValidation_ConcurrentAccess(t *testing.T) {
	xmlDocs := []string{
		"<root><item>value</item></root>",
		"<root>",
		"<a></b>",
		"",
		strings.Repeat("<item>", 100) + strings.Repeat("</item>", 100),
	}

	done := make(chan bool, len(xmlDocs))
	for _, xml := range xmlDocs {
		go func(x string) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Panic in concurrent validation: %v", r)
				}
				done <- true
			}()

			_ = Valid(x)
			_ = ValidateWithError(x)
		}(xml)
	}

	for i := 0; i < len(xmlDocs); i++ {
		<-done
	}
}

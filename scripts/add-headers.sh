#!/bin/bash
# Script to add SPDX license identifier and copyright notice to all Go files

set -e

# Header to add
SPDX_LINE="// SPDX-License-Identifier: MIT"
COPYRIGHT_LINE="// Copyright (c) 2025 Daniel Schmidt"

# Find all .go files (excluding vendor and examples)
find . -name "*.go" -not -path "./vendor/*" -not -path "./examples/*" | while read -r file; do
    # Check if file already has SPDX header
    if head -n 1 "$file" | grep -q "SPDX-License-Identifier"; then
        echo "✓ $file (already has header)"
        continue
    fi

    # Create temp file with header
    temp_file=$(mktemp)

    # Add header
    echo "$SPDX_LINE" > "$temp_file"
    echo "$COPYRIGHT_LINE" >> "$temp_file"
    echo "" >> "$temp_file"

    # Append original file content
    cat "$file" >> "$temp_file"

    # Replace original file
    mv "$temp_file" "$file"

    echo "✓ $file (header added)"
done

echo ""
echo "Header addition complete!"

#!/bin/bash
# Script to verify SPDX license identifier and copyright notice in all Go files

set -e

# Expected header lines
EXPECTED_SPDX="// SPDX-License-Identifier: MIT"
EXPECTED_COPYRIGHT="// Copyright (c) 2025 Daniel Schmidt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Counter for missing headers
missing_count=0
files_checked=0

echo "Checking Go file headers..."
echo ""

# Find all .go files (excluding vendor and examples) - store in temp file for portability
temp_file=$(mktemp)
find . -name "*.go" -not -path "./vendor/*" -not -path "./examples/*" -type f > "$temp_file"

# Read files from temp file
while IFS= read -r file; do
    files_checked=$((files_checked + 1))

    # Read first two lines
    line1=$(head -n 1 "$file")
    line2=$(head -n 2 "$file" | tail -n 1)

    # Check if both lines match expected headers
    if [[ "$line1" != "$EXPECTED_SPDX" ]] || [[ "$line2" != "$EXPECTED_COPYRIGHT" ]]; then
        echo -e "${RED}✗${NC} $file (missing or incorrect header)"
        echo "  Expected line 1: $EXPECTED_SPDX"
        echo "  Got line 1:      $line1"
        echo "  Expected line 2: $EXPECTED_COPYRIGHT"
        echo "  Got line 2:      $line2"
        echo ""
        missing_count=$((missing_count + 1))
    fi
done < "$temp_file"

# Cleanup
rm -f "$temp_file"

echo "----------------------------------------"
echo "Files checked: $files_checked"

if [ $missing_count -eq 0 ]; then
    echo -e "${GREEN}✓ All Go files have correct headers!${NC}"
    exit 0
else
    echo -e "${RED}✗ $missing_count file(s) missing or have incorrect headers${NC}"
    exit 1
fi

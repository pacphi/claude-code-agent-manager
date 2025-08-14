#!/bin/bash

# Fix documentation links to point to the correct agent locations
# This script is typically called as a post-install action by agent-manager

set -e

# Color codes for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse command line arguments
CONFIG_SOURCE=""
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --config)
            CONFIG_SOURCE="$2"
            shift 2
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        *)
            shift
            ;;
    esac
done

if [ "$VERBOSE" = true ]; then
    echo -e "${BLUE}Fixing documentation links...${NC}"
    if [ -n "$CONFIG_SOURCE" ]; then
        echo -e "${BLUE}Config source: $CONFIG_SOURCE${NC}"
    fi
fi

# Function to get category name from doc filename
get_category() {
    case "$1" in
        "BUSINESS_PRODUCT.md") echo "business-product" ;;
        "CORE_DEVELOPMENT.md") echo "core-development" ;;
        "DATA_AI.md") echo "data-ai" ;;
        "DEVELOPER_EXPERIENCE.md") echo "developer-experience" ;;
        "INFRASTRUCTURE.md") echo "infrastructure" ;;
        "LANGUAGE_SPECIALISTS.md") echo "language-specialists" ;;
        "META_ORCHESTRATION.md") echo "meta-orchestration" ;;
        "QUALITY_SECURITY.md") echo "quality-security" ;;
        "RESEARCH_ANALYSIS.md") echo "research-analysis" ;;
        "SPECIALIZED_DOMAINS.md") echo "specialized-domains" ;;
        *)
            # Try to infer from filename by converting to lowercase and replacing underscores
            local name=$(echo "$1" | sed 's/\.md$//' | tr '[:upper:]' '[:lower:]' | tr '_' '-')
            echo "$name"
            ;;
    esac
}

# Function to get base directory from agents-config.yaml
get_base_dir() {
    if [ -f "agents-config.yaml" ]; then
        # Extract base_dir from YAML (simple grep approach)
        grep "base_dir:" agents-config.yaml | head -n1 | sed 's/.*base_dir: *//' | tr -d '"' || echo ".claude/agents"
    else
        echo ".claude/agents"
    fi
}

# Function to get docs directory from agents-config.yaml
get_docs_dir() {
    if [ -f "agents-config.yaml" ]; then
        # Extract docs_dir from YAML (simple grep approach)
        grep "docs_dir:" agents-config.yaml | head -n1 | sed 's/.*docs_dir: *//' | tr -d '"' || echo "docs"
    else
        echo "docs"
    fi
}

# Get configuration
BASE_DIR=$(get_base_dir)
DOCS_DIR=$(get_docs_dir)/voltagent

if [ "$VERBOSE" = true ]; then
    echo -e "${BLUE}Base directory: $BASE_DIR${NC}"
    echo -e "${BLUE}Docs directory: $DOCS_DIR${NC}"
fi

# Check if docs directory exists
if [ ! -d "$DOCS_DIR" ]; then
    if [ "$VERBOSE" = true ]; then
        echo -e "${YELLOW}Warning: docs directory not found: $DOCS_DIR${NC}"
    fi
    exit 0
fi

# Count total links fixed
total_links_fixed=0
files_processed=0

# Process each documentation file
for doc_path in "$DOCS_DIR"/*.md; do
    if [ -f "$doc_path" ]; then
        doc_file=$(basename "$doc_path")

        # Skip files that don't match agent category patterns
        if [[ ! "$doc_file" =~ ^[A-Z_]+\.md$ ]] && [[ ! "$doc_file" =~ ^[a-z-]+\.md$ ]]; then
            if [ "$VERBOSE" = true ]; then
                echo "  - Skipping $doc_file (not an agent category doc)"
            fi
            continue
        fi

        category=$(get_category "$doc_file")

        if [ -z "$category" ]; then
            if [ "$VERBOSE" = true ]; then
                echo "  - Skipping $doc_file (unable to determine category)"
            fi
            continue
        fi

        # Check if the category directory exists
        category_dir="$BASE_DIR/$category"
        if [ ! -d "$category_dir" ]; then
            if [ "$VERBOSE" = true ]; then
                echo "  - Skipping $doc_file (category directory not found: $category_dir)"
            fi
            continue
        fi

        echo "Processing $doc_file (category: $category)..."

        # Create a temporary file for the updated content
        temp_file=$(mktemp)
        links_fixed=0

        # Read the file and update links
        while IFS= read -r line; do
            # Match markdown links that point to .md files (not already with paths)
            # Pattern: [**agent-name**](agent-name.md) or [text](agent-name.md)
            if [[ "$line" =~ \[.*\]\([^/]*\.md\) ]]; then
                # Replace simple .md links with full paths
                original_line="$line"
                updated_line=$(echo "$line" | sed -E "s|\]\(([^/)]+\.md)\)|](../../$BASE_DIR/$category/\1)|g")

                if [ "$original_line" != "$updated_line" ]; then
                    # Count how many links were fixed in this line
                    line_fixes=$(echo "$original_line" | grep -o '\]([^/]*\.md)' | wc -l)
                    links_fixed=$((links_fixed + line_fixes))
                fi

                echo "$updated_line" >> "$temp_file"
            else
                echo "$line" >> "$temp_file"
            fi
        done < "$doc_path"

        # Replace the original file with the updated content if changes were made
        if [ $links_fixed -gt 0 ]; then
            mv "$temp_file" "$doc_path"
            echo "  ✓ Fixed $links_fixed links in $doc_file"
            total_links_fixed=$((total_links_fixed + links_fixed))
            files_processed=$((files_processed + 1))
        else
            rm "$temp_file"
            if [ "$VERBOSE" = true ]; then
                echo "  - No links to fix in $doc_file"
            fi
        fi
    fi
done

if [ $files_processed -gt 0 ]; then
    echo -e "${GREEN}✓ Documentation links fixed!${NC}"
    echo "Fixed $total_links_fixed links in $files_processed files"
    echo "All agent links now point to the correct locations in $BASE_DIR/"
else
    if [ "$VERBOSE" = true ]; then
        echo -e "${YELLOW}No documentation files required link fixes${NC}"
    fi
fi
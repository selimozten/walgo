#!/bin/bash

# Update version to v0.3.1
# Run from project root: ./scripts/update-version.sh

NEW_VERSION="0.3.1"
OLD_VERSION="0.2.1"

echo "🔄 Updating version from $OLD_VERSION to $NEW_VERSION..."

# Update cmd/version.go
sed -i.bak "s/Version = \"$OLD_VERSION\"/Version = \"$NEW_VERSION\"/" cmd/version.go
echo "✅ Updated cmd/version.go"

# Update pkg/api/api.go
sed -i.bak "s/Version:   \"$OLD_VERSION\"/Version:   \"$NEW_VERSION\"/" pkg/api/api.go
echo "✅ Updated pkg/api/api.go"

# Update docs/README.md
sed -i.bak "s/Current Version:\*\* $OLD_VERSION/Current Version:** $NEW_VERSION/" docs/README.md
echo "✅ Updated docs/README.md"

# Update docs/COMMANDS.md
sed -i.bak "s/(v$OLD_VERSION)/(v$NEW_VERSION)/g" docs/COMMANDS.md
sed -i.bak "s/version $OLD_VERSION/version $NEW_VERSION/g" docs/COMMANDS.md
sed -i.bak "s/v$OLD_VERSION/v$NEW_VERSION/g" docs/COMMANDS.md
echo "✅ Updated docs/COMMANDS.md"

# Clean up backup files
find . -name "*.bak" -type f -delete

echo ""
echo "✅ Version updated to $NEW_VERSION"
echo ""

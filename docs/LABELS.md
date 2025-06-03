# GitHub Labels Guide

This document outlines the recommended labels for the Walgo repository and their meanings.

## Issue Type Labels

### `bug` 🐛
- **Color**: `d73a4a` (red)
- **Description**: Something isn't working as expected
- **Usage**: Apply to issues reporting problems, errors, or unexpected behavior

### `enhancement` ✨
- **Color**: `a2eeef` (light blue)
- **Description**: New feature or request
- **Usage**: Apply to feature requests and enhancement suggestions

### `documentation` 📚
- **Color**: `0075ca` (blue)
- **Description**: Improvements or additions to documentation
- **Usage**: Apply to issues related to README, guides, comments, or man pages

### `question` ❓
- **Color**: `d876e3` (purple)
- **Description**: Further information is requested
- **Usage**: Apply to issues asking for help or clarification

## Priority Labels

### `priority: high` 🔥
- **Color**: `ff6b6b` (bright red)
- **Description**: Critical issues that need immediate attention
- **Usage**: Security vulnerabilities, data loss, deployment blockers

### `priority: medium` ⚡
- **Color**: `ffa500` (orange)
- **Description**: Important issues that should be addressed soon
- **Usage**: Performance issues, usability problems

### `priority: low` 🔹
- **Color**: `0e8a16` (green)
- **Description**: Nice to have improvements
- **Usage**: Minor enhancements, code cleanup

## Difficulty Labels

### `good first issue` 🌱
- **Color**: `7057ff` (purple)
- **Description**: Good for newcomers to the project
- **Usage**: Simple, well-defined issues suitable for first-time contributors

### `help wanted` 🙋
- **Color**: `008672` (teal)
- **Description**: Extra attention is needed from the community
- **Usage**: Issues that would benefit from community involvement

### `advanced` 🧠
- **Color**: `5319e7` (dark purple)
- **Description**: Requires deep knowledge of the codebase
- **Usage**: Complex issues requiring significant expertise

## Component Labels

### `cmd` 🖥️
- **Color**: `1f77b4` (dark blue)
- **Description**: Related to CLI commands and interfaces
- **Usage**: Issues with specific walgo commands

### `internal/config` ⚙️
- **Color**: `ff7f0e` (orange)
- **Description**: Configuration management and parsing
- **Usage**: Issues with walgo.yaml, configuration loading

### `internal/hugo` 🏗️
- **Color**: `2ca02c` (green)
- **Description**: Hugo integration functionality
- **Usage**: Issues with Hugo site building, serving, content creation

### `internal/walrus` 🌊
- **Color**: `d62728` (red)
- **Description**: Walrus Sites integration
- **Usage**: Issues with deployment, site updates, status checking

### `internal/obsidian` 📝
- **Color**: `9467bd` (purple)
- **Description**: Obsidian import functionality
- **Usage**: Issues with vault imports, wikilink conversion

### `internal/optimizer` ⚡
- **Color**: `8c564b` (brown)
- **Description**: Asset optimization features
- **Usage**: Issues with HTML/CSS/JS minification and optimization

## Status Labels

### `in progress` 🚧
- **Color**: `fbca04` (yellow)
- **Description**: Currently being worked on
- **Usage**: Issues that have been assigned and are in development

### `needs investigation` 🔍
- **Color**: `b60205` (dark red)
- **Description**: Requires further analysis to understand the issue
- **Usage**: Bug reports that need reproduction or investigation

### `blocked` 🚫
- **Color**: `000000` (black)
- **Description**: Cannot proceed due to external dependencies
- **Usage**: Issues waiting on upstream fixes, external tools, etc.

### `duplicate` 🔄
- **Color**: `cfd3d7` (gray)
- **Description**: This issue or pull request already exists
- **Usage**: Mark duplicate issues before closing

### `wontfix` ❌
- **Color**: `ffffff` (white)
- **Description**: This will not be worked on
- **Usage**: Issues that are decided against or out of scope

### `invalid` ⚠️
- **Color**: `e4e669` (light yellow)
- **Description**: This doesn't seem right
- **Usage**: Issues that are not actually issues or are incorrectly reported

## Platform Labels

### `os: windows` 🪟
- **Color**: `0078d4` (blue)
- **Description**: Windows-specific issues
- **Usage**: Issues that only occur on Windows

### `os: macos` 🍎
- **Color**: `000000` (black)
- **Description**: macOS-specific issues
- **Usage**: Issues that only occur on macOS

### `os: linux` 🐧
- **Color**: `ffa500` (orange)
- **Description**: Linux-specific issues
- **Usage**: Issues that only occur on Linux distributions

## Creating Labels

To create these labels manually in your GitHub repository:

1. Go to your repository on GitHub
2. Click on "Issues" tab
3. Click on "Labels" 
4. Click "New label" for each label below
5. Copy the name, description, and color code

Alternatively, you can use the GitHub CLI:

```bash
# Install GitHub CLI if needed
brew install gh  # macOS
# or follow instructions at https://cli.github.com/

# Authenticate
gh auth login

# Create labels (run these commands in your repository)
gh label create "bug" --description "Something isn't working" --color "d73a4a"
gh label create "enhancement" --description "New feature or request" --color "a2eeef"
gh label create "documentation" --description "Improvements or additions to documentation" --color "0075ca"
gh label create "question" --description "Further information is requested" --color "d876e3"
gh label create "good first issue" --description "Good for newcomers" --color "7057ff"
gh label create "help wanted" --description "Extra attention is needed" --color "008672"
gh label create "priority: high" --description "Critical issues" --color "ff6b6b"
gh label create "priority: medium" --description "Important issues" --color "ffa500"
gh label create "priority: low" --description "Nice to have" --color "0e8a16"
gh label create "in progress" --description "Currently being worked on" --color "fbca04"
gh label create "needs investigation" --description "Requires further analysis" --color "b60205"
gh label create "blocked" --description "Cannot proceed due to dependencies" --color "000000"
gh label create "cmd" --description "CLI commands and interfaces" --color "1f77b4"
gh label create "internal/config" --description "Configuration management" --color "ff7f0e"
gh label create "internal/hugo" --description "Hugo integration" --color "2ca02c"
gh label create "internal/walrus" --description "Walrus Sites integration" --color "d62728"
gh label create "internal/obsidian" --description "Obsidian import functionality" --color "9467bd"
gh label create "internal/optimizer" --description "Asset optimization" --color "8c564b"
```

## Label Usage Guidelines

### For Issues
- Always add at least one **type** label (bug, enhancement, documentation, question)
- Add **priority** labels for bugs and important enhancements
- Add **component** labels to help with organization
- Use **difficulty** labels to help contributors find appropriate issues
- Update **status** labels as work progresses

### For Pull Requests
- Add relevant **component** labels based on what's being changed
- Add **type** labels to indicate the nature of the change
- Remove **status** labels from related issues when PR is merged

### Label Maintenance
- Review and clean up labels periodically
- Archive or remove unused labels
- Keep label descriptions up to date
- Ensure consistent color coding within categories 
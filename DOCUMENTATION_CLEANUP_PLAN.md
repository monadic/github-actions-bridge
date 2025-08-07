# Documentation Cleanup Plan

## Overview
This plan addresses the confusion between ConfigHub workflow (`cub`) and local development workflow (`cub-worker-actions` → `cub-local-actions`).

## Major Pain Points Identified

1. **Unclear Tool Purpose**: `cub-worker-actions` name suggests it's related to ConfigHub workers
2. **Mixed Workflows**: Documentation mixes ConfigHub and local workflows without clear distinction
3. **Inconsistent Commands**: Different files show different commands for same tasks
4. **Missing Guidance**: No clear "when to use which" documentation
5. **Example Confusion**: Examples don't specify which tool they work with
6. **Incomplete Documentation**: USER_GUIDE ignores local development entirely

## Proposed Solutions

### 1. Tool Renaming
- Rename `cub-worker-actions` → `cub-local-actions`
- Update all references in:
  - Source code files
  - Documentation
  - Examples
  - Makefiles
  - Docker configurations

### 2. Clear Workflow Separation
Restructure documentation to clearly separate:
- **ConfigHub Workflow**: Production use with `cub` CLI
- **Local Development Workflow**: Testing with `cub-local-actions`

### 3. Examples Compatibility Table
Create a table showing which examples work with:
- `cub` (ConfigHub) only
- `cub-local-actions` (local) only  
- Both tools

### 4. Documentation Structure

#### README.md
- Add "Two Ways to Use This Project" section upfront
- Separate "ConfigHub Workflow" and "Local Development" sections
- Clear decision tree: "Which tool should I use?"

#### USER_GUIDE.md
- Add "Local Development" section
- Show both workflows for common tasks
- Clear guidance on transitioning from local to ConfigHub

#### CLI_REFERENCE.md
- Rename to match new tool name
- Add comparison with ConfigHub workflow

### 5. Consistency Fixes
- Ensure all examples use correct tool
- Fix command inconsistencies
- Remove outdated references

### 6. AI Hallucination Detection
Look for and fix:
- Non-existent features
- Incorrect command syntax
- Wrong file paths
- Imaginary configuration options
- Incorrect version numbers or dates

## Implementation Steps

1. **Rename Tool** (Priority: Highest)
   - Update source code
   - Update build scripts
   - Update all documentation

2. **Create Compatibility Table** (Priority: High)
   - Test each example with both tools
   - Document which work where

3. **Restructure Documentation** (Priority: High)
   - Clear workflow separation
   - Consistent command usage
   - Remove confusion

4. **Review and Fix** (Priority: Medium)
   - Consistency check
   - Hallucination detection
   - Final cleanup

## Success Criteria
- New users immediately understand the two workflows
- Clear guidance on which tool to use when
- No conflicting information between documents
- All examples clearly state tool compatibility
- No AI-generated errors or hallucinations
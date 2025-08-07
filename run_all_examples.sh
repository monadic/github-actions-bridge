#!/bin/bash

# Script to run all local examples
set -e

echo "🚀 Running all local examples..."
echo "================================"

# List of examples that work locally
examples=(
    "hello-world.yml"
    "environment-variables.yml"
    "with-secrets.yml"
    "build-test-deploy.yml"
    "multi-job.yml"
    "conditional-execution.yml"
    "matrix-builds.yml"
    "docker-compose-improved.yml"
    "artifact-handling-improved.yml"
    "file-persistence-improved.yml"
    "gitops-preview-improved.yml"
)

# Create a simple secrets file for examples that need it
cat > test-secrets.env << EOF
API_KEY=test-api-key-12345
DATABASE_URL=postgresql://test:test@localhost/testdb
GITHUB_TOKEN=ghp_test_token_12345
EOF

# Counter for results
passed=0
failed=0

# Run each example
for example in "${examples[@]}"; do
    echo ""
    echo "📋 Running: $example"
    echo "-------------------"
    
    if [[ "$example" == "with-secrets.yml" ]]; then
        # Run with secrets file
        if ./bin/cub-local-actions run "examples/$example" --secrets-file test-secrets.env > /tmp/example_output.log 2>&1; then
            echo "✅ PASSED: $example"
            ((passed++))
        else
            echo "❌ FAILED: $example"
            echo "Error output:"
            tail -20 /tmp/example_output.log
            ((failed++))
        fi
    else
        # Run without secrets
        if ./bin/cub-local-actions run "examples/$example" > /tmp/example_output.log 2>&1; then
            echo "✅ PASSED: $example"
            ((passed++))
        else
            echo "❌ FAILED: $example"
            echo "Error output:"
            tail -20 /tmp/example_output.log
            ((failed++))
        fi
    fi
done

# Clean up
rm -f test-secrets.env
rm -f /tmp/example_output.log

# Summary
echo ""
echo "================================"
echo "📊 Results Summary"
echo "================================"
echo "✅ Passed: $passed"
echo "❌ Failed: $failed"
echo "📋 Total: $((passed + failed))"

if [ $failed -eq 0 ]; then
    echo ""
    echo "🎉 All examples passed!"
    exit 0
else
    echo ""
    echo "⚠️  Some examples failed. Check the output above."
    exit 1
fi
#!/bin/bash
# GITHUB_TOKEN variable should be set when running this script
if [ -z "$GITHUB_TOKEN" ]; then
    echo "GITHUB_TOKEN environment variable is not set"
    echo "See https://docs.github.com/en/code-security/how-tos/find-and-fix-code-vulnerabilities/scan-from-the-command-line/publish-and-use-packs#authenticating-to-github-container-registries"
    exit 1
fi
codeql pack publish

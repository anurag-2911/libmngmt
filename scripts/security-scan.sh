#!/bin/bash

# Security Scanner Script
# Scans for potentially exposed secrets in the repository

echo "🔍 Scanning for exposed secrets..."

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Function to check for secrets
check_secrets() {
    local issues_found=0
    
    echo "Checking for hardcoded passwords..."
    if grep -r "password.*=" --include="*.go" --include="*.yml" --include="*.yaml" . | grep -v "example\|test\|README\|%s.*password\|password.*%s\|Password.*string"; then
        echo -e "${RED}❌ Found hardcoded passwords${NC}"
        issues_found=1
    fi
    
    echo "Checking for API keys..."
    if grep -rE "(api_key|apikey|api-key)" --include="*.go" --include="*.yml" --include="*.yaml" . | grep -v "example\|test\|README"; then
        echo -e "${RED}❌ Found potential API keys${NC}"
        issues_found=1
    fi
    
    echo "Checking for tokens..."
    if grep -rE "(token|secret)" --include="*.go" --include="*.yml" --include="*.yaml" . | grep -v "example\|test\|README\|GITHUB_TOKEN\|RequestIDKey\|contextKey\|TokenResponse\|GetRequestID"; then
        echo -e "${RED}❌ Found potential tokens/secrets${NC}"
        issues_found=1
    fi
    
    echo "Checking for .env files..."
    if find . -name ".env" -type f | grep -v ".env.example"; then
        echo -e "${RED}❌ Found .env files that should not be committed${NC}"
        issues_found=1
    fi
    
    echo "Checking for private keys..."
    if find . -name "*.pem" -o -name "*.key" -o -name "*.p12" -type f | grep -v "nginx.conf"; then
        echo -e "${RED}❌ Found private key files${NC}"
        issues_found=1
    fi
    
    echo "Checking for common secret patterns..."
    secrets_found=$(grep -rE "(ghp_|gho_|ghu_|ghs_|ghr_|pk_test_|pk_live_|sk_test_|sk_live_|xoxb-)" --exclude="security-scan.sh" . 2>/dev/null)
    if [ ! -z "$secrets_found" ]; then
        echo -e "${RED}❌ Found potential secret tokens${NC}"
        echo "$secrets_found"
        issues_found=1
    fi
    
    return $issues_found
}

# Run security checks
check_secrets

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ No secrets found in repository${NC}"
    exit 0
else
    echo -e "${RED}❌ Security issues detected! Please review and fix the above issues.${NC}"
    echo -e "${YELLOW}💡 Tip: Add sensitive files to .gitignore and use environment variables instead${NC}"
    exit 1
fi

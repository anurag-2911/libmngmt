#!/bin/bash

# Library Management API Test Script
# This script tests all the API endpoints with various scenarios

BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api"

echo "🚀 Testing Library Management API"
echo "================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local method="${2:-GET}"
    local url="$3"
    local data="$4"
    local expected_status="$5"
    
    echo -e "\n${YELLOW}Test: $test_name${NC}"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    
    local temp_file=$(mktemp)
    local status
    
    if [ "$method" = "GET" ]; then
        status=$(curl -s -o "$temp_file" -w "%{http_code}" "$url")
    elif [ "$method" = "POST" ]; then
        status=$(curl -s -o "$temp_file" -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$data" "$url")
    elif [ "$method" = "PUT" ]; then
        status=$(curl -s -o "$temp_file" -w "%{http_code}" -X PUT -H "Content-Type: application/json" -d "$data" "$url")
    elif [ "$method" = "DELETE" ]; then
        status=$(curl -s -o "$temp_file" -w "%{http_code}" -X DELETE "$url")
    fi
    
    local body=$(cat "$temp_file")
    rm "$temp_file"
    
    echo "Method: $method $url"
    echo "Response: $body"
    echo "Status: $status"
    
    if [ "$status" = "$expected_status" ]; then
        echo -e "${GREEN}✅ PASSED${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}❌ FAILED (Expected: $expected_status, Got: $status)${NC}"
    fi
}

echo -e "\n📋 Starting API Tests..."

# Health Check
run_test "Health Check" "GET" "$BASE_URL/health" "" "200"

# Create Books
run_test "Create Book - The Hobbit" "POST" "$API_URL/books" '{
    "title": "The Hobbit",
    "author": "J.R.R. Tolkien",
    "isbn": "9780547928210",
    "publisher": "Houghton Mifflin Harcourt",
    "genre": "Fantasy",
    "published_at": "1937-09-21T00:00:00Z",
    "pages": 366,
    "language": "English"
}' "201"

run_test "Create Book - Clean Code" "POST" "$API_URL/books" '{
    "title": "Clean Code",
    "author": "Robert C. Martin",
    "isbn": "9780132350884",
    "publisher": "Prentice Hall",
    "genre": "Programming",
    "published_at": "2008-08-01T00:00:00Z",
    "pages": 464,
    "language": "English"
}' "201"

# Test Duplicate ISBN
run_test "Create Book - Duplicate ISBN (Should Fail)" "POST" "$API_URL/books" '{
    "title": "Another Hobbit",
    "author": "Someone Else",
    "isbn": "9780547928210",
    "pages": 300
}' "409"

# Test Validation Errors
run_test "Create Book - Missing Title (Should Fail)" "POST" "$API_URL/books" '{
    "author": "Test Author",
    "isbn": "9781234567890",
    "pages": 100
}' "400"

# Get All Books
run_test "Get All Books" "GET" "$API_URL/books" "" "200"

# Get Books with Filters
run_test "Get Books - Filter by Author" "GET" "$API_URL/books?author=tolkien" "" "200"

run_test "Get Books - Filter by Genre" "GET" "$API_URL/books?genre=programming" "" "200"

# Get specific book (we'll need to extract an ID first)
echo -e "\n${YELLOW}Getting book ID for individual book tests...${NC}"
BOOK_ID=$(curl -s "$API_URL/books" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Using Book ID: $BOOK_ID"

if [ -n "$BOOK_ID" ]; then
    run_test "Get Book by ID" "GET" "$API_URL/books/$BOOK_ID" "" "200"
    
    # Update Book
    run_test "Update Book - Mark as Unavailable" "PUT" "$API_URL/books/$BOOK_ID" '{
        "available": false
    }' "200"
    
    # Delete Book
    run_test "Delete Book" "DELETE" "$API_URL/books/$BOOK_ID" "" "200"
    
    run_test "Get Deleted Book (Should Fail)" "GET" "$API_URL/books/$BOOK_ID" "" "404"
fi

# Test Invalid Book ID
run_test "Get Book - Invalid ID (Should Fail)" "GET" "$API_URL/books/invalid-id" "" "400"

run_test "Get Book - Non-existent ID (Should Fail)" "GET" "$API_URL/books/123e4567-e89b-12d3-a456-426614174000" "" "404"

# Summary
echo -e "\n📊 Test Results"
echo "==============="
echo -e "Tests Run: $TESTS_RUN"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$((TESTS_RUN - TESTS_PASSED))${NC}"

if [ $TESTS_PASSED -eq $TESTS_RUN ]; then
    echo -e "\n🎉 ${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n❌ ${RED}Some tests failed!${NC}"
    exit 1
fi

# Test Health Check
make_request "Health Check" "$API_BASE/health"

# Test Create Book 1
echo "📖 Creating Book 1: The Hobbit"
book1_response=$(curl -s -X POST "$API_BASE/api/books" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Hobbit",
    "author": "J.R.R. Tolkien",
    "isbn": "9780547928210",
    "publisher": "Houghton Mifflin Harcourt",
    "genre": "Fantasy",
    "published_at": "1937-09-21T00:00:00Z",
    "pages": 366,
    "language": "English"
  }')

echo "$book1_response" | jq '.'
book1_id=$(echo "$book1_response" | jq -r '.data.id')

# Test Create Book 2
echo ""
echo "📖 Creating Book 2: 1984"
book2_response=$(curl -s -X POST "$API_BASE/api/books" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "1984",
    "author": "George Orwell",
    "isbn": "9780451524935",
    "publisher": "Signet Classic",
    "genre": "Dystopian Fiction",
    "published_at": "1949-06-08T00:00:00Z",
    "pages": 328,
    "language": "English"
  }')

echo "$book2_response" | jq '.'
book2_id=$(echo "$book2_response" | jq -r '.data.id')

# Test Get All Books
make_request "Get All Books" "$API_BASE/api/books" | jq '.'

# Test Get Book by ID
if [ "$book1_id" != "null" ]; then
    make_request "Get Book by ID (Book 1)" "$API_BASE/api/books/$book1_id" | jq '.'
fi

# Test Update Book
if [ "$book1_id" != "null" ]; then
    echo ""
    echo "✏️  Updating Book 1"
    echo "---"
    update_response=$(curl -s -X PUT "$API_BASE/api/books/$book1_id" \
      -H "Content-Type: application/json" \
      -d '{
        "available": false,
        "title": "The Hobbit: There and Back Again"
      }')
    echo "$update_response" | jq '.'
fi

# Test Filtering
make_request "Filter by Author (Tolkien)" "$API_BASE/api/books?author=tolkien" | jq '.'
make_request "Filter by Genre (Fantasy)" "$API_BASE/api/books?genre=fantasy" | jq '.'
make_request "Filter by Availability (false)" "$API_BASE/api/books?available=false" | jq '.'

# Test Error Cases
echo ""
echo "❌ Testing Error Cases"
echo "---"

# Invalid ID
make_request "Get Book with Invalid ID" "$API_BASE/api/books/invalid-id"

# Duplicate ISBN
echo ""
echo "Trying to create book with duplicate ISBN:"
duplicate_response=$(curl -s -X POST "$API_BASE/api/books" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Another Hobbit",
    "author": "Someone Else",
    "isbn": "9780547928210",
    "publisher": "Another Publisher",
    "genre": "Fantasy",
    "published_at": "2023-01-01T00:00:00Z",
    "pages": 100,
    "language": "English"
  }')
echo "$duplicate_response" | jq '.'

# Test Delete Book
if [ "$book2_id" != "null" ]; then
    echo ""
    echo "🗑️  Deleting Book 2"
    echo "---"
    delete_response=$(curl -s -X DELETE "$API_BASE/api/books/$book2_id")
    echo "$delete_response" | jq '.'
fi

# Final book list
make_request "Final Book List" "$API_BASE/api/books" | jq '.'

echo "✅ API Testing Complete!"

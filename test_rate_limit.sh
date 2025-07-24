#!/bin/bash

echo "Testing rate limiting..."

# Fill up the rate limiter (capacity is 100)
for i in {1..100}; do
    curl -s http://localhost:8080/api/books > /dev/null &
done

# Wait a moment for all background requests to start
sleep 0.2

# Now try to create a book - this should be rate limited
echo "Trying to create book when rate limiter is full:"
curl -i -X POST http://localhost:8080/api/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Should Be Rate Limited",
    "author": "Test Author",
    "isbn": "9781234567777",
    "publisher": "Test Publisher",
    "genre": "Fiction",
    "pages": 250,
    "language": "English"
  }'

# Wait for background processes to complete
wait
echo -e "\nRate limiting test completed."

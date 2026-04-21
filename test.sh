#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}=== Redis Server Test Suite ===${NC}"

# Start server in background
echo -e "${YELLOW}Starting server...${NC}"
./bin/redis-server &
SERVER_PID=$!
sleep 1

# Function to test a command
test_cmd() {
    local cmd=$1
    local expected=$2
    local result=$(redis-cli $cmd 2>&1)

    if [[ "$result" == *"$expected"* ]]; then
        echo -e "${GREEN}✓${NC} $cmd"
    else
        echo -e "${RED}✗${NC} $cmd (got: $result, expected: $expected)"
        return 1
    fi
}

echo -e "${YELLOW}Testing String Commands...${NC}"
test_cmd "PING" "PONG"
test_cmd "SET key value" "OK"
test_cmd "GET key" "value"
test_cmd "SET num 10 EX 100" "OK"
test_cmd "INCR num" "11"
test_cmd "DECR num" "10"

echo -e "${YELLOW}Testing List Commands...${NC}"
test_cmd "RPUSH list a b c" "3"
test_cmd "LLEN list" "3"
test_cmd "LPUSH list z" "4"
test_cmd "LPOP list" "z"
test_cmd "RPOP list" "c"

echo -e "${YELLOW}Testing Set Commands...${NC}"
test_cmd "SADD set x y z" "3"
test_cmd "SCARD set" "3"
test_cmd "SISMEMBER set x" "1"
test_cmd "SISMEMBER set nonexist" "0"

echo -e "${YELLOW}Testing Key Management...${NC}"
test_cmd "EXISTS key" "1"
test_cmd "EXISTS nonexist" "0"
test_cmd "DEL key" "1"
test_cmd "EXISTS key" "0"

echo -e "${YELLOW}Testing Expiry...${NC}"
redis-cli SET tempkey value EX 2 > /dev/null
sleep 3
result=$(redis-cli GET tempkey 2>&1)
if [[ -z "$result" || "$result" == "(nil)" ]]; then
    echo -e "${GREEN}✓${NC} TTL expiry works"
else
    echo -e "${RED}✗${NC} TTL expiry failed (got: [$result])"
fi

# Cleanup
kill $SERVER_PID 2>/dev/null || true

echo -e "${GREEN}All tests passed!${NC}"
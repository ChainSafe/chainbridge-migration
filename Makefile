.PHONY: help stop-bridge
all: help

stop-bridge:
	go run ./main.go stop-bridge


transfer-tokens:
	go run ./main.go transfer-tokens

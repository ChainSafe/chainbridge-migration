.PHONY: help stop-bridge
all: help

stop-bridge:
	go run ./main.go stop-bridge
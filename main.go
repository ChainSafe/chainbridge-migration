package main

import (
	"bridge-scripts/scripts"
	"bridge-scripts/util"
	"fmt"
	"os"
)

func main() {
	cfgPath := ""
	if len(os.Args) == 3 {
		cfgPath = os.Args[2]
	}

	fmt.Printf("Starting ChainBridge scripts: %s\n", os.Args[1])
	util.DisplayLine()

	// load general config
	config, err := util.GetConfig(cfgPath)
	if err != nil {
		fmt.Printf("Unable to load configuration: %v", err)
		return
	}
	fmt.Println("Successfully loaded configuration!")
	util.DisplayLine()

	// load v1 bridge config
	v1BridgeConfig, err := util.GetV1BridgeConfig(config.ConfigurationPath)
	if err != nil {
		fmt.Println("Error on loading v1BridgeConfig:")
		fmt.Print(err)
	} else {
		fmt.Println("Successfully loaded v1BridgeConfig!")
		util.DisplayLine()
	}

	// run action
	switch os.Args[1] {
	case "stop-bridge":
		err := scripts.PauseBridge(v1BridgeConfig, config)
		if err != nil {
			fmt.Print(err)
		}
		break
	case "transfer-tokens":
		err := scripts.TransferTokens(v1BridgeConfig, config)
		if err != nil {
			fmt.Print(err)
		}
		break
	default:
		fmt.Println("Invalid action")
	}
}

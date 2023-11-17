package main

import (
	"encoding/json"
	"fmt"

	"github.com/dcaravel/acs-secret-mem-eval/analyze"
	"github.com/dcaravel/acs-secret-mem-eval/collect"
)

func main() {
	fmt.Printf("Collecting Secrets, Ctrl+C to analyze\n\n")
	secrets, err := collect.Secrets()
	check(err)

	result, err := analyze.PullSecrets(secrets)
	check(err)

	dataB, err := json.MarshalIndent(result, "", "  ")
	check(err)

	fmt.Printf("\n\nResult:\n%s\n", dataB)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

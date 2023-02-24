package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gofireflyio/aiac/v2/libaiac"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
	if len(os.Args) == 0 {
		os.Args = append(os.Args, "--help")
	}

    client := libaiac.NewClient(apiKey)

	err := client.Ask(
		context.TODO(),
		// NOTE: we are prepending the word "generate" to the prompt, this
		// ensures the language model actually generates code. The word "get",
		// on the other hand, doesn't necessarily result in code being generated.
		fmt.Sprintf("generate %s", strings.Join(os.Args[1:], " ")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %s\n", err)
		os.Exit(1)
	}
}

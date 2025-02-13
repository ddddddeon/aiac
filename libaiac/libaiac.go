package libaiac

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ido50/requests"
    "github.com/atotto/clipboard"
)

// Client is a structure used to continuously generate IaC code via OpenAPI/ChatGPT
type Client struct {
	*requests.HTTPClient
	apiKey string
}

// NewClient creates a new instance of the Client struct, with the provided
// input options. Neither the OpenAI API nor ChatGPT are yet contacted at this
// point.
func NewClient(apiKey string) *Client {
	if apiKey == "" {
		return nil
	}

	cli := &Client{
		apiKey: strings.TrimPrefix(apiKey, "Bearer "),
	}

	cli.HTTPClient = requests.NewClient("https://api.openai.com/v1").
		Accept("application/json").
		Header("Authorization", fmt.Sprintf("Bearer %s", cli.apiKey)).
		ErrorHandler(func(
			httpStatus int,
			contentType string,
			body io.Reader,
		) error {
			var res struct {
				Error struct {
					Message string `json:"message"`
					Type    string `json:"type"`
				} `json:"error"`
			}

			err := json.NewDecoder(body).Decode(&res)
			if err != nil {
				return fmt.Errorf(
					"OpenAI returned response %s",
					http.StatusText(httpStatus),
				)
			}

			return fmt.Errorf("[%s] %s", res.Error.Type, res.Error.Message)
		})

	return cli
}

// Ask asks the OpenAI API to generate code based on the provided prompt.
// It is only meant to be used in command line applications (see GenerateCode
// for library usage). The generated code will always be printed to standard
// output, but may optionally be stored in the file whose path is provided by
// the outputPath argument. To only print to standard output, provide an empty
// string or a dash ("-") via outputPath. If shouldRetry is true, you will be
// prompted whether to regenerate the response after it is printed to standard output,
// in case you are unhappy with the response. If shouldQuit is true, the code
// is printed to standard output and the function returns, without storing to a
// file or asking whether to regenerate the response.
func (client *Client) Ask(
	ctx context.Context,
	prompt string,
) (err error) {
	code, err := client.GenerateCode(ctx, prompt)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, code)

    // Copy generated code to clipboard
    clipboard.WriteAll(code)

	return nil
}

// GenerateCode sends the provided prompt to the OpenAI API and returns the
// generated code.
func (client *Client) GenerateCode(ctx context.Context, prompt string) (
	code string,
	err error,
) {
	var answer struct {
		Choices []struct {
			Text         string `json:"text"`
			Index        int64  `json:"index"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	var status int
	err = client.NewRequest("POST", "/completions").
		JSONBody(map[string]interface{}{
			"model":      "text-davinci-003",
			"prompt":     prompt,
            "temperature": 0.2,
			"max_tokens": 4097 - len(prompt),
		}).
		Into(&answer).
		StatusInto(&status).
		RunContext(ctx)
	if err != nil {
		return code, fmt.Errorf("failed sending prompt: %w", err)
	}

	if len(answer.Choices) == 0 {
		return code, fmt.Errorf("no results returned from API")
	}

	if answer.Choices[0].FinishReason != "stop" {
		return code, fmt.Errorf(
			"result was truncated by API due to %s",
			answer.Choices[0].FinishReason,
		)
	}

	return strings.TrimSpace(answer.Choices[0].Text), nil
}

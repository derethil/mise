// Package testutil provides deterministic Genkit test doubles for AI features.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"

	genai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type ModelResponse struct {
	Text string
	Err  error
}

func TextResponse(text string) ModelResponse {
	return ModelResponse{Text: text}
}

func JSONResponse(value any) ModelResponse {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ErrorResponse(err)
	}

	return TextResponse(string(encoded))
}

func ErrorResponse(err error) ModelResponse {
	return ModelResponse{Err: err}
}

func DefineModel(g *genkit.Genkit, name string, responses ...ModelResponse) *genai.ModelAction {
	call := 0
	return genkit.DefineModelAction(g, name, &genai.ModelOptions{
		Supports: &genai.ModelSupports{Multiturn: true, Tools: true, SystemRole: true},
	}, func(context.Context, *genai.ModelRequest, *genai.GenerationCommonConfig, genai.ModelStreamCallback) (*genai.ModelResponse, error) {
		if call >= len(responses) {
			return nil, fmt.Errorf("fake model %q called %d times, but has %d responses", name, call+1, len(responses))
		}

		response := responses[call]
		call++
		if response.Err != nil {
			return nil, response.Err
		}

		return &genai.ModelResponse{
			FinishReason: genai.FinishReasonStop,
			Message:      genai.NewModelTextMessage(response.Text),
		}, nil
	})
}

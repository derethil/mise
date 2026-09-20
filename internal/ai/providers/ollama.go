package providers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/derethil/mise/internal/config"
	genai "github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/ollama/ollama/api"
)

var ErrPullDeclined = errors.New("model download declined")
var ErrClearDeclined = errors.New("model deletion declined")
var ErrOllamaUnavailable = errors.New("ollama unavailable")

type OllamaProvider struct {
	Provider
	client *api.Client
}

type ConfirmFunc func(question string) (bool, error)

type ModelStatus struct {
	Model ModelRef
	Info  *ModelInfo
}

type ModelInfo struct {
	Name              string
	Size              int64
	ModifiedAt        time.Time
	Family            string
	ParameterSize     string
	QuantizationLevel string
	Capabilities      []string
}

type PullProgress struct {
	Status           string
	Total, Completed int64
}

type PullProgressFunc func(PullProgress) error

func NewOllamaProvider(cfg config.ProviderConfig) (*OllamaProvider, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("%w: providers.ollama.base_url is not set", config.ErrInvalidConfig)
	}

	base, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse providers.ollama.base_url: %w", err)
	}

	return &OllamaProvider{
		Provider: Provider{
			Name:           ProviderOllama,
			Config:         cfg,
			Plugin:         &ollama.Ollama{ServerAddress: cfg.BaseURL, Timeout: cfg.Timeout},
			GenerateConfig: ollamaConfig,
			Middleware:     ollamaMiddleware,
		},
		client: api.NewClient(base, http.DefaultClient),
	}, nil
}

func CheckOllama(ctx context.Context, cfg config.ProviderConfig) error {
	provider, err := NewOllamaProvider(cfg)
	if err != nil {
		return err
	}

	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if _, err := provider.client.List(checkCtx); err != nil {
		return fmt.Errorf("%w: %w", ErrOllamaUnavailable, err)
	}

	return nil
}

func ollamaConfig(cfg GenerateConfig) any {
	out := &ollama.GenerateContentConfig{Temperature: cfg.Temperature, TopP: cfg.TopP}
	if cfg.Reasoning != nil {
		out.Think = ollama.ThinkEnabled(*cfg.Reasoning)
	}
	return out
}

func ollamaMiddleware(GenerateConfig) []genai.Middleware {
	return []genai.Middleware{stripThinkArtifacts}
}

func (c *OllamaProvider) models(ctx context.Context) ([]ModelInfo, error) {
	response, err := c.client.List(ctx)
	if err != nil {
		return nil, err
	}

	models := make([]ModelInfo, len(response.Models))

	for i, model := range response.Models {
		capabilities := make([]string, len(model.Capabilities))
		for j, capability := range model.Capabilities {
			capabilities[j] = capability.String()
		}

		models[i] = ModelInfo{
			Name:              model.Name,
			Size:              model.Size,
			ModifiedAt:        model.ModifiedAt,
			Family:            model.Details.Family,
			ParameterSize:     model.Details.ParameterSize,
			QuantizationLevel: model.Details.QuantizationLevel,
			Capabilities:      capabilities,
		}
	}

	return models, nil
}

func (c *OllamaProvider) hasModel(ctx context.Context, model ModelRef) (bool, error) {
	models, err := c.models(ctx)
	if err != nil {
		return false, err
	}

	name := ModelName(model)

	return slices.ContainsFunc(models, func(m ModelInfo) bool { return m.Name == name }), nil
}

func (c *OllamaProvider) pullModel(ctx context.Context, model ModelRef, onProgress PullProgressFunc) error {
	request := api.PullRequest{
		Model: ModelName(model),
	}

	pullProgress := func(progress api.ProgressResponse) error {
		return onProgress(PullProgress{Status: progress.Status, Total: progress.Total, Completed: progress.Completed})
	}

	slog.DebugContext(ctx, "pulling ollama model", slog.String("model", request.Model))

	err := c.client.Pull(ctx, &request, pullProgress)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "pulled ollama model", slog.String("model", request.Model))

	return nil
}

func (c *OllamaProvider) deleteModel(ctx context.Context, name string) error {
	slog.DebugContext(ctx, "deleting ollama model", slog.String("model", name))
	return c.client.Delete(ctx, &api.DeleteRequest{Model: name})
}

func (c *OllamaProvider) installedModels(ctx context.Context, models []ModelRef) (map[string]ModelInfo, error) {
	if !slices.ContainsFunc(models, func(m ModelRef) bool { return m.Provider == ProviderOllama }) {
		return nil, nil
	}

	available, err := c.models(ctx)
	if err != nil {
		return nil, err
	}

	installed := make(map[string]ModelInfo, len(available))
	for _, model := range available {
		installed[model.Name] = model
	}

	return installed, nil
}

func (c *OllamaProvider) Clear(ctx context.Context, keep []ModelRef, confirm ConfirmFunc) ([]ModelInfo, error) {
	installed, err := c.models(ctx)
	if err != nil {
		return nil, err
	}

	stale := staleModels(installed, keep)
	if len(stale) == 0 {
		slog.DebugContext(ctx, "no stale ollama models found")
		return nil, nil
	}

	slog.DebugContext(ctx, "found stale ollama models", slog.Int("count", len(stale)))

	names := make([]string, len(stale))
	for i, model := range stale {
		names[i] = model.Name
	}

	ok, err := confirm(fmt.Sprintf("Delete %d model(s) (%s)?", len(stale), strings.Join(names, ", ")))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrClearDeclined
	}

	for _, model := range stale {
		if err := c.deleteModel(ctx, model.Name); err != nil {
			return nil, fmt.Errorf("failed to delete %s: %w", model.Name, err)
		}
	}

	return stale, nil
}

func (c *OllamaProvider) Statuses(ctx context.Context, models []ModelRef) ([]ModelStatus, error) {
	installed, err := c.installedModels(ctx, models)
	if err != nil {
		return nil, err
	}

	statuses := make([]ModelStatus, len(models))
	for i, model := range models {
		statuses[i] = ModelStatus{Model: model}

		if model.Provider != ProviderOllama {
			continue
		}

		if info, ok := installed[ModelName(model)]; ok {
			statuses[i].Info = &info
		}
	}

	return statuses, nil
}

func (c *OllamaProvider) Ensure(ctx context.Context, model ModelRef, confirm ConfirmFunc, onProgress PullProgressFunc) error {
	has, err := c.hasModel(ctx, model)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	slog.DebugContext(ctx, "model not found locally", slog.String("model", model.String()))

	ok, err := confirm(fmt.Sprintf("Model %q is not available on your Ollama instance. Download it now?", model))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: %s", ErrPullDeclined, model)
	}

	return c.pullModel(ctx, model, onProgress)

}

func ModelName(model ModelRef) string {
	if model.Provider != ProviderOllama {
		return model.String()
	}

	tag := model.Tag
	if tag == "" {
		tag = "latest"
	}

	return model.Name + ":" + tag
}

func staleModels(installed []ModelInfo, keep []ModelRef) []ModelInfo {
	keepNames := make(map[string]bool, len(keep))
	for _, model := range keep {
		if model.Provider != ProviderOllama {
			continue
		}
		keepNames[ModelName(model)] = true
	}

	var stale []ModelInfo
	for _, model := range installed {
		if !keepNames[model.Name] {
			stale = append(stale, model)
		}
	}

	return stale
}

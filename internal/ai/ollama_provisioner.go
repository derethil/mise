package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/derethil/mise/internal/config"
	"github.com/ollama/ollama/api"
)

var ErrPullDeclined = errors.New("model download declined")

type ConfirmFunc func(question string) (bool, error)

type ModelStatus struct {
	Model ModelRef
	Info  *ModelInfo
}

type OllamaProvisioner struct {
	client *api.Client
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

func NewOllamaProvisioner(baseURL string) (*OllamaProvisioner, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("%w: providers.ollama.base_url is not set", config.ErrInvalidConfig)
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	c := api.NewClient(base, http.DefaultClient)

	return &OllamaProvisioner{client: c}, nil
}

func (c *OllamaProvisioner) Models(ctx context.Context) ([]ModelInfo, error) {
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

func (c *OllamaProvisioner) HasModel(ctx context.Context, model ModelRef) (bool, error) {
	models, err := c.Models(ctx)
	if err != nil {
		return false, err
	}

	name := OllamaModelName(model)

	return slices.ContainsFunc(models, func(m ModelInfo) bool { return m.Name == name }), nil
}

func (c *OllamaProvisioner) PullModel(ctx context.Context, model ModelRef, onProgress PullProgressFunc) error {
	request := api.PullRequest{
		Model: OllamaModelName(model),
	}

	pullProgress := func(progress api.ProgressResponse) error {
		return onProgress(PullProgress{Status: progress.Status, Total: progress.Total, Completed: progress.Completed})
	}

	err := c.client.Pull(ctx, &request, pullProgress)
	if err != nil {
		return err
	}

	return nil
}

func (c *OllamaProvisioner) Statuses(ctx context.Context, models []ModelRef) ([]ModelStatus, error) {
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

		if info, ok := installed[OllamaModelName(model)]; ok {
			statuses[i].Info = &info
		}
	}

	return statuses, nil
}

func (c *OllamaProvisioner) installedModels(ctx context.Context, models []ModelRef) (map[string]ModelInfo, error) {
	if !slices.ContainsFunc(models, func(m ModelRef) bool { return m.Provider == ProviderOllama }) {
		return nil, nil
	}

	available, err := c.Models(ctx)
	if err != nil {
		return nil, err
	}

	installed := make(map[string]ModelInfo, len(available))
	for _, model := range available {
		installed[model.Name] = model
	}

	return installed, nil
}

func (c *OllamaProvisioner) Ensure(ctx context.Context, model ModelRef, confirm ConfirmFunc, onProgress PullProgressFunc) error {
	has, err := c.HasModel(ctx, model)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	ok, err := confirm(fmt.Sprintf("Model %q is not available on your Ollama instance. Download it now?", model))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: %s", ErrPullDeclined, model)
	}

	return c.PullModel(ctx, model, onProgress)

}

func OllamaModelName(model ModelRef) string {
	if model.Provider != ProviderOllama {
		return model.String()
	}

	tag := model.Tag
	if tag == "" {
		tag = "latest"
	}

	return model.Name + ":" + tag
}

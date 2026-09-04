package ollama

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/derethil/mise/internal/ai"
	"github.com/derethil/mise/internal/config"
	"github.com/ollama/ollama/api"
)

var ErrPullDeclined = errors.New("model download declined")
var ErrClearDeclined = errors.New("model deletion declined")

type ConfirmFunc func(question string) (bool, error)

type ModelStatus struct {
	Model ai.ModelRef
	Info  *ModelInfo
}

type Provisioner struct {
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

func NewProvisioner(baseURL string) (*Provisioner, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("%w: providers.ollama.base_url is not set", config.ErrInvalidConfig)
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse providers.ollama.base_url: %w", err)
	}

	c := api.NewClient(base, http.DefaultClient)

	return &Provisioner{client: c}, nil
}

func (c *Provisioner) models(ctx context.Context) ([]ModelInfo, error) {
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

func (c *Provisioner) hasModel(ctx context.Context, model ai.ModelRef) (bool, error) {
	models, err := c.models(ctx)
	if err != nil {
		return false, err
	}

	name := ModelName(model)

	return slices.ContainsFunc(models, func(m ModelInfo) bool { return m.Name == name }), nil
}

func (c *Provisioner) pullModel(ctx context.Context, model ai.ModelRef, onProgress PullProgressFunc) error {
	request := api.PullRequest{
		Model: ModelName(model),
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

func (c *Provisioner) deleteModel(ctx context.Context, name string) error {
	return c.client.Delete(ctx, &api.DeleteRequest{Model: name})
}

func (c *Provisioner) installedModels(ctx context.Context, models []ai.ModelRef) (map[string]ModelInfo, error) {
	if !slices.ContainsFunc(models, func(m ai.ModelRef) bool { return m.Provider == ai.ProviderOllama }) {
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

func (c *Provisioner) Clear(ctx context.Context, keep []ai.ModelRef, confirm ConfirmFunc) ([]ModelInfo, error) {
	installed, err := c.models(ctx)
	if err != nil {
		return nil, err
	}

	stale := staleModels(installed, keep)
	if len(stale) == 0 {
		return nil, nil
	}

	names := make([]string, len(stale))
	for i, model := range stale {
		names[i] = model.Name
	}

	ok, err := confirm(fmt.Sprintf("Delete %d unused model(s) (%s)?", len(stale), strings.Join(names, ", ")))
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

func (c *Provisioner) Statuses(ctx context.Context, models []ai.ModelRef) ([]ModelStatus, error) {
	installed, err := c.installedModels(ctx, models)
	if err != nil {
		return nil, err
	}

	statuses := make([]ModelStatus, len(models))
	for i, model := range models {
		statuses[i] = ModelStatus{Model: model}

		if model.Provider != ai.ProviderOllama {
			continue
		}

		if info, ok := installed[ModelName(model)]; ok {
			statuses[i].Info = &info
		}
	}

	return statuses, nil
}

func (c *Provisioner) Ensure(ctx context.Context, model ai.ModelRef, confirm ConfirmFunc, onProgress PullProgressFunc) error {
	has, err := c.hasModel(ctx, model)
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

	return c.pullModel(ctx, model, onProgress)

}

func ModelName(model ai.ModelRef) string {
	if model.Provider != ai.ProviderOllama {
		return model.String()
	}

	tag := model.Tag
	if tag == "" {
		tag = "latest"
	}

	return model.Name + ":" + tag
}

func staleModels(installed []ModelInfo, keep []ai.ModelRef) []ModelInfo {
	keepNames := make(map[string]bool, len(keep))
	for _, model := range keep {
		if model.Provider != ai.ProviderOllama {
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

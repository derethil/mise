package ai

import (
	"fmt"
	"reflect"

	"github.com/derethil/mise/internal/ai/providers"
	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/middleware"
)

type Deps struct {
	Tandoor *tandoor.Client
}

type Registry struct {
	Genkit    *genkit.Genkit
	Providers map[string]*providers.Provider
	Model     providers.ModelRef
	Deps
}

func (r Registry) PromptOptions(cfg providers.GenerateConfig, opts ...ai.PromptExecuteOption) []ai.PromptExecuteOption {
	return r.PromptOptionsFor(r.Model, cfg, opts...)
}

func (r Registry) PromptOptionsFor(model providers.ModelRef, cfg providers.GenerateConfig, opts ...ai.PromptExecuteOption) []ai.PromptExecuteOption {
	mw := []ai.Middleware{&middleware.Retry{}}
	provider := r.Providers[model.Provider]

	if cfg != (providers.GenerateConfig{}) && provider != nil {
		if translate := provider.GenerateConfig; translate != nil {
			opts = append(opts, ai.WithConfig(translate(cfg)))
		}

		if factory := provider.Middleware; factory != nil {
			mw = append(mw, factory(cfg)...)
		}
	}

	return append(opts, ai.WithModelName(model.String()), ai.WithUse(mw...))
}

type Feature interface {
	Register(Registry) error
}

var featureFactories []func() Feature

func RegisterFeature(newFeature func() Feature) {
	featureFactories = append(featureFactories, newFeature)
}

func FeatureOf[T Feature](c *Client) (T, error) {
	var zero T

	feature, ok := c.features[reflect.TypeFor[T]()]
	if !ok {
		return zero, fmt.Errorf("%w: %s", ErrFeatureNotRegistered, reflect.TypeFor[T]())
	}

	return feature.(T), nil
}

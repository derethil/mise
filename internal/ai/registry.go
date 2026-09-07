package ai

import (
	"fmt"
	"reflect"

	"github.com/derethil/mise/internal/tandoor"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/middleware"
)

type Deps struct {
	Tandoor *tandoor.Client
}

type Registry struct {
	Genkit   *genkit.Genkit
	Provider string
	Deps
}

func (r Registry) PromptOptions(cfg GenerateConfig, opts ...ai.PromptExecuteOption) []ai.PromptExecuteOption {
	mw := []ai.Middleware{&middleware.Retry{}}

	if cfg != (GenerateConfig{}) {
		if translate, ok := providerConfigFactories[r.Provider]; ok {
			opts = append(opts, ai.WithConfig(translate(cfg)))
		}

		if factory, ok := providerMiddlewareFactories[r.Provider]; ok {
			mw = append(mw, factory(cfg)...)
		}
	}

	return append(opts, ai.WithUse(mw...))
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

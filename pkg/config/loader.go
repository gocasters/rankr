package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Options struct {
	Prefix       string
	Delimiter    string
	Separator    string
	YamlFilePath string
	Transformer  func(key string, value string) (string, any)
}

const (
	defaultDelimiter = "."
	defaultSeparator = "__"
)

func defaultTransformer(k, v, prefix, delimiter, separator string) (string, any) {
	key := strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, prefix)), separator, delimiter)

	return key, v
}

func fillDefaultOptions(options *Options) *Options {
	if options.Delimiter == "" {
		options.Delimiter = defaultDelimiter
	}
	if (options.Separator) == "" {
		options.Separator = defaultSeparator
	}
	if options.Transformer == nil {
		options.Transformer = func(k, v string) (string, any) {
			return defaultTransformer(k, v, options.Prefix, options.Delimiter, options.Separator)
		}
	}

	return options
}

func Load(options Options, config interface{}) error {

	options = *fillDefaultOptions(&options)

	k := koanf.New(options.Delimiter)

	if options.YamlFilePath != "" {
		if err := k.Load(file.Provider(options.YamlFilePath), yaml.Parser()); err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}
	}

	if err := k.Load(env.Provider(options.Delimiter, env.Opt{
		Prefix:        options.Prefix,
		TransformFunc: options.Transformer,
	}), nil); err != nil {
		return fmt.Errorf("error loading environment variables: %w", err)
	}

	fmt.Printf("koanf %+v\n", k)

	if err := k.Unmarshal("", &config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

package config

import (
	"fmt"
	"log"
	"reflect"
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
	key := k
	// Remove prefix and separator if present
	if prefix != "" {
		prefixWithSep := prefix + separator
		if strings.HasPrefix(k, prefixWithSep) {
			key = strings.TrimPrefix(k, prefixWithSep)
		} else if strings.HasPrefix(k, prefix) {
			key = strings.TrimPrefix(k, prefix)
		}
	}

	key = strings.ReplaceAll(strings.ToLower(key), separator, delimiter)

	return key, v
}

func fillDefaultOptions(options *Options) *Options {
	if options.Delimiter == "" {
		options.Delimiter = defaultDelimiter
	}
	if options.Separator == "" {
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
	if config == nil {
		err := fmt.Errorf("config cannot be nil")
		log.Printf("Invalid config: %v", err)
		return err
	}

	if reflect.ValueOf(config).Kind() != reflect.Ptr || reflect.ValueOf(config).Elem().Kind() != reflect.Struct {
		err := fmt.Errorf("config must be a pointer to a struct, but got %v", reflect.TypeOf(config))
		log.Printf("Invalid config type: %v", reflect.TypeOf(config))
		return err
	}

	theOptions := fillDefaultOptions(&options)

	k := koanf.New(theOptions.Delimiter)

	if theOptions.YamlFilePath != "" {
		if err := k.Load(file.Provider(theOptions.YamlFilePath), yaml.Parser()); err != nil {
			log.Printf("Error loading config file: %v", err)
			return err
		}
	}

	if err := k.Load(env.Provider(theOptions.Delimiter, env.Opt{
		Prefix:        theOptions.Prefix,
		TransformFunc: theOptions.Transformer,
	}), nil); err != nil {
		log.Printf("Error loading environment variables: %v", err)
		return err
	}

	fmt.Printf("koanf %+v\n", k)

	if err := k.Unmarshal("", config); err != nil {
		log.Printf("Error unmarshaling config: %v", err)
		return err
	}

	return nil
}

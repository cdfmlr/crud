package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cdfmlr/crud/log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"reflect"
	"strings"
)

var logger = log.ZoneLogger("crud/config")

// Init reads configure, write values into given configModel
//  - FromFile: read config from file
//  - FromEnv: read config from environment variables
//  - WatchFileChange: watch config file and reload config when changed
func Init(configModel any, options ...Option) error {
	for _, option := range options {
		err := option(configModel)
		if err != nil {
			return err
		}
	}
	return nil
}

type Option func(configModel any) error

// FromFile reads config from file at path, and unmarshal to config.
// YAML, JSON, TOML, etc. files are supported.
// This is the recommended way to read config.
func FromFile(path string) Option {
	return func(config any) error {
		err := readFromFile(path, config)
		if err != nil {
			logger.WithError(err).
				Error("Init config FromFile: readFromFile error")
			return err
		}
		return nil
	}
}

// WatchFileChange works with FromFile:
//     var config MyConfig
//     config.Init(&config, FromFile(path), WatchFileChange(hook))
//
// WatchFileChange watch current viper config file,
// and reload config when changed.
//
// Notice: you do not need to reset your `config` variable in the hook,
// we will do it for you. BUT THIS FEATURE MAKE THE `config` NOT THREAD-SAFE.
// It's to be fixed in the future.
func WatchFileChange(hook func(oldConfig any, newConfig any)) Option {
	return func(config any) error {
		err := watch(config, hook)
		if err != nil {
			logger.WithError(err).
				Error("Init config WatchFileChange: watch error")
			return err
		}
		return nil
	}
}

// FromEnv reads config from environment variables, and unmarshal to config
// prefix is the prefix of environment variables:
//     type MyConfig struct {
//         Foo struct {
//             Bar string
//         }
//     }
//     config := MyConfig{}
//     config.Init(&config, FromEnv("MYAPP"))
// will read `config.Foo.Bar` from env `MYAPP_FOO_BAR`.
func FromEnv(prefix string) Option {
	return func(config any) error {
		err := readFromEnv(prefix, config)
		if err != nil {
			logger.WithError(err).
				Error("Init config FromEnv: readFromEnv error")
			return err
		}
		return nil
	}
}

// readFromFile path of config file, and unmarshal to config
func readFromFile(path string, config any) error {
	if !configMustPtrToStruct(config) {
		return ErrConfigNotPtrToStruct
	}

	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return viper.Unmarshal(config)
}

// readFromEnv read config from environment variables, and unmarshal to config
func readFromEnv(prefix string, config any) error {
	if !configMustPtrToStruct(config) {
		return ErrConfigNotPtrToStruct
	}

	// AutomaticEnv would not create keys.
	if len(viper.AllKeys()) == 0 {
		_ = setDefaultStruct(config)
	}

	viper.SetEnvPrefix(prefix)
	// Get("foo.bar") => PREFIX_FOO_BAR
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	return viper.Unmarshal(config)
}

// setDefaultStruct reads config from a struct
func setDefaultStruct(config any) error {
	if !configMustPtrToStruct(config) {
		return ErrConfigNotPtrToStruct
	}

	configJson, err := json.Marshal(config)
	if err != nil {
		return err
	}

	viper.SetConfigType("json")
	err = viper.ReadConfig(bytes.NewBuffer(configJson))

	return err
}

// watch current viper config file, and reload config when changed
// TODO: I don't think this is thread-safe
func watch(config any, hook func(oldConfig any, newConfig any)) error {
	if !configMustPtrToStruct(config) {
		return ErrConfigNotPtrToStruct
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		v := reflect.ValueOf(config).Elem()
		oldConfig := reflect.New(v.Type()).Interface()

		err := deepCopyStruct(config, oldConfig)
		if err != nil {
			logger.WithError(err).
				Errorf("OnConfigChange: deepCopyStruct error")
		}

		_ = viper.Unmarshal(config)
		hook(oldConfig, config)
	})

	viper.WatchConfig()
	return nil
}

func deepCopyStruct(src any, dst any) error {
	if !isPtrToStruct(src) || !isPtrToStruct(dst) {
		return errors.New("copyStruct: src and dst must be ptr to struct")
	}

	srcJson, err := json.Marshal(src)
	if err != nil {
		return err
	}

	err = json.Unmarshal(srcJson, dst)
	if err != nil {
		return err
	}

	return nil
}

func isPtrToStruct(p any) bool {
	tpy := reflect.TypeOf(p)
	ok := tpy.Kind() == reflect.Ptr && tpy.Elem().Kind() == reflect.Struct
	return ok
}

func configMustPtrToStruct(config any) bool {
	if !isPtrToStruct(config) {
		logger.
			WithField("config", fmt.Sprintf("%T", config)).
			Fatalf(ErrConfigNotPtrToStruct.Error())
		return false
	}
	return true
}

var ErrConfigNotPtrToStruct = errors.New("config must be a pointer to struct")

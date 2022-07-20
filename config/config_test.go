package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// `mapstructure:",squash"` is required for embedded structs
type testConfig struct {
	BaseConfig `mapstructure:",squash"`
	Name       string
}

var testConfigInstance = testConfig{
	BaseConfig: BaseConfig{
		DB: DBConfig{
			Driver: "sqlite",
			DSN:    "./test.db",
		},
		HTTP: HTTPConfig{
			Addr: ":8080",
		},
		LogLevel: "info",
	},
	Name: "Test Config",
}

func Test_readFromFile(t *testing.T) {
	var config1 testConfig
	var config2 testConfig

	type args struct {
		path   string
		config any
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{"yaml",
			args{
				"./test_config.yaml",
				&config1,
			},
			&testConfigInstance,
			false,
		},
		{"json",
			args{
				"./test_config.json",
				&config2,
			},
			&testConfigInstance,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := readFromFile(tt.args.path, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("readFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.config, tt.want) {
				t.Errorf("readFromFile() got = %v, want %v", tt.args.config, tt.want)
			}
		})
	}
}

func Test_readFromEnv(t *testing.T) {
	var config1 testConfig
	var config2 testConfig

	type args struct {
		prefix string
		config any
	}
	tests := []struct {
		name    string
		setup   func()
		args    args
		want    any
		wantErr bool
	}{
		{"prefix_CRUD",
			func() {
				_ = os.Setenv("CRUD_DB_DRIVER", "sqlite")
				_ = os.Setenv("CRUD_DB_DSN", "./test.db")
				_ = os.Setenv("CRUD_HTTP_ADDR", ":8080")
				_ = os.Setenv("CRUD_LOGLEVEL", "info")
				_ = os.Setenv("CRUD_NAME", "Test Config")
			},
			args{
				"CRUD",
				&config1,
			},
			&testConfigInstance,
			false,
		},
		{"no_prefix",
			func() {
				_ = os.Setenv("DB_DRIVER", "sqlite")
				_ = os.Setenv("DB_DSN", "./test.db")
				_ = os.Setenv("HTTP_ADDR", ":8080")
				_ = os.Setenv("LOGLEVEL", "info")
				_ = os.Setenv("NAME", "Test Config")
			},
			args{
				"",
				&config2,
			},
			&testConfigInstance,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if err := readFromEnv(tt.args.prefix, tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("readFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			//t.Logf("config: %#v\n", viper.AllSettings())
			if !reflect.DeepEqual(tt.args.config, tt.want) {
				t.Errorf("readFromEnv() got = %v, want %v", tt.args.config, tt.want)
			}
		})
	}
}

func Test_watch(t *testing.T) {
	type args struct {
		config any
		hook   func(oldConfig any, newConfig any)
	}
	tests := []struct {
		name     string
		args     args
		setup    func()
		change   func()
		rollback func()
	}{
		{
			"watch",
			args{
				&testConfigInstance,
				func(oldConfig any, newConfig any) {
					t.Logf("old config: %#v\n", oldConfig)
					t.Logf("new config: %#v\n", newConfig)
					t.Logf("now config: %#v\n", testConfigInstance)
				},
			},
			func() {
				_ = readFromFile("./test_config.yaml", &testConfigInstance)
			},
			func() {
				fmt.Println("change config")
				err := changeLine("./test_config.yaml", 6 /* Name */, "Name: Config Changed")
				if err != nil {
					t.Fatalf("changeLine() error = %v", err)
				}
			},
			func() {
				fmt.Println("rollback config")
				err := changeLine("./test_config.yaml", 6 /* Name */, "Name: Test Config")
				if err != nil {
					t.Fatalf("changeLine() error = %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := watch(tt.args.config, tt.args.hook)
			if err != nil {
				t.Errorf("watch() error = %v", err)
			}
			tt.change()
			time.Sleep(1 * time.Second)
			tt.rollback()
		})
	}
}

func changeLine(path string, line int, newLine string) error {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	lines[line] = newLine

	output := strings.Join(lines, "\n")
	return ioutil.WriteFile(path, []byte(output), 0644)
}

func TestInit(t *testing.T) {
	type MyConfig struct {
		BaseConfig `mapstructure:",squash"`
		Foo        string
		Bar        int
	}
	var config MyConfig

	err := Init(&config,
		FromFile("./test_config.yaml"),
		FromEnv("MYAPP"),
		WatchFileChange(func(oldConfig any, newConfig any) {
			logger.WithField("oldConfig", oldConfig).
				WithField("newConfig", newConfig).
				Info("config changed")
		}),
	)
	if err != nil {
		t.Errorf("Init() error = %v", err)
	}

	t.Run("config", func(t *testing.T) {
		if !reflect.DeepEqual(config, config) {
			t.Errorf("Init() got config = %v, want %v", config, config)
		}
	})
}

func ExampleInit() {
	type MyConfig struct {
		BaseConfig `mapstructure:",squash"`
		Foo        string
		Bar        int
	}
	var config MyConfig

	err := Init(&config,
		FromFile("./test_config.yaml"),
		FromEnv("MYAPP"),
		WatchFileChange(func(oldConfig any, newConfig any) {
			logger.WithField("oldConfig", oldConfig).
				WithField("newConfig", newConfig).
				Info("config changed")
		}),
	)
	if err != nil {
		logger.WithError(err).Fatal("failed to read config.")
	}
}

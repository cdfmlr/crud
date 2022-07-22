// Package config implements a model based config system for your app.
// It's based on Viper. Instead of using Viper's Get and Set, this package
// binds all config to a struct, and you can get values by codes like
//
//     fmt.Println(Config.Foo.Bar)
//
// This structured config makes it easy and straightforward to get config
// values. And it's also avoiding viper's string problems:
//
//   1. Get("typo"): it's annoying to debug typos in strings. Compiler do
//      nothing to make sure you are getting a right value. You can only find
//      those issues at runtime as unexpected empty value errors.
//
//   2. We have no place (inside the project) to see what configs are required
//      and what types they are. I heat writing codes with a sample config
//      file aside.
//
// You can use config.Init to read (and watch) config from file or environment
// variables and bind it to a struct.
//
// This package also provides a BaseConfig struct, which is a base struct for
// common configs can be applied in a crud app. It's a good idea to embed it
// in your own config struct.
//
// Example:
//
//     type MyConfig struct {
//            BaseConfig `mapstructure:",squash"`
//            Foo        string
//            Bar        int
//        }
//        var config MyConfig
//
//        err := Init(&config,
//            FromFile("./test_config.yaml"),  // read config from file
//            FromEnv("MYAPP"),  // read config from env with prefix MYAPP, for example MYAPP_FOO, MYAPP_DB_DSN
//            WatchFileChange(func(oldConfig any, newConfig any) {  // only file changes are watched, env changes will not trigger.
//                logger.WithField("oldConfig", oldConfig).
//                    WithField("newConfig", newConfig).
//                    Info("config changed")
//            }),
//        )
//        if err != nil {
//            logger.WithError(err).Fatal("failed to read config.")
//        }
package config

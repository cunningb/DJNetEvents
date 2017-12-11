package app

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

var Config appConfig

type appConfig struct {
	// the adrdress of RabiitMq server
	Url string `mapstructure:"url"`
}

func (config appConfig) Validate() error {
	if len(config.Url) == 0 {
		return fmt.Errorf("Url cannot be empty")
	}

	return nil
}

func copyResources(src, dest string) error {
	files, err := ioutil.ReadDir(src)

	if err != nil {
		return err
	}

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err = os.Mkdir(dest, os.ModePerm)

		if err != nil {
			return err
		}

		err = os.Chmod(dest, os.ModePerm)

		if err != nil {
			return err
		}
	}

	for _, file := range files {

		srcPath := fmt.Sprint(src, "/", file.Name())
		destPath := fmt.Sprint(dest, "/", file.Name())

		//Walk directory paths
		if file.IsDir() {
			err = copyResources(srcPath, destPath)

			if err != nil {
				return err
			}
		} else {
			// Copy only files that dont exist
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				err = copyFile(srcPath, destPath)
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return err
	}
	defer out.Close()

	os.Chmod(dest, os.ModePerm)

	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)

	if err != nil {
		return err
	}

	return nil
}

// LoadConfig loads configuration from the given list of paths and populates it into the Config variable.
// The configuration file(s) should be named as app.yaml.
// Environment variables with the prefix "SERVER_" in their names are also read automatically.
func LoadConfig(configPath string) error {
	v := viper.New()
	v.SetConfigFile(fmt.Sprint(configPath, "config/config.yml"))
	v.SetConfigType("yaml")
	v.SetEnvPrefix("server")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("url", "")

	err := copyResources(fmt.Sprint(configPath, "resources"), fmt.Sprint(configPath, "config"))

	if err != nil {
		return err
	}

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("Failed to read the configuration file: %s", err)
	}
	if err := v.Unmarshal(&Config); err != nil {
		return err
	}

	return Config.Validate()
}

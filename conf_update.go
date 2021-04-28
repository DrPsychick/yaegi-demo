///usr/bin/env yaegi run "$0" "$@"; exit
package main

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func updateConfigFromEnv(conf []byte, pfx string) ([]byte, error) {
	config, err := toml.Load(string(conf))
	if err != nil {
		return nil, err
	}

	var envs = make(map[string]string)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, pfx) {
			continue
		}
		// PFX_VARIABLE=section.sub.name=value
		parts := strings.Split(e, "=")
		if strings.Contains(parts[1], "|") {
			// backwards compatibility
			// PFX_VARIABLE=[section.sub]|value=["list","elements"]
			parts[1] = strings.ReplaceAll(strings.ReplaceAll(parts[1], "[", ""), "]", "")
			parts[1] = strings.ReplaceAll(parts[1], "|", ".")
		}
		// trim quotes from the value
		envs[parts[1]] = strings.Trim(strings.Join(parts[2:], "="), "\"")
	}
	if len(envs) == 0 {
		return nil, fmt.Errorf("No ENVs with prefix %s", pfx)
	}

	for k, v := range envs {
		var withComment = false
		if strings.HasPrefix(v, "#") {
			withComment = true
			v = strings.TrimLeft(v, "#")
		}

		// set value with correct type
		var val interface{}
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			val = i
		} else if f, err := strconv.ParseFloat(v, 64); err == nil {
			val = f
		} else if b, err := strconv.ParseBool(v); err == nil {
			val = b
		} else if strings.HasPrefix(v, "[") {
			// ["foo", "bar"] - we have to remove the quotes or they will be escaped and included
			list := strings.Split(strings.Trim(v, "[]"), ",")
			for i, v := range list {
				// trim quotes and spaces from each element
				list[i] = strings.Trim(v, "\" ")
			}
			val = list
		} else {
			// default is string with quotes
			val = v
		}

		if withComment {
			config.SetWithComment(k, "", true, val)
		} else {
			config.Set(k, val)
		}
	}

	out, err := toml.Marshal(config)
	return out, err
}

func main() {
	var conf_file = os.Getenv("CONF_UPDATE")
	var prefix = os.Getenv("CONF_PREFIX")
	if conf_file == "" || prefix == "" {
		fmt.Println("No CONF_UPDATE or CONF_PREFIX defined - exiting.")
		os.Exit(0)
	}

	var buf []byte
	var err error
	if buf, err = ioutil.ReadFile(conf_file); err != nil {
		fmt.Printf("Failed to read config file: %s\n", conf_file)
		os.Exit(1)
	}
	if buf, err = updateConfigFromEnv(buf, prefix); err != nil {
		fmt.Printf("Failed to update config from ENV: %s\n", err)
		os.Exit(1)
	}
	if err = ioutil.WriteFile(conf_file, buf, 0644); err != nil {
		fmt.Printf("Failed to write back config to file '%s': %s\n", conf_file, err)
		os.Exit(1)
	}
}

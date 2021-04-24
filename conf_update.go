///usr/bin/env yaegi run "$0" "$@"; exit
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	debug            = false
	removeComments   = false
	removeEmtpyLines = false
	indentation      = "  "
	sectionPrefixes  = []string{"["} //, "[["}
	sectionPostfixes = []string{"]"} //, "]]"}
)

type Key struct {
	Env      string   `json:"env"`
	Name     string   `json:"name"`
	Sections []string `json:"sections"`
	Value    string   `json:"value"`
}

func envValueMatch(prefix string) map[string]Key {
	var envs = make(map[string]Key)

	for _, e := range os.Environ() {
		// PREFIX_ANYTHING=[section|[subsection|[...]]]keyname="value"
		if !strings.HasPrefix(e, prefix) {
			continue
		}

		parts := strings.Split(e, "=")
		if len(parts) < 3 {
			fmt.Printf("Skipping '%s', expecting 'ENV_VAR=[section[|subsection|]]key_name=\"value\"'", e)
			continue
		}

		c := strings.Count(parts[1], "|")
		name := strings.Split(parts[1], "|")[c]
		sections := []string{}
		if c > 0 {
			sections = strings.Split(parts[1], "|")[:c]
		}

		value := strings.Join(parts[2:], "=")

		envs[strings.Join(sections, ".")+"."+name] = Key{e, name, sections, value}

	}

	return envs
}

func updateOrCreate(buf []byte, k Key) ([]byte, bool) {
	match := false
	hasSection := len(k.Sections) > 0
	prefixIndex := -1
	out := make([]byte, len(buf))
	out = []byte{}

	s := bufio.NewScanner(bytes.NewReader(buf))
	inSection := 0
	for s.Scan() {
		ltrimmed := strings.TrimLeft(s.Text(), "\t ")
		// skip comments (allow whitespace at the beginning)
		if strings.HasPrefix(ltrimmed, "#") || strings.HasPrefix(ltrimmed, "//") {
			if debug {
				fmt.Print("c")
			}
			if !removeComments {
				out = append(out, s.Text()+"\n"...)
			}
			continue
		}
		// skip empty lines
		if removeEmtpyLines && strings.Trim(s.Text(), "\t ") == "" {
			if debug {
				print(".")
			}
			continue
		}

		// already matched -> copy rest as-is
		if match {
			if debug {
				fmt.Print("-")
			}
			out = append(out, s.Text()+"\n"...)
			continue
		}

		// determine if this line is a section
		isSection := false
		isTopLevelSection := false
		matchSection := false
		for i, p := range sectionPrefixes {
			if !strings.HasPrefix(ltrimmed, p) {
				continue
			}
			prefixIndex = i
			isSection = true
			// matches top level section (untrimmed match)
			if strings.HasPrefix(s.Text(), p) {
				isTopLevelSection = true
			}
			// matches key in current section level
			if hasSection && inSection < len(k.Sections) && strings.HasPrefix(ltrimmed, p+k.Sections[inSection]+sectionPostfixes[i]) {
				matchSection = true
			}
		}

		// shortcut: key has no section, but we're in a section
		if !hasSection && isSection {
			if debug {
				fmt.Print("_")
			}
			out = append(out, s.Text()+"\n"...)
			continue
		}

		// "descend" sub sections until at deepest level
		if hasSection && matchSection && inSection < len(k.Sections) {
			if debug {
				fmt.Print("s")
			}
			inSection++
			out = append(out, s.Text()+"\n"...)
			continue
		}

		// leaving deepest level: reset section and continue
		if hasSection && isSection && inSection == len(k.Sections) {
			// key not found -> create it
			if !match {
				ind := ""
				for _ = range k.Sections {
					ind += indentation
				}
				if debug {
					fmt.Print("A")
				}
				out = append(out, ind+k.Name+" = "+k.Value+"\n"...)
				match = true
			}
			if debug {
				fmt.Print("R")
			}
			inSection = 0
			out = append(out, s.Text()+"\n"...)
			continue
		}

		// leaving section without match -> add subsection(s) and key
		if hasSection && inSection > 0 && isTopLevelSection && inSection < len(k.Sections) {
			// add missing section(s)
			ind := ""
			for i, s := range k.Sections {
				if i < inSection {
					continue
				}

				ind = ""
				for j := 0; j < i; j++ {
					ind += indentation
				}
				sec := sectionPrefixes[prefixIndex] + s + sectionPostfixes[prefixIndex]
				if debug {
					fmt.Print("C")
				}
				out = append(out, ind+sec+"\n"...)
			}
			// add key
			if debug {
				fmt.Print("A")
			}
			ind += indentation
			out = append(out, ind+k.Name+" = "+k.Value+"\n"...)
			match = true

			// current section -> continue
			if debug {
				fmt.Print("s")
			}
			inSection = 0
			out = append(out, s.Text()+"\n"...)
			continue
		}

		// keys that are not in our section
		if hasSection && !isSection && inSection < len(k.Sections) {
			if debug {
				fmt.Print(".")
			}
			out = append(out, s.Text()+"\n"...)
			continue
		}

		// we're in the right section and need to match the key
		// no match
		if !strings.HasPrefix(ltrimmed, fmt.Sprintf("%s=", k.Name)) && !strings.HasPrefix(ltrimmed, fmt.Sprintf("%s =", k.Name)) {
			if debug {
				fmt.Print(".")
			}
			out = append(out, s.Text()+"\n"...)
			continue
		}

		// match, but no change
		if ltrimmed == k.Name+" = "+k.Value {
			if debug {
				fmt.Print("k")
			}
			out = append(out, s.Text()+"\n"...)
			match = true
			continue
		}

		// replace "    key_name = value"
		if debug {
			fmt.Print("U")
		}
		line := strings.Split(s.Text(), "=")[0] + "= " + k.Value + "\n"
		out = append(out, line...)
		match = true
	}
	// no match -> create new section at the end
	if !match {
		ind := ""
		for i, s := range k.Sections {
			if i < inSection {
				continue
			}
			name := sectionPrefixes[prefixIndex] + s + sectionPostfixes[prefixIndex]
			ind = ""
			for j := 0; j < i; j++ {
				ind += indentation
			}
			// add section(s)
			if debug {
				fmt.Print("C")
			}
			out = append(out, ind+name+"\n"...)
		}
		// add key with last indentation + 1
		if debug {
			fmt.Print("A")
		}
		ind += indentation
		out = append(out, ind+k.Name+" = "+k.Value+"\n"...)
		match = true
	}

	if debug {
		fmt.Println()
	}

	return out, match
}

func main() {
	var conf_file = os.Getenv("CONF_UPDATE")
	var prefix = os.Getenv("CONF_PREFIX")
	if conf_file == "" || prefix == "" {
		if debug {
			fmt.Println("No CONF_UPDATE or CONF_PREFIX defined - exiting.")
		}
		os.Exit(0)
	}
	if os.Getenv("CONF_DEBUG") == "true" {
		debug = true
	}
	if os.Getenv("CONF_STRIP_COMMENTS") == "true" {
		removeComments = true
	}
	if os.Getenv("CONF_STRIP_EMPTYLINES") == "true" {
		removeEmtpyLines = true
	}

	envs := envValueMatch(prefix)
	str, _ := json.MarshalIndent(envs, "", "  ")
	if debug {
		fmt.Println(string(str))
	}

	// read the config file
	buf, _ := ioutil.ReadFile(conf_file)

	var match bool
	for i, k := range envs {
		if debug {
			fmt.Printf("Key: %s\n", i)
		}
		// replace matching value or create the key in config
		if buf, match = updateOrCreate(buf, k); match {
			continue
		}
		if debug {
			fmt.Printf("No match: %s = %s\n", i, k.Value)
		}
	}

	if debug {
		fmt.Printf("%s\n", buf)
	}

	// write the config file
	ioutil.WriteFile(conf_file, buf, 0644)
}

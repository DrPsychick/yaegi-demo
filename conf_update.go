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
	removeComments   = true
	removeEmtpyLines = true
	indentation      = "  "
	sectionPrefixes  = []string{"[", "[["}
	sectionPostfixes = []string{"]", "]]"}
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
	out := make([]byte, len(buf))
	out = []byte{}

	s := bufio.NewScanner(bytes.NewReader(buf))
	inSection := 0
	for s.Scan() {
		ltrimmed := strings.TrimLeft(s.Text(), "\t ")
		// skip comments (allow whitespace at the beginning)
		if strings.HasPrefix(ltrimmed, "#") || strings.HasPrefix(ltrimmed, "//") {
			fmt.Print("c")
			if !removeComments {
				out = append(out, s.Text()+"\n"...)
			}
			continue
		}
		if removeEmtpyLines && strings.Trim(s.Text(), "\t ") == "" {
			print(".")
			continue
		}

		// determine if this line is a section
		isSection := false
		matchSection := false
		for i, p := range sectionPrefixes {
			if !strings.HasPrefix(ltrimmed, p) {
				continue
			}
			isSection = true
			// shortcut: skip if we're at the deepest level
			if inSection == len(k.Sections) {
				break
			}
			// matches key in current section level
			if strings.HasPrefix(ltrimmed, p+k.Sections[inSection]+sectionPostfixes[i]) {
				matchSection = true
				break
			}
		}

		// leaving deepest level: reset section and continue
		if inSection == len(k.Sections) && isSection {
			// key not found -> create it
			if !match {
				ind := ""
				for _ = range k.Sections {
					ind += indentation
				}
				out = append(out, ind+k.Name+" = "+k.Value+"\n"...)
				fmt.Print("A")
			}
			fmt.Print("R")
			inSection = 0
		}
		// top level section
		if inSection == 0 && hasSection && matchSection {
			fmt.Print("S")
			inSection = 1
			out = append(out, s.Text()+"\n"...)
			continue
		}
		// "descend" sub sections until at deepest level
		if hasSection && len(k.Sections) > inSection {
			if matchSection {
				fmt.Print("s")
				inSection++
			} else {
				fmt.Print(".")
			}
			out = append(out, s.Text()+"\n"...)
			continue
		}
		// no match
		if !strings.HasPrefix(ltrimmed, fmt.Sprintf("%s=", k.Name)) && !strings.HasPrefix(ltrimmed, fmt.Sprintf("%s =", k.Name)) {
			fmt.Print(".")
			out = append(out, s.Text()+"\n"...)
			continue
		}

		// match, but no change
		if ltrimmed == k.Name+" = "+k.Value {
			fmt.Print("k")
			out = append(out, s.Text()+"\n"...)
			match = true
			continue
		}

		// replace "    key_name = value"
		fmt.Print("U")
		line := strings.Split(s.Text(), "=")[0] + "= " + k.Value + "\n"
		out = append(out, line...)
		match = true
	}
	// TODO: no match -> create new section at the end

	fmt.Println("DONE")

	return out, match
}

func main() {
	var prefix = "PFX"
	envs := envValueMatch(prefix)
	str, _ := json.MarshalIndent(envs, "", "  ")
	fmt.Println(string(str))

	// read the config file
	buf, _ := ioutil.ReadFile("test.conf")

	var match bool
	for i, k := range envs {
		fmt.Printf("Key: %s\n", i)
		// replace matching value or create the key in config
		if buf, match = updateOrCreate(buf, k); match {
			continue
		}
		fmt.Printf("No match: %s = %s\n", i, k.Value)
	}

	fmt.Printf("%s\n", buf)

	// write the config file
	ioutil.WriteFile("test.conf", buf, 0644)
}

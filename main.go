package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"
	"github.com/ladydascalie/litmus/format"
	"github.com/ladydascalie/litmus/internal/extract"
	"github.com/ladydascalie/litmus/p"
	"github.com/ladydascalie/litmus/pkg"
	e "github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	environment = map[string]interface{}{}
	client      = &http.Client{
		Timeout: time.Second,
	}
)

func init() {
	log.SetFlags(0)
}

func main() {
	// prepare the environment
	var configPath string
	var testByName string
	var eVariables pkg.KeyValuePairs

	rootCmd := cobra.Command{
		Use:   "litmus",
		Short: "Run automated HTTP requests.",
		Long:  litmusBanner + longHelp,
		Run: func(cmd *cobra.Command, args []string) {
			if err := setEnvironmentFile(configPath); err != nil {
				log.Fatal(err)
			}

			// Set environment from user args, taking precedence
			// over the environment config in env.yaml.
			for _, kvp := range eVariables {
				environment[kvp.Key] = kvp.Value
			}

			if err := runRequests(configPath, testByName); err != nil {
				fmt.Printf("\t[%s] %v\n", p.Red("FAIL"), err)
			}
		},
	}
	// see usages.go for all usages
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", cFlagUsage)
	rootCmd.Flags().StringVarP(&testByName, "test", "n", "", nFlagUsage)
	rootCmd.Flags().VarP(&eVariables, "env", "e", eFlagUsage)

	// enforce the required flags
	rootCmd.MarkFlagRequired("config")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runRequests(config string, name string) (err error) {
	litmusFiles, err := loadRequests(config)
	if err != nil {
		return err
	}

	for _, file := range litmusFiles {
		for _, test := range file.Litmus.Test {
			if name != "" && test.Name != name {
				continue
			}

			if err = runRequest(test); err != nil {
				return
			}
		}
	}

	return
}

func loadRequests(config string) (tests []format.LitmusFile, err error) {
	const testFileGlob = "*_test.toml"

	config = strings.TrimSuffix(config, "/") + "/"
	files, err := filepath.Glob(config + testFileGlob)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		log.Fatalf("no test files found in %s folder", config)
	}

	for _, file := range files {
		var lit format.LitmusFile
		if err = unmarhsal(file, &lit); err != nil {
			return
		}

		tests = append(tests, lit)
	}
	return
}

func runRequest(r format.RequestTest) (err error) {
	if err = applyEnvironments(&r); err != nil {
		return e.Wrap(err, "applying environment")
	}

	fmt.Printf("[%s] %s - %s\n", p.Blue("TEST"), r.Name, r.URL)

	request, err := http.NewRequest(r.Method, r.URL, strings.NewReader(r.Body))
	if err != nil {
		return e.Wrap(err, "creating request")
	}

	for k, v := range r.Headers {
		request.Header.Set(k, v)
	}

	resp, err := client.Do(request)
	if err != nil {
		return e.Wrap(err, "performing request")
	}
	defer resp.Body.Close()

	// Get, set and assert stuff from the response body.
	if err = processResponse(r, resp); err != nil {
		return e.Wrap(err, "extracting body")
	}

	fmt.Printf("\t[%s]\n", p.Green("PASS"))
	return
}

func processResponse(r format.RequestTest, resp *http.Response) error {
	if err := extract.StatusCode(r, resp, environment); err != nil {
		return err
	}
	if err := extract.Header(r, resp, environment); err != nil {
		return err
	}
	if err := extract.Body(r, resp, environment); err != nil {
		return err
	}

	return nil
}

func applyEnvironments(r *format.RequestTest) (err error) {
	r.URL, err = applyEnvironment(r.URL)
	if err != nil {
		return
	}

	r.Body, err = applyEnvironment(r.Body)
	if err != nil {
		return
	}

	for k, v := range r.Headers {
		r.Headers[k], err = applyEnvironment(v)
		if err != nil {
			return
		}
	}

	for i := range r.Getters {
		if r.Getters[i].Expected, err = applyEnvironment(r.Getters[i].Expected); err != nil {
			return
		}
	}

	return
}

func applyEnvironment(input string) (output string, err error) {
	buf := &bytes.Buffer{}
	t, err := template.New("anon").Parse(input)
	if err != nil {
		return "", err
	}
	if err = t.Execute(buf, environment); err != nil {
		return
	}

	return buf.String(), nil
}

func setEnvironmentFile(config string) (error) {
	fullPath := filepath.Join(config, "env.toml")
	spew.Dump(fullPath)

	// If the file doesn't exist, nil out the error because
	// this isn't a show-stopper.
	if _, err := os.Stat(fullPath); err != nil {
		log.Printf("env file %q does not exist", fullPath)
		return nil
	}
	if err := unmarhsal(fullPath, &environment); err != nil {
		return err
	}
	return nil
}

func unmarhsal(fullPath string, target interface{}) (err error) {
	file, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return e.Wrap(err, "reading file")
	}
	if err = toml.Unmarshal(file, target); err != nil {
		return e.Wrap(err, "unmarshalling")
	}

	return
}

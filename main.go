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

	"github.com/fatih/color"
	"github.com/ladydascalie/litmus/internal"
	"github.com/ladydascalie/litmus/pkg"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"gopkg.in/yaml.v2"
)

var (
	environment = map[string]string{}
	client      = &http.Client{
		Timeout: time.Second,
	}

	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
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
			setEnvironmentFile(configPath)

			// Set environment from user args, taking precedence
			// over the environment config in env.yaml.
			for _, kvp := range eVariables {
				environment[kvp.Key] = kvp.Value
			}

			if err := runRequests(configPath, testByName); err != nil {
				fmt.Printf("\t[%s] %v\n", red("FAIL"), err)
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
	requests, err := loadRequests(config)
	if err != nil {
		return err
	}

	for _, request := range requests {
		if name != "" && request.Name != name {
			continue
		}

		if err = runRequest(request); err != nil {
			return
		}
	}

	return
}

func loadRequests(config string) (requests []pkg.RequestConfig, err error) {
	config = strings.TrimSuffix(config, "/") + "/"
	files, err := filepath.Glob(config + "*_test.yaml")
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		log.Fatalf("no test files found in %s folder", config)
	}

	for _, file := range files {
		var r []pkg.RequestConfig
		if err = unmarhsal(file, &r); err != nil {
			return
		}

		requests = append(requests, r...)
	}

	return
}

func runRequest(r pkg.RequestConfig) (err error) {
	if err = applyEnvironments(&r); err != nil {
		return errors.Wrap(err, "applying environment")
	}

	fmt.Printf("[%s] %s - %s\n", blue("TEST"), r.Name, r.URL)

	request, err := http.NewRequest(r.Method, r.URL, strings.NewReader(r.Body))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	for k, v := range r.Headers {
		request.Header.Set(k, v)
	}

	resp, err := client.Do(request)
	if err != nil {
		return errors.Wrap(err, "performing request")
	}
	defer resp.Body.Close()

	// Get, set and assert stuff from the response body.
	if err = processBody(r, resp); err != nil {
		return errors.Wrap(err, "extracting body")
	}

	if r.WantsCode != 0 && r.WantsCode != resp.StatusCode {
		return errors.Errorf("expected response code: %s, but got: %s",
			internal.HttpStatusFmt(r.WantsCode),
			internal.HttpStatusFmt(resp.StatusCode),
		)
	}

	fmt.Printf("\t[%s]\n", green("PASS"))
	return
}

func processBody(r pkg.RequestConfig, resp *http.Response) error {
	if err := extractBody(r, resp); err != nil {
		return err
	}

	if err := extractHeader(r, resp); err != nil {
		return err
	}

	return nil
}

func extractBody(r pkg.RequestConfig, resp *http.Response) (err error) {
	getters := r.Getters.Filter("body")
	if len(getters) == 0 {
		return
	}

	// If we're unable to ascertain the body type, we won't
	// be able to extract anything and needn't bother reading
	// the response body.
	bodyGetter, err := pkg.NewBodyGetter(resp)
	if err != nil {
		return errors.Wrap(err, "creating body getter")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body")
	}
	defer resp.Body.Close()

	for _, getter := range getters {
		act, err := bodyGetter.Get(getter, respBody)
		if err != nil {
			return err
		}

		if getter.Expected != "" {
			if err = equals(getter.Expected, act); err != nil {
				return errors.Wrap(err, "assertion failed")
			}
		}

		if getter.Set != "" {
			environment[getter.Set] = act
			fmt.Printf("\t[%s]  %s -> %s\n", yellow("SET"), act, getter.Set)
		}
	}

	return
}

func extractHeader(r pkg.RequestConfig, resp *http.Response) (err error) {
	getters := r.Getters.Filter("head")
	if len(getters) == 0 {
		return
	}

	headerGetter := &pkg.HeaderGetter{}

	for _, getter := range getters {
		act, err := headerGetter.Get(getter, resp.Header)
		if err != nil {
			return err
		}

		if getter.Expected != "" {
			if err = equals(getter.Expected, act); err != nil {
				return errors.Wrap(err, "assertion failed")
			}
		}

		if getter.Set != "" {
			environment[getter.Set] = act
			fmt.Printf("\t[%s]  %s -> %s\n", yellow("SET"), act, getter.Set)
		}
	}

	return
}

func applyEnvironments(r *pkg.RequestConfig) (err error) {
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

func setEnvironmentFile(config string) (err error) {
	fullPath := filepath.Join(config, "env.yaml")

	// If the file doesn't exist, nil out the error because
	// this isn't a show-stopper.
	if _, err = os.Stat(fullPath); err != nil {
		log.Printf("env file %q does not exist", fullPath)
		return nil
	}

	var envConfigs []pkg.EnvironmentConfig
	if err = unmarhsal(fullPath, &envConfigs); err != nil {
		return
	}

	for _, envConfig := range envConfigs {
		environment[envConfig.Key] = envConfig.Value
	}

	return
}

func unmarhsal(fullPath string, target interface{}) (err error) {
	file, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}

	if err = yaml.Unmarshal(file, target); err != nil {
		return errors.Wrap(err, "unmarshalling")
	}

	return
}

func equals(exp string, act string) (err error) {
	if exp != act {
		return errors.Errorf("\n\texp: %v\n\tgot: %v", exp, act)
	}
	return
}

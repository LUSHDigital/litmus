package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/LUSHDigital/litmus/format"
	"github.com/LUSHDigital/litmus/p"
	"github.com/LUSHDigital/litmus/pkg"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/LUSHDigital/litmus/domain/extract"
)

var (
	client = &http.Client{
		Timeout: 5 * time.Second,
	}
)

func init() {
	log.SetFlags(0)
}

func main() {
	// prepare the environment
	var timeoutLen int
	var configPath string
	var testByName string
	var targetEnv string
	var eVariables pkg.KeyValuePairs

	rootCmd := cobra.Command{
		Use:   "litmus",
		Short: "Run automated HTTP requests.",
		Long:  litmusBanner + longHelp,
		Run: func(cmd *cobra.Command, args []string) {
			// pick the env.toml and unmarshal it into a map
			env, err := setEnvironmentFile(configPath, targetEnv)
			if err != nil {
				log.Fatal(err)
			}

			// Set environment from user args, taking precedence
			// over the environment config in env.toml.
			for _, kvp := range eVariables {
				env[kvp.Key] = kvp.Value
			}

			// Ensure timeout is checked, if provided by the user
			if timeoutLen != 0 {
				client.Timeout = time.Duration(timeoutLen) * time.Second
			}

			if err := runRequests(configPath, testByName, env); err != nil {
				fmt.Printf("\t[%s] %v\n", p.Red("FAIL"), err)
			}
		},
	}
	// see usages.go for all usages
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", cFlagUsage)
	rootCmd.Flags().StringVarP(&testByName, "test", "n", "", nFlagUsage)
	rootCmd.Flags().IntVarP(&timeoutLen, "timeout", "t", 0, tFlagUsage)
	rootCmd.Flags().StringVarP(&targetEnv, "using", "u", "", uFlagUsage)
	rootCmd.Flags().VarP(&eVariables, "env", "e", eFlagUsage)

	// enforce the required flags
	rootCmd.MarkFlagRequired("config")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runRequests(config string, name string, env map[string]interface{}) (err error) {
	litmusFiles, err := loadRequests(config)
	if err != nil {
		return err
	}

	for _, file := range litmusFiles {
		for _, test := range file.Litmus.Test {
			if name != "" && test.Name != name {
				continue
			}

			if err = runRequest(&test, env); err != nil {
				return
			}
		}
	}

	return
}

func loadRequests(config string) (tests []format.TestFile, err error) {
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
		var lit format.TestFile
		if err = unmarhsal(file, &lit); err != nil {
			return
		}

		tests = append(tests, lit)
	}
	return
}

func runRequest(r *format.RequestTest, env map[string]interface{}) (err error) {
	if err := r.ApplyEnv(env); err != nil {
		return errors.Wrap(err, "applying environment")
	}

	fmt.Printf("[%s] %s - %s\n", p.Blue("TEST"), r.Name, r.URL)

	request, err := http.NewRequest(r.Method, r.URL, strings.NewReader(r.Body))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	for k, v := range r.Headers {
		request.Header.Set(k, v)
	}

	q := request.URL.Query()
	for k, v := range r.Query {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()

	resp, err := client.Do(request)
	if err != nil {
		return errors.Wrap(err, "performing request")
	}
	defer resp.Body.Close()

	// Get, set and assert stuff from the response body.
	if err = extract.ProcessResponse(r, resp, env); err != nil {
		return errors.Wrap(err, "extracting body")
	}

	fmt.Printf("\t[%s]\n", p.Green("PASS"))
	return
}

func setEnvironmentFile(config string, targetEnv string) (env map[string]interface{}, err error) {
	const envFile = "env.toml"
	var fullPath string

	// default path
	fullPath = filepath.Join(config, envFile)

	if targetEnv != "" {
		// if not using default, warn
		fmt.Println(p.Green("Running tests using: ", filepath.Base(targetEnv)))
		fullPath = targetEnv
	}

	// If the file doesn't exist, nil out the error because
	// this isn't a show-stopper.
	if _, err := os.Stat(fullPath); err != nil {
		log.Printf("env file %q does not exist", fullPath)
		return nil, err
	}

	if err := unmarhsal(fullPath, &env); err != nil {
		return nil, err
	}
	return env, nil
}

func unmarhsal(fullPath string, target interface{}) (err error) {
	file, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}
	if err = toml.Unmarshal(file, target); err != nil {
		return errors.Wrap(err, "unmarshalling")
	}

	return
}

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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/fatih/color"
	"github.com/LUSHDigital/litmus/domain"
)

var (
	red   = color.New(color.FgHiRed).SprintFunc()
	green = color.New(color.FgHiGreen).SprintFunc()
	blue  = color.New(color.FgHiBlue).SprintFunc()
)

type runner struct {
	client *http.Client
	env    map[string]interface{}
}

func main() {
	// kill prefixes
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// prepare the environment
	var timeoutLen int
	var configPath string
	var testByName string
	var targetEnv string
	var eVariables domain.KeyValuePairs

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
			client := &http.Client{Timeout: 5 * time.Second}
			if timeoutLen != 0 {
				client.Timeout = time.Duration(timeoutLen) * time.Second
			}

			runner := runner{
				client: client,
				env:    env,
			}

			if err := runner.runRequests(configPath, testByName); err != nil {
				fmt.Printf("\t[%s] %v\n", red("FAIL"), err)
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

func (r *runner) runRequests(config string, name string) (err error) {
	litmusFiles, err := loadRequests(config)
	if err != nil {
		return err
	}

	for _, file := range litmusFiles {
		for _, test := range file.Litmus.Test {
			if name != "" && test.Name != name {
				continue
			}

			if err = r.runRequest(&test); err != nil {
				return
			}
		}
	}

	return
}

func loadRequests(config string) (tests []domain.TestFile, err error) {
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
		var lit domain.TestFile
		if err = unmarhsal(file, &lit); err != nil {
			return
		}

		tests = append(tests, lit)
	}
	return
}

func (r *runner) runRequest(req *domain.RequestTest) (err error) {
	if err := req.ApplyEnv(r.env); err != nil {
		return errors.Wrap(err, "applying environment")
	}

	fmt.Printf("[%s] %s - %s\n", blue("TEST"), req.Name, req.URL)

	request, err := http.NewRequest(req.Method, req.URL, strings.NewReader(req.Payload))
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	for k, v := range req.Headers {
		request.Header.Set(k, v)
	}

	q := request.URL.Query()
	for k, v := range req.Query {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()

	resp, err := r.client.Do(request)
	if err != nil {
		return errors.Wrap(err, "performing request")
	}
	defer resp.Body.Close()

	// Get, set and assert stuff from the response body.
	if err = domain.ProcessResponse(req, resp, r.env); err != nil {
		return errors.Wrap(err, "extracting body")
	}

	fmt.Printf("\t[%s]\n", green("PASS"))
	return
}

func setEnvironmentFile(config string, targetEnv string) (env map[string]interface{}, err error) {
	const envFile = "env.toml"
	var fullPath string

	// default path
	fullPath = filepath.Join(config, envFile)

	if targetEnv != "" {
		// if not using default, warn
		fmt.Println(green("Running tests using: ", filepath.Base(targetEnv)))
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

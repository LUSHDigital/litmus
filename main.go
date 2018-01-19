package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/codingconcepts/requestrunner/pkg"
	"github.com/fatih/color"
	"github.com/pkg/errors"

	"gopkg.in/yaml.v2"
)

var (
	environment map[string]string
	client      *http.Client

	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
)

type kvp struct {
	key   string
	value string
}

type kvps []kvp

func (s *kvps) String() string {
	return "my string representation"
}

func (i *kvps) Set(value string) (err error) {
	parts := strings.Split(value, "=")

	*i = append(*i, kvp{
		key:   parts[0],
		value: parts[1],
	})
	return
}

func main() {
	config := flag.String("c", "", "config path")
	name := flag.String("n", "", "name of specific test to run")

	var envs kvps
	flag.Var(&envs, "e", "environment variable")

	flag.Parse()

	if config == nil || *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Apply user environment configuration.
	environment = map[string]string{}
	for _, kvp := range envs {
		environment[kvp.key] = kvp.value
	}

	client = &http.Client{
		Timeout: time.Second,
	}

	if err := runRequests(*config, *name); err != nil {
		fmt.Printf("\t[%s]: %v\n", red("FAIL"), err)
	}
}

func runRequests(config string, name string) (err error) {
	var requests []pkg.RequestConfig

	file, err := ioutil.ReadFile(config)
	if err != nil {
		return errors.Wrap(err, "reading config file")
	}

	if err = yaml.Unmarshal(file, &requests); err != nil {
		return errors.Wrap(err, "unmarshalling requests")
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

func runRequest(r pkg.RequestConfig) (err error) {
	if err = applyEnvironments(&r); err != nil {
		return errors.Wrap(err, "applying environment")
	}

	fmt.Printf("[%s]: %s - %s\n", blue("TEST"), r.Name, r.URL)

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

	fmt.Printf("\t[%s]\n", green("PASS"))
	return
}

func processBody(r pkg.RequestConfig, resp *http.Response) (err error) {
	// If there's nothing in the body to extract, return early.
	bodyGetters := r.Getters.Filter("body")
	if len(bodyGetters) == 0 {
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

	for _, getter := range bodyGetters {
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
			fmt.Printf("\t[%s]:  %s -> %s\n", yellow("SET"), act, getter.Set)
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

func equals(exp string, act string) (err error) {
	if exp != act {
		return errors.Errorf("\n\texp: %v\n\tgot: %v", exp, act)
	}
	return
}

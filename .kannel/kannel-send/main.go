package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/plush/v4"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
)

func main() {
	if len(os.Args) != 2 {
		exit("Missing configuration file parameter")
	}

	//

	template, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		exit("Could not read config: %v", err)
	}

	ctx := plush.NewContext()
	ctx.Set("repeat", func(s string, count int) string {
		return strings.Repeat(s, count)
	})

	config, err := plush.Render(string(template), ctx)
	if err != nil {
		exit("Could not render config: %v", err)
	}

	//

	konf := koanf.New(".")
	if err := konf.Load(rawbytes.Provider([]byte(config)), yaml.Parser()); err != nil {
		exit("Could not load config: %v", err)
	}

	//

	url := craftURL(konf)

	//

	http.DefaultClient.Timeout = 500 * time.Millisecond

	r, err := http.Get(url)
	if err != nil {
		exit("Clould not perform request: %v", err)
	}
	defer r.Body.Close()

	if r.StatusCode >= 400 {
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			exit("Clould not read response body: %v", err)
		}

		exit("Invalid request: %d %d\n%s", r.StatusCode, r.Status, payload)
	}

	fmt.Println("OK")
}

func craftURL(konf *koanf.Koanf) string {
	params := url.Values{}

	for _, parameter := range konf.MapKeys("params") {
		if parameter != "dlr-url" {
			params.Set(parameter, konf.String("params."+parameter))
			continue
		}

		dlr := []string{}
		for _, parameter := range konf.MapKeys("params.dlr-url.params") {
			v := fmt.Sprintf("%s=%s", parameter, konf.String("params.dlr-url.params."+parameter))
			dlr = append(dlr, v)
		}

		v := fmt.Sprintf("%s?%s", konf.String("params.dlr-url.url"), strings.Join(dlr, "&"))
		params.Set(parameter, v)

		if konf.Bool("debug") {
			fmt.Println("dlr-url:", v)
		}
	}

	v := fmt.Sprintf("%s?%s", konf.String("url"), params.Encode())

	if konf.Bool("debug") {
		fmt.Println("    url:", v)
		fmt.Println("")
	}
	return v
}

func exit(message string, args ...interface{}) {
	if len(args) != 0 {
		fmt.Printf(message+"\n", args)
	} else {
		fmt.Println(message)
	}
	os.Exit(1)
}

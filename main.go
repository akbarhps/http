package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		panic("Can't run program, no config file provided")
	}

	run(parseConfig(os.Args[1]))
}

func run(configs []RequestConfig) {
	responses := make(map[int]interface{})

	fmt.Print("[")
	for i, config := range configs {
		ctype := ""
		buffer := bytes.NewBuffer(nil)

		if len(config.Form) > 0 {
			buffer, ctype = parseFormBody(&config, responses)
		}
		if len(config.JSON) > 0 {
			buffer, ctype = parseJSONBody(&config, responses)
		}

		fullUrl := parseUrl(&config, responses)
		request, err := http.NewRequest(config.Method, fullUrl, buffer)
		if err != nil {
			panic(err)
		}

		if len(ctype) > 0 {
			request.Header.Set("Content-Type", ctype)
		}
		for k, v := range config.Headers {
			request.Header.Set(k, replaceReferenceWithValue(v, responses))
		}

		timestamp := time.Now()
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			panic(err)
		}

		body := response.Body
		responseBody, err := io.ReadAll(body)
		if err != nil {
			panic(err)
		}

		err = body.Close()
		if err != nil {
			panic(err)
		}

		var responseJSON interface{}
		err = json.Unmarshal(responseBody, &responseJSON)
		if err != nil {
			panic(err)
		}

		pretty, err := json.MarshalIndent(map[string]interface{}{
			"_____url":    fullUrl,
			"____status":  response.Status,
			"___duration": time.Since(timestamp).String(),
			"__headers":   response.Header,
			"_body":       responseJSON,
		}, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s", pretty)

		// break early if status code is not 200
		if response.StatusCode != 200 {
			break
		}

		if i+1 < len(configs) {
			fmt.Println(",\n\n")
		}

		responses[i] = responseJSON
	}
	fmt.Print("]")
}

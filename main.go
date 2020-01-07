package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/fatih/color"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hokaccha/go-prettyjson"
	"github.com/itchyny/gojq"
	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Action: getjson,
		Name:   "getjson",
		Usage:  "get JSON from Packet API",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path, p", Required: true},
			&cli.StringFlag{Name: "method, m", Value: "GET"},
			&cli.StringFlag{Name: "query, q"},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getjson(c *cli.Context) error {
	p := c.String("path")
	m := c.String("method")
	q := c.String("query")
	urlp, err := url.Parse("https://api.packet.net/" + p)
	if err != nil {
		return err
	}
	req, err := retryablehttp.NewRequest(m, urlp.String(), nil)
	if err != nil {
		return err
	}
	ak := os.Getenv("PACKET_AUTH_TOKEN")
	if ak == "" {
		return fmt.Errorf("set PACKET_AUTH_TOKEN")
	}
	mediaType := "application/json"
	req.Header.Add("X-Auth-Token", ak)
	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = nil

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if q != "" {

		var buf bytes.Buffer
		dec := json.NewDecoder(io.TeeReader(resp.Body, &buf))
		dec.UseNumber()

		code, err := gojq.Parse(q)
		if err != nil {
			return err
		}
		for {
			var v interface{}
			if err := dec.Decode(&v); err != nil {
				if err == io.EOF {
					break
				}
				log.Println(err, "invalid JSON:", buf.String())
				return fmt.Errorf("invalid json: %s\n", err)
			}
			if err := printValue(code.Run(v)); err != nil {
				return fmt.Errorf("whiile running: %s", err)
			}
		}
	} else {
		dst := &bytes.Buffer{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Indent(dst, body, "", "  "); err != nil {
			return err
		}
		fmt.Println(dst.String())
	}

	return nil
}

func printValue(v gojq.Iter) error {
	m := jsonFormatter()
	for {
		x, ok := v.Next()
		if !ok {
			break
		}
		switch v := x.(type) {
		case error:
			return v
		case [2]interface{}:
			if s, ok := v[0].(string); ok {
				if s == "HALT:" {
					return nil
				}
				if s == "STDERR:" {
					x = v[1]
				}
			}
		}
		xs, err := m.Marshal(x)
		if err != nil {
			return err
		}
		fmt.Println(string(xs))
	}
	return nil
}

func jsonFormatter() *prettyjson.Formatter {
	f := prettyjson.NewFormatter()
	f.StringColor = color.New(color.FgGreen)
	f.BoolColor = color.New(color.FgYellow)
	f.NumberColor = color.New(color.FgCyan)
	f.NullColor = color.New(color.FgHiBlack)
	return f
}

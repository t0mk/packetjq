package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hokaccha/go-prettyjson"
	"github.com/itchyny/gojq"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Action: getAndParse,
		Name:   "getandparse",
		Usage:  "get JSON from Packet API",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "path", Aliases: []string{"p"}},
			&cli.StringFlag{Name: "method", Aliases: []string{"m"}, Value: "GET"},
			&cli.StringFlag{Name: "query", Aliases: []string{"q"}},
			&cli.BoolFlag{Name: "debug", Aliases: []string{"d"}},
			&cli.StringFlag{Name: "requestbody", Aliases: []string{"r"}},
			&cli.PathFlag{Name: "dummyjsonresponse", Aliases: []string{"j"}},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func dumpRequest(req *retryablehttp.Request) {
	o, _ := httputil.DumpRequestOut(req.Request, false)
	strReq := string(o)
	reg, _ := regexp.Compile(`X-Auth-Token: (\w*)`)
	reMatches := reg.FindStringSubmatch(strReq)
	if len(reMatches) == 2 {
		strReq = strings.Replace(strReq, reMatches[1], strings.Repeat("-", len(reMatches[1])), 1)
	}
	bbs, _ := req.BodyBytes()
	log.Printf("\n=======[REQUEST]=============\n%s%s\n", strReq, string(bbs))
}

func getJsonFromAPI(c *cli.Context) ([]byte, error) {
	p := c.String("path")
	m := c.String("method")
	r := c.String("requestbody")
	d := c.Bool("debug")
	urlp, err := url.Parse("https://api.equinix.com/metal/v1/" + p)
	if err != nil {
		return nil, err
	}

	var reqbody []byte
	if r == "" {
		reqbody = nil
	} else {
		reqbody = []byte(r)
	}

	req, err := retryablehttp.NewRequest(m, urlp.String(), reqbody)
	if err != nil {
		return nil, err
	}
	ak := os.Getenv("PACKET_AUTH_TOKEN")
	if ak == "" {
		return nil, fmt.Errorf("set PACKET_AUTH_TOKEN")
	}
	mediaType := "application/json"
	req.Header.Add("X-Auth-Token", ak)
	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = nil
	if d {
		dumpRequest(req)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if d {
		o, _ := httputil.DumpResponse(resp, true)
		log.Printf("\n=======[RESPONSE]============%s\n\n", string(o))
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func checkFlags(c *cli.Context) {
	j := c.Path("dummyjsonresponse")
	p := c.String("path")
	r := c.String("requestbody")
	if len(j) > 0 {
		if len(p) > 0 {
			panic(fmt.Errorf("No reason to set path if you supply local json"))
		}
		if len(r) > 0 {
			panic(fmt.Errorf("No reason to set request body if you supply local json"))
		}
	}

}

func getAndParse(c *cli.Context) error {
	var jsbytes []byte
	var err error
	q := c.String("query")
	j := c.Path("dummyjsonresponse")

	checkFlags(c)

	if len(j) > 0 {
		jsbytes, err = ioutil.ReadFile(j)
		if err != nil {
			return err
		}
	} else {
		jsbytes, err = getJsonFromAPI(c)
	}

	if q != "" {

		//var buf bytes.Buffer
		dec := json.NewDecoder(bytes.NewReader(jsbytes))
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
				log.Println(err, "invalid JSON:", string(jsbytes))
				return fmt.Errorf("invalid json: %s\n", err)
			}
			if err := printValue(code.Run(v)); err != nil {
				return fmt.Errorf("while running: %s", err)
			}
		}
	} else {
		dst := &bytes.Buffer{}
		if len(jsbytes) == 0 {
			fmt.Println("Empty response body")
			return nil
		}
		if err := json.Indent(dst, jsbytes, "", "  "); err != nil {
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

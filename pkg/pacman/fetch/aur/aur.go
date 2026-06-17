package aur

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Riyyi/declpac/pkg/log"
)

var AURInfoURL = "https://aur.archlinux.org/rpc?v=5&type=info"

type Package struct {
	Name        string   `json:"Name"`
	PackageBase string   `json:"PackageBase"`
	Version     string   `json:"Version"`
	URL         string   `json:"URL"`
	Depends     []string `json:"Depends"`
	MakeDepends []string `json:"MakeDepends"`
}

func (p Package) AllDepends() []string {
	return append(p.Depends, p.MakeDepends...)
}

type Response struct {
	Results []Package `json:"results"`
}

type Client struct {
	cache map[string]Package
}

func New() *Client {
	return &Client{
		cache: make(map[string]Package),
	}
}

func (c *Client) Fetch(packages []string) (map[string]Package, error) {
	start := time.Now()
	log.Debug("aur.Fetch: starting...")

	result := make(map[string]Package)

	if len(packages) == 0 {
		return result, nil
	}

	var uncached []string
	for _, pkg := range packages {
		if _, ok := c.cache[pkg]; !ok {
			uncached = append(uncached, pkg)
		}
	}

	if len(uncached) == 0 {
		log.Debug("aur.Fetch: done (cached) (%.2fs)", time.Since(start).Seconds())
		for _, pkg := range packages {
			result[pkg] = c.cache[pkg]
		}
		return result, nil
	}

	v := url.Values{}
	for _, pkg := range packages {
		v.Add("arg[]", pkg)
	}

	resp, err := http.Get(AURInfoURL + "&" + v.Encode())
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	var aurResp Response
	if err := json.Unmarshal(body, &aurResp); err != nil {
		return result, err
	}

	for _, r := range aurResp.Results {
		c.cache[r.Name] = r
		result[r.Name] = r
	}

	log.Debug("aur.Fetch: done (%.2fs)", time.Since(start).Seconds())
	return result, nil
}

func (c *Client) Get(name string) (Package, bool) {
	pkg, ok := c.cache[name]
	return pkg, ok
}

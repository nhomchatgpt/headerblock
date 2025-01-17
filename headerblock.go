// Package headerblock is a plugin to block headers which regex matched by their name and/or value
package headerblock

import (
	"context"
	"log"
	"net/http"
	"regexp"
)

// Config the plugin configuration.
type Config struct {
	RequestHeaders          []HeaderConfig `json:"requestHeaders,omitempty"`
	WhitelistRequestHeaders []HeaderConfig `json:"whitelistRequestHeaders,omitempty"`
	Log                     bool           `json:"log,omitempty"`
}

// HeaderConfig is part of the plugin configuration.
type HeaderConfig struct {
	Name  string `json:"header,omitempty"`
	Value string `json:"env,omitempty"`
}

type rule struct {
	name  *regexp.Regexp
	value *regexp.Regexp
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		Log: false,
	}
}

// headerBlock a Traefik plugin.
type headerBlock struct {
	next                  http.Handler
	requestHeaderRules    []rule
	whitelistRequestRules []rule
	log                   bool
}

// New creates a new headerBlock plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &headerBlock{
		next:                  next,
		requestHeaderRules:    prepareRules(config.RequestHeaders),
		whitelistRequestRules: prepareRules(config.WhitelistRequestHeaders),
		log:                   config.Log,
	}, nil
}

func prepareRules(headerConfig []HeaderConfig) []rule {
	headerRules := make([]rule, 0)
	for _, requestHeader := range headerConfig {
		requestRule := rule{}
		if len(requestHeader.Name) > 0 {
			requestRule.name = regexp.MustCompile(requestHeader.Name)
		}
		if len(requestHeader.Value) > 0 {
			requestRule.value = regexp.MustCompile(requestHeader.Value)
		}
		headerRules = append(headerRules, requestRule)
	}
	return headerRules
}

func (c *headerBlock) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Check whitelist rules only if they are defined
	if len(c.whitelistRequestRules) > 0 {
		for name, values := range req.Header {
			for _, rule := range c.whitelistRequestRules {
				if applyRule(rule, name, values) {
					if c.log {
						log.Printf("%s: access allowed - whitelisted header: %s", req.URL.String(), name)
					}
					c.next.ServeHTTP(rw, req)
					return
				}
			}
		}

		// If no whitelist rules match, block the request
		if c.log {
			log.Printf("%s: access denied - no matching whitelist headers", req.URL.String())
		}
		rw.WriteHeader(http.StatusForbidden)
		return
	}

	// Apply blocklist rules
	for name, values := range req.Header {
		for _, rule := range c.requestHeaderRules {
			if applyRule(rule, name, values) {
				// Block the request if a blocking rule matches
				if c.log {
					log.Printf("%s: access denied - blocked header: %s", req.URL.String(), name)
				}
				rw.WriteHeader(http.StatusForbidden)
				return
			}
		}
	}

	// Allow the request if no rules match
	if c.log {
		log.Printf("%s: access allowed - no rules matched", req.URL.String())
	}
	c.next.ServeHTTP(rw, req)
}

func applyRule(rule rule, name string, values []string) bool {
	nameMatch := rule.name != nil && rule.name.MatchString(name)
	if rule.value == nil && nameMatch {
		return true
	} else if rule.value != nil && (nameMatch || rule.name == nil) {
		for _, value := range values {
			if rule.value.MatchString(value) {
				return true
			}
		}
	}
	return false
}

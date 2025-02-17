package main

import (
	"context"
	"encoding/json"
	"log"

	jsonengine "github.com/kyverno/kyverno-json/pkg/json-engine"
	"github.com/kyverno/kyverno-json/pkg/policy"
	"go.uber.org/multierr"
)

const policyYAML = `
apiVersion: json.kyverno.io/v1alpha1
kind: ValidatingPolicy
metadata:
  name: authz
spec:
  rules:
  - name: delete-checks
    identifier: "name"
    match:
      all:
        (input.method): "DELETE"
    assert:
      all:
      - check:
          role: "admin"
`

func main() {
	// load policies
	policies, err := policy.Parse([]byte(policyYAML))
	if err != nil {
		panic(err)
	}

	// load payloads
	requestJSON := `{
		"name": "Annie",
		"role": "admin",
		"input": {
			"method": "DELETE",
			"path":   "/red-files"
		}
	}`

	var payload any
	if err := json.Unmarshal([]byte(requestJSON), &payload); err != nil {
		panic(err)
	}

	// create a Request
	request := jsonengine.Request{
		Resource: payload,
		Policies: policies,
	}

	// create an engine
	engine := jsonengine.New()

	// apply polices to get the response
	response := engine.Run(context.Background(), request)

	// process the engine response
	logger := log.Default()

	for _, policy := range response.Policies {
		for _, rule := range policy.Rules {
			if rule.Error != nil {
				logger.Printf("error: %s/%s -> %s: %s", policy.Policy.Name, rule.Rule.Name, rule.Identifier, rule.Error)
			} else if len(rule.Violations) != 0 {
				logger.Printf("fail: %s/%s -> %s: %s", policy.Policy.Name, rule.Rule.Name, rule.Identifier, multierr.Combine(rule.Violations...))
			} else {
				logger.Printf("pass: %s/%s -> %s", policy.Policy.Name, rule.Rule.Name, rule.Identifier)
			}
		}
	}
}

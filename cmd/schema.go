package main

import (
	"encoding/json"

	"github.com/alecthomas/jsonschema"
)

type nodeIf struct {
	If struct {
		Properties struct {
			Type struct {
				Const string `json:"const"`
			} `json:"type"`
		} `json:"properties"`
	} `json:"if"`
	Then struct {
		Properties struct {
			Args struct {
				AllOf []struct {
					Ref string `json:"$ref"`
				} `json:"allOf"`
			} `json:"args"`
		} `json:"properties"`
	} `json:"then"`
}

func schema(availableNodes map[string]NewFacility) string {
	ifList := make([]nodeIf, len(availableNodes))
	keys := make([]string, len(availableNodes))
	sc := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-04/schema#",
		"definitions": nil,
		"type":        "object",
		"properties": map[string]interface{}{
			"nodes": map[string]interface{}{
				"type":     "array",
				"required": []string{"name", "type", "args"},
				"minItems": 1,
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":    "string",
							"pattern": "^[a-zA-Z]+[a-zA-Z0-9_]*$",
						},
						"type": map[string]interface{}{
							"type": "string",
							"enum": keys,
						},
						"fetch": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type":    "string",
								"pattern": "^[a-zA-Z]+[a-zA-Z0-9_]*$",
							},
						},
					},
					"allOf": ifList,
				},
			},
		},
	}

	def := make(map[string]*jsonschema.Schema)
	ref := jsonschema.Reflector{ExpandedStruct: true}
	i := 0
	for k, v := range availableNodes {
		def[k] = ref.Reflect(v())
		n := nodeIf{}
		n.If.Properties.Type.Const = k
		n.Then.Properties.Args.AllOf = append(n.Then.Properties.Args.AllOf, struct {
			Ref string "json:\"$ref\""
		}{Ref: "#/definitions/" + k})
		keys[i] = k
		ifList[i] = n
		i++
	}
	sc["definitions"] = def
	b, _ := json.MarshalIndent(sc, "", "   ")
	return string(b)
}

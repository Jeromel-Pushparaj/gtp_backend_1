package spec

import (
	"encoding/json"
	"fmt"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/api"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

// Parser parses OpenAPI specifications
type Parser struct{}

// NewParser creates a new spec parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses an OpenAPI spec from bytes
// Supports both OpenAPI 3.x and Swagger 2.0 (automatically converts to OpenAPI 3.x)
func (p *Parser) Parse(content []byte, format string) (*openapi3.T, error) {
	// First, try to detect if this is Swagger 2.0
	var rawSpec map[string]interface{}
	if err := json.Unmarshal(content, &rawSpec); err == nil {
		if swagger, ok := rawSpec["swagger"].(string); ok && swagger == "2.0" {
			// This is a Swagger 2.0 spec, convert it to OpenAPI 3.0
			return p.convertSwagger2ToOpenAPI3(content)
		}
	}

	// Otherwise, parse as OpenAPI 3.x
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	var doc *openapi3.T
	var err error

	switch format {
	case "yaml", "yml":
		doc, err = loader.LoadFromData(content)
	case "json":
		doc, err = loader.LoadFromData(content)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse spec: %w", err)
	}

	// Validate the spec
	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("spec validation failed: %w", err)
	}

	return doc, nil
}

// convertSwagger2ToOpenAPI3 converts a Swagger 2.0 spec to OpenAPI 3.0
func (p *Parser) convertSwagger2ToOpenAPI3(content []byte) (*openapi3.T, error) {
	// Parse as Swagger 2.0
	var v2Doc openapi2.T
	if err := json.Unmarshal(content, &v2Doc); err != nil {
		return nil, fmt.Errorf("failed to parse Swagger 2.0 spec: %w", err)
	}

	// Convert to OpenAPI 3.0
	v3Doc, err := openapi2conv.ToV3(&v2Doc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Swagger 2.0 to OpenAPI 3.0: %w", err)
	}

	// Always populate servers array from Swagger 2.0 host/basePath
	// The converter might not do this correctly, so we do it explicitly
	if v2Doc.Host != "" {
		scheme := "https"
		if len(v2Doc.Schemes) > 0 {
			scheme = v2Doc.Schemes[0]
		}
		baseURL := fmt.Sprintf("%s://%s%s", scheme, v2Doc.Host, v2Doc.BasePath)

		// Replace any existing servers with the correct one from Swagger 2.0
		v3Doc.Servers = openapi3.Servers{
			&openapi3.Server{
				URL:         baseURL,
				Description: "Converted from Swagger 2.0",
			},
		}
	}

	return v3Doc, nil
}

// ExtractEndpoints extracts all endpoints from the spec
func (p *Parser) ExtractEndpoints(doc *openapi3.T, specID string) ([]*api.Endpoint, error) {
	var endpoints []*api.Endpoint

	if doc.Paths == nil {
		return endpoints, nil
	}

	// Iterate through paths
	for path, pathItem := range doc.Paths {
		if pathItem == nil {
			continue
		}
		for method, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}
			endpoint, err := p.convertOperation(specID, path, method, operation)
			if err != nil {
				return nil, fmt.Errorf("failed to convert operation %s %s: %w", method, path, err)
			}
			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints, nil
}

// convertOperation converts an OpenAPI operation to domain Endpoint
func (p *Parser) convertOperation(specID, path, method string, op *openapi3.Operation) (*api.Endpoint, error) {
	endpoint := &api.Endpoint{
		ID:          uuid.New().String(), // Generate unique ID for endpoint
		SpecID:      specID,
		Path:        path,
		Method:      method,
		OperationID: op.OperationID,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
		Deprecated:  op.Deprecated,
		Responses:   make(map[string]api.Response),
	}

	// Convert parameters
	for _, paramRef := range op.Parameters {
		param := paramRef.Value
		endpoint.Parameters = append(endpoint.Parameters, api.Parameter{
			Name:        param.Name,
			In:          param.In,
			Description: param.Description,
			Required:    param.Required,
			Schema:      p.convertSchema(param.Schema),
		})
	}

	// Convert request body
	if op.RequestBody != nil {
		rb := op.RequestBody.Value
		endpoint.RequestBody = &api.RequestBody{
			Description: rb.Description,
			Required:    rb.Required,
			Content:     p.convertContent(rb.Content),
		}
	}

	// Convert responses
	if op.Responses != nil {
		for status, respRef := range op.Responses {
			if respRef == nil || respRef.Value == nil {
				continue
			}
			resp := respRef.Value
			desc := ""
			if resp.Description != nil {
				desc = *resp.Description
			}
			endpoint.Responses[status] = api.Response{
				Description: desc,
				Content:     p.convertContent(resp.Content),
			}
		}
	}

	// Convert security requirements
	if op.Security != nil {
		for _, secReq := range *op.Security {
			for name, scopes := range secReq {
				endpoint.Security = append(endpoint.Security, api.SecurityRequirement{
					Name:   name,
					Scopes: scopes,
				})
			}
		}
	}

	return endpoint, nil
}

// convertSchema converts OpenAPI schema to domain schema
func (p *Parser) convertSchema(schemaRef *openapi3.SchemaRef) *api.Schema {
	if schemaRef == nil || schemaRef.Value == nil {
		return nil
	}

	schema := schemaRef.Value

	// Convert MinLength and MaxLength from uint64 to *int
	var minLen, maxLen *int
	if schema.MinLength > 0 {
		val := int(schema.MinLength)
		minLen = &val
	}
	if schema.MaxLength != nil {
		val := int(*schema.MaxLength)
		maxLen = &val
	}

	domainSchema := &api.Schema{
		Type:      schema.Type,
		Format:    schema.Format,
		Required:  schema.Required,
		Enum:      schema.Enum,
		Pattern:   schema.Pattern,
		Minimum:   schema.Min,
		Maximum:   schema.Max,
		MinLength: minLen,
		MaxLength: maxLen,
	}

	// Convert properties
	if len(schema.Properties) > 0 {
		domainSchema.Properties = make(map[string]*api.Schema)
		for name, propRef := range schema.Properties {
			domainSchema.Properties[name] = p.convertSchema(propRef)
		}
	}

	// Convert items (for arrays)
	if schema.Items != nil {
		domainSchema.Items = p.convertSchema(schema.Items)
	}

	// Handle $ref
	if schemaRef.Ref != "" {
		domainSchema.Ref = schemaRef.Ref
	}

	return domainSchema
}

// convertContent converts OpenAPI content to domain MediaType map
func (p *Parser) convertContent(content openapi3.Content) map[string]api.MediaType {
	result := make(map[string]api.MediaType)
	for contentType, mediaType := range content {
		result[contentType] = api.MediaType{
			Schema: p.convertSchema(mediaType.Schema),
		}
	}
	return result
}

// ExtractMetadata extracts metadata from the spec
func (p *Parser) ExtractMetadata(doc *openapi3.T, specID string) *api.SpecMetadata {
	metadata := &api.SpecMetadata{
		SpecID:          specID,
		Title:           doc.Info.Title,
		Version:         doc.Info.Version,
		EndpointsByTag:  make(map[string]int),
		SecuritySchemes: make([]string, 0),
	}

	// Extract base URL
	if len(doc.Servers) > 0 {
		metadata.BaseURL = doc.Servers[0].URL
	}

	// Count endpoints
	if doc.Paths != nil {
		metadata.TotalEndpoints = len(doc.Paths) * 4 // Rough estimate

		// Count by tags
		for _, pathItem := range doc.Paths {
			if pathItem == nil {
				continue
			}
			for _, op := range pathItem.Operations() {
				if op == nil {
					continue
				}
				for _, tag := range op.Tags {
					metadata.EndpointsByTag[tag]++
				}
			}
		}
	}

	// Extract security schemes
	if doc.Components != nil && doc.Components.SecuritySchemes != nil {
		for name := range doc.Components.SecuritySchemes {
			metadata.SecuritySchemes = append(metadata.SecuritySchemes, name)
		}
	}

	return metadata
}

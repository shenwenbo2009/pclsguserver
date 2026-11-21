package apidoc

type OpenAPISpec struct {
	OpenAPI    string                    `json:"openapi"`
	Info       InfoObject                `json:"info"`
	Paths      map[string]PathItemObject `json:"paths"`
	Tags       []TagObject               `json:"tags"`
	Components ComponentsObject          `json:"components"`
}

type InfoObject struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

type PathItemObject struct {
	Get    *OperationObject `json:"get,omitempty"`
	Post   *OperationObject `json:"post,omitempty"`
	Put    *OperationObject `json:"put,omitempty"`
	Delete *OperationObject `json:"delete,omitempty"`
}

type OperationObject struct {
	Tags        []string                  `json:"tags"`
	Summary     string                    `json:"summary"`
	Description string                    `json:"description,omitempty"`
	Parameters  []ParameterObject         `json:"parameters,omitempty"`
	RequestBody *RequestBodyObject        `json:"requestBody,omitempty"`
	Responses   map[string]ResponseObject `json:"responses"`
	Security    []map[string][]string     `json:"security,omitempty"`
}

type ParameterObject struct {
	Name        string       `json:"name"`
	In          string       `json:"in"`
	Required    bool         `json:"required"`
	Schema      SchemaObject `json:"schema"`
	Description string       `json:"description,omitempty"`
}

type RequestBodyObject struct {
	Content  map[string]MediaTypeObject `json:"content"`
	Required bool                       `json:"required,omitempty"`
}

type ResponseObject struct {
	Description string                     `json:"description"`
	Content     map[string]MediaTypeObject `json:"content,omitempty"`
}

type MediaTypeObject struct {
	Schema SchemaObject `json:"schema"`
}

type SchemaObject struct {
	Type        string                  `json:"type,omitempty"`
	Properties  map[string]SchemaObject `json:"properties,omitempty"`
	Required    []string                `json:"required,omitempty"`
	Items       *SchemaObject           `json:"items,omitempty"`
	Format      string                  `json:"format,omitempty"`
	Description string                  `json:"description,omitempty"`
	Ref         string                  `json:"$ref,omitempty"`
}

type TagObject struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ComponentsObject struct {
	Schemas         map[string]SchemaObject         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecuritySchemeObject `json:"securitySchemes,omitempty"`
}

type SecuritySchemeObject struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

func BuildSpec(version string) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: InfoObject{
			Title:       "pclsguserver API",
			Version:     version,
			Description: "软件更新分发与插件市场托管服务 API",
		},
		Paths:      make(map[string]PathItemObject),
		Tags:       buildTags(),
		Components: buildComponents(),
	}

	for _, meta := range AllRoutes() {
		op := buildOperation(meta)
		pathItem, exists := spec.Paths[meta.Path]
		if !exists {
			pathItem = PathItemObject{}
		}
		switch meta.Method {
		case "GET":
			pathItem.Get = op
		case "POST":
			pathItem.Post = op
		case "PUT":
			pathItem.Put = op
		case "DELETE":
			pathItem.Delete = op
		}
		spec.Paths[meta.Path] = pathItem
	}

	return spec
}

func buildTags() []TagObject {
	tags := []TagObject{}
	for _, t := range AllTags() {
		tags = append(tags, TagObject{Name: t.Name, Description: t.Description})
	}
	return tags
}

func buildComponents() ComponentsObject {
	return ComponentsObject{
		Schemas: map[string]SchemaObject{
			"Response": {
				Type: "object",
				Properties: map[string]SchemaObject{
					"code":      {Type: "string", Description: "业务状态码"},
					"message":   {Type: "string", Description: "描述信息"},
					"data":      {Description: "业务数据"},
					"requestId": {Type: "string", Description: "请求追踪ID"},
				},
				Required: []string{"code", "message"},
			},
		},
		SecuritySchemes: map[string]SecuritySchemeObject{
			"BearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		},
	}
}

func buildOperation(meta RouteMeta) *OperationObject {
	op := &OperationObject{
		Tags:        []string{meta.Tag},
		Summary:     meta.Summary,
		Description: meta.Description,
		Responses:   buildResponses(meta),
	}

	if len(meta.Parameters) > 0 {
		op.Parameters = buildParameters(meta.Parameters)
	}

	if meta.RequestBody != nil {
		op.RequestBody = buildRequestBody(meta.RequestBody)
	}

	if meta.RequiresAuth {
		op.Security = []map[string][]string{{"BearerAuth": {}}}
	}

	return op
}

func buildParameters(params []ParamMeta) []ParameterObject {
	result := []ParameterObject{}
	for _, p := range params {
		result = append(result, ParameterObject{
			Name:        p.Name,
			In:          p.In,
			Required:    p.Required,
			Schema:      SchemaObject{Type: p.Type, Description: p.Description},
			Description: p.Description,
		})
	}
	return result
}

func buildRequestBody(body *BodyMeta) *RequestBodyObject {
	content := map[string]MediaTypeObject{}
	if body.ContentType == "multipart/form-data" {
		props := map[string]SchemaObject{}
		for _, f := range body.Fields {
			s := SchemaObject{Type: f.Type, Description: f.Description}
			if f.Required {
				props[f.Name] = s
			} else {
				props[f.Name] = s
			}
		}
		content["multipart/form-data"] = MediaTypeObject{
			Schema: SchemaObject{Type: "object", Properties: props},
		}
	} else {
		props := map[string]SchemaObject{}
		required := []string{}
		for _, f := range body.Fields {
			props[f.Name] = SchemaObject{Type: f.Type, Description: f.Description}
			if f.Required {
				required = append(required, f.Name)
			}
		}
		content["application/json"] = MediaTypeObject{
			Schema: SchemaObject{Type: "object", Properties: props, Required: required},
		}
	}
	return &RequestBodyObject{Content: content, Required: body.Required}
}

func buildResponses(meta RouteMeta) map[string]ResponseObject {
	responses := map[string]ResponseObject{}

	for _, r := range meta.Responses {
		statusCode := httpStatusString(r.StatusCode)
		resp := ResponseObject{Description: r.Description}
		if statusCode == "200" {
			resp.Content = map[string]MediaTypeObject{
				"application/json": {
					Schema: SchemaObject{Ref: "#/components/schemas/Response"},
				},
			}
		}
		responses[statusCode] = resp
	}

	return responses
}

func httpStatusString(code int) string {
	switch code {
	case 200:
		return "200"
	case 400:
		return "400"
	case 401:
		return "401"
	case 403:
		return "403"
	case 404:
		return "404"
	case 409:
		return "409"
	case 423:
		return "423"
	case 429:
		return "429"
	case 500:
		return "500"
	case 503:
		return "503"
	default:
		return "200"
	}
}

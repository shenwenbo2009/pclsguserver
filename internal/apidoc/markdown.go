package apidoc

import (
	"fmt"
	"strings"
)

func GenerateMarkdown(spec *OpenAPISpec) string {
	if spec == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("# pclsguserver API 文档\n\n")
	sb.WriteString("软件更新分发与插件市场托管服务 API\n\n")

	tagRoutes := groupByTag(spec)

	for _, tag := range spec.Tags {
		routes, ok := tagRoutes[tag.Name]
		if !ok {
			continue
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n", tag.Name))
		if tag.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", tag.Description))
		}

		for _, route := range routes {
			writeRouteMarkdown(&sb, route)
		}
	}

	return sb.String()
}

type routeInfo struct {
	method string
	path   string
	op     *OperationObject
}

func groupByTag(spec *OpenAPISpec) map[string][]routeInfo {
	result := map[string][]routeInfo{}
	for path, item := range spec.Paths {
		if item.Get != nil {
			for _, tag := range item.Get.Tags {
				result[tag] = append(result[tag], routeInfo{method: "GET", path: path, op: item.Get})
			}
		}
		if item.Post != nil {
			for _, tag := range item.Post.Tags {
				result[tag] = append(result[tag], routeInfo{method: "POST", path: path, op: item.Post})
			}
		}
		if item.Put != nil {
			for _, tag := range item.Put.Tags {
				result[tag] = append(result[tag], routeInfo{method: "PUT", path: path, op: item.Put})
			}
		}
		if item.Delete != nil {
			for _, tag := range item.Delete.Tags {
				result[tag] = append(result[tag], routeInfo{method: "DELETE", path: path, op: item.Delete})
			}
		}
	}
	return result
}

func writeRouteMarkdown(sb *strings.Builder, route routeInfo) {
	op := route.op
	sb.WriteString(fmt.Sprintf("### %s %s — %s\n\n", route.method, route.path, op.Summary))

	if op.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", op.Description))
	}

	if len(op.Security) > 0 {
		sb.WriteString("> 需要认证 (Bearer Token)\n\n")
	}

	if len(op.Parameters) > 0 {
		sb.WriteString("**参数**\n\n")
		sb.WriteString("| 参数名 | 位置 | 类型 | 必填 | 说明 |\n")
		sb.WriteString("|--------|------|------|------|------|\n")
		for _, p := range op.Parameters {
			required := ""
			if p.Required {
				required = "是"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				p.Name, p.In, p.Schema.Type, required, p.Description))
		}
		sb.WriteString("\n")
	}

	if op.RequestBody != nil {
		sb.WriteString("**请求体**\n\n")
		for contentType, media := range op.RequestBody.Content {
			sb.WriteString(fmt.Sprintf("Content-Type: `%s`\n\n", contentType))
			if media.Schema.Type == "object" && len(media.Schema.Properties) > 0 {
				sb.WriteString("| 字段 | 类型 | 必填 | 说明 |\n")
				sb.WriteString("|------|------|------|------|\n")
				for name, prop := range media.Schema.Properties {
					required := ""
					for _, r := range media.Schema.Required {
						if r == name {
							required = "是"
							break
						}
					}
					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
						name, prop.Type, required, prop.Description))
				}
				sb.WriteString("\n")
			}
		}
	}

	sb.WriteString("**请求示例**\n\n")
	sb.WriteString(fmt.Sprintf("```\n%s %s\n```\n\n", route.method, route.path))

	if op.RequestBody != nil {
		for contentType := range op.RequestBody.Content {
			if contentType == "application/json" {
				sb.WriteString("```json\n")
				sb.WriteString(writeRequestExample(route.op))
				sb.WriteString("\n```\n\n")
			}
			break
		}
	}

	sb.WriteString("**响应示例**\n\n")
	sb.WriteString("```json\n")
	sb.WriteString(`{
  "code": "OK",
  "message": "success",
  "data": {}
}`)
	sb.WriteString("\n```\n\n")

	if len(op.Responses) > 1 {
		sb.WriteString("**错误码**\n\n")
		sb.WriteString("| 错误码 | HTTP 状态码 | 说明 |\n")
		sb.WriteString("|--------|------------|------|\n")
		for status, resp := range op.Responses {
			if status == "200" {
				continue
			}
			sb.WriteString(fmt.Sprintf("| — | %s | %s |\n", status, resp.Description))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")
}

func writeRequestExample(op *OperationObject) string {
	if op.RequestBody == nil {
		return ""
	}
	for _, media := range op.RequestBody.Content {
		if media.Schema.Type == "object" && len(media.Schema.Properties) > 0 {
			var sb strings.Builder
			sb.WriteString("{\n")
			first := true
			for name, prop := range media.Schema.Properties {
				if !first {
					sb.WriteString(",\n")
				}
				first = false
				if name == "password" {
					sb.WriteString(fmt.Sprintf(`  "%s": "***"`, name))
				} else {
					sb.WriteString(fmt.Sprintf(`  "%s": "%s"`, name, prop.Type))
				}
			}
			sb.WriteString("\n}")
			return sb.String()
		}
		break
	}
	return "{}"
}

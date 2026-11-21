package apidoc

type RouteMeta struct {
	Method       string
	Path         string
	Tag          string
	Summary      string
	Description  string
	Parameters   []ParamMeta
	RequestBody  *BodyMeta
	Responses    []ResponseMeta
	RequiresAuth bool
}

type ParamMeta struct {
	Name        string
	In          string
	Type        string
	Required    bool
	Description string
}

type BodyMeta struct {
	ContentType string
	Required    bool
	Fields      []FieldMeta
}

type FieldMeta struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

type ResponseMeta struct {
	StatusCode  int
	Code        string
	Description string
}

type TagDef struct {
	Name        string
	Description string
}

func AllTags() []TagDef {
	return []TagDef{
		{Name: "软件更新服务", Description: "软件版本检查、下载与上报"},
		{Name: "插件市场", Description: "插件列表、搜索、详情与下载"},
		{Name: "管理后台-认证", Description: "管理员登录、令牌刷新、登出、权限查询"},
		{Name: "管理后台-软件版本管理", Description: "软件版本发布、下架、删除"},
		{Name: "管理后台-插件管理", Description: "插件提交、下架、删除、重新提交"},
		{Name: "管理后台-审核", Description: "插件审核列表与审核操作"},
		{Name: "管理后台-索引管理", Description: "索引重建与状态查询"},
		{Name: "管理后台-统计", Description: "指标维度查询"},
		{Name: "管理后台-审计", Description: "审计日志查询"},
		{Name: "管理后台-用户管理", Description: "管理员账号 CRUD"},
		{Name: "管理后台-角色管理", Description: "角色 CRUD 与权限分配"},
		{Name: "管理后台-配置管理", Description: "系统配置读写"},
		{Name: "系统", Description: "健康检查与索引文件"},
	}
}

func AllRoutes() []RouteMeta {
	return []RouteMeta{
		// 软件更新服务
		{
			Method: "GET", Path: "/api/update/check", Tag: "软件更新服务",
			Summary: "检查软件更新",
			Parameters: []ParamMeta{
				{Name: "swCode", In: "query", Type: "string", Required: true, Description: "软件编码"},
				{Name: "currentVersion", In: "query", Type: "string", Required: true, Description: "当前版本号"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "更新检查结果"},
				{StatusCode: 400, Code: "INVALID_PARAM", Description: "参数缺失"},
				{StatusCode: 404, Code: "SW_NOT_FOUND", Description: "软件不存在"},
			},
		},
		{
			Method: "GET", Path: "/api/update/download", Tag: "软件更新服务",
			Summary: "下载更新包",
			Parameters: []ParamMeta{
				{Name: "swCode", In: "query", Type: "string", Required: true, Description: "软件编码"},
				{Name: "version", In: "query", Type: "string", Required: true, Description: "版本号"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "更新包二进制流"},
				{StatusCode: 400, Code: "INVALID_PARAM", Description: "参数缺失"},
				{StatusCode: 404, Code: "VERSION_NOT_FOUND", Description: "版本不存在"},
			},
		},
		{
			Method: "POST", Path: "/api/update/report", Tag: "软件更新服务",
			Summary: "上报更新事件",
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "swCode", Type: "string", Required: true, Description: "软件编码"},
					{Name: "version", Type: "string", Required: true, Description: "版本号"},
					{Name: "eventType", Type: "string", Required: true, Description: "事件类型(query/download)"},
					{Name: "clientId", Type: "string", Required: false, Description: "客户端ID"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "上报成功"},
				{StatusCode: 400, Code: "INVALID_PARAM", Description: "参数错误"},
			},
		},

		// 插件市场
		{
			Method: "GET", Path: "/api/plugins", Tag: "插件市场",
			Summary: "插件列表",
			Parameters: []ParamMeta{
				{Name: "page", In: "query", Type: "integer", Required: false, Description: "页码(默认1)"},
				{Name: "size", In: "query", Type: "integer", Required: false, Description: "每页数量(默认20)"},
				{Name: "category", In: "query", Type: "string", Required: false, Description: "分类筛选"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "插件列表"},
			},
		},
		{
			Method: "GET", Path: "/api/plugins/search", Tag: "插件市场",
			Summary: "搜索插件",
			Parameters: []ParamMeta{
				{Name: "keyword", In: "query", Type: "string", Required: false, Description: "搜索关键词"},
				{Name: "category", In: "query", Type: "string", Required: false, Description: "分类筛选"},
				{Name: "name", In: "query", Type: "string", Required: false, Description: "插件名称"},
				{Name: "page", In: "query", Type: "integer", Required: false, Description: "页码"},
				{Name: "size", In: "query", Type: "integer", Required: false, Description: "每页数量"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "搜索结果"},
				{StatusCode: 503, Code: "INDEX_UNAVAILABLE", Description: "索引未就绪"},
			},
		},
		{
			Method: "GET", Path: "/api/plugins/{id}", Tag: "插件市场",
			Summary: "插件详情",
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "插件ID"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "插件详情"},
				{StatusCode: 404, Code: "PLUGIN_NOT_FOUND", Description: "插件不存在"},
			},
		},
		{
			Method: "GET", Path: "/api/plugins/{id}/download", Tag: "插件市场",
			Summary: "下载插件",
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "插件ID"},
				{Name: "version", In: "query", Type: "string", Required: false, Description: "版本号"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "插件二进制流"},
				{StatusCode: 404, Code: "PLUGIN_NOT_FOUND", Description: "插件不存在"},
			},
		},

		// 管理后台-认证
		{
			Method: "POST", Path: "/api/admin/login", Tag: "管理后台-认证",
			Summary: "管理员登录",
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "username", Type: "string", Required: true, Description: "用户名"},
					{Name: "password", Type: "string", Required: true, Description: "密码"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "登录成功，返回令牌"},
				{StatusCode: 401, Code: "UNAUTHORIZED", Description: "凭证无效"},
				{StatusCode: 423, Code: "ACCOUNT_LOCKED", Description: "账号锁定"},
			},
		},
		{
			Method: "POST", Path: "/api/admin/auth/refresh", Tag: "管理后台-认证",
			Summary: "刷新令牌",
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "refreshToken", Type: "string", Required: true, Description: "刷新令牌"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "刷新成功"},
				{StatusCode: 401, Code: "UNAUTHORIZED", Description: "刷新令牌无效"},
			},
		},
		{
			Method: "POST", Path: "/api/admin/logout", Tag: "管理后台-认证",
			Summary: "登出",
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "登出成功"},
			},
		},
		{
			Method: "GET", Path: "/api/admin/auth/permissions", Tag: "管理后台-认证",
			Summary:      "查询当前用户权限",
			RequiresAuth: true,
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "权限列表"},
				{StatusCode: 401, Code: "UNAUTHORIZED", Description: "未认证"},
			},
		},

		// 管理后台-软件版本管理
		{
			Method: "GET", Path: "/api/admin/software/versions", Tag: "管理后台-软件版本管理",
			Summary: "软件版本列表", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "page", In: "query", Type: "integer", Required: false, Description: "页码"},
				{Name: "size", In: "query", Type: "integer", Required: false, Description: "每页数量"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "版本列表"},
				{StatusCode: 401, Code: "UNAUTHORIZED", Description: "未认证"},
			},
		},
		{
			Method: "POST", Path: "/api/admin/software/versions", Tag: "管理后台-软件版本管理",
			Summary: "发布软件版本", RequiresAuth: true,
			RequestBody: &BodyMeta{
				ContentType: "multipart/form-data", Required: true,
				Fields: []FieldMeta{
					{Name: "package", Type: "file", Required: true, Description: "软件包文件"},
					{Name: "swCode", Type: "string", Required: true, Description: "软件编码"},
					{Name: "version", Type: "string", Required: true, Description: "版本号"},
					{Name: "releaseNote", Type: "string", Required: false, Description: "发布说明"},
					{Name: "swUrl", Type: "string", Required: false, Description: "软件URL"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "发布成功"},
				{StatusCode: 400, Code: "INVALID_PARAM", Description: "参数错误"},
				{StatusCode: 500, Code: "INTERNAL_ERROR", Description: "发布失败"},
			},
		},
		{
			Method: "PUT", Path: "/api/admin/software/versions/{id}/offline", Tag: "管理后台-软件版本管理",
			Summary: "下架软件版本", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "版本ID"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "下架成功"},
				{StatusCode: 404, Code: "SW_NOT_FOUND", Description: "版本不存在"},
			},
		},
		{
			Method: "DELETE", Path: "/api/admin/software/versions/{id}", Tag: "管理后台-软件版本管理",
			Summary: "删除软件版本", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "版本ID"},
				{Name: "confirm", In: "query", Type: "string", Required: false, Description: "确认为true时执行删除"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "删除成功"},
				{StatusCode: 400, Code: "CONFIRMATION_REQUIRED", Description: "需要确认"},
				{StatusCode: 404, Code: "SW_NOT_FOUND", Description: "版本不存在"},
			},
		},

		// 管理后台-插件管理
		{
			Method: "GET", Path: "/api/admin/plugins", Tag: "管理后台-插件管理",
			Summary: "插件列表(管理)", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "page", In: "query", Type: "integer", Required: false, Description: "页码"},
				{Name: "size", In: "query", Type: "integer", Required: false, Description: "每页数量"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "插件列表"},
			},
		},
		{
			Method: "POST", Path: "/api/admin/plugins", Tag: "管理后台-插件管理",
			Summary: "提交插件", RequiresAuth: true,
			RequestBody: &BodyMeta{
				ContentType: "multipart/form-data", Required: true,
				Fields: []FieldMeta{
					{Name: "package", Type: "file", Required: true, Description: "插件包文件"},
					{Name: "name", Type: "string", Required: true, Description: "插件名称"},
					{Name: "version", Type: "string", Required: true, Description: "版本号"},
					{Name: "author", Type: "string", Required: false, Description: "作者"},
					{Name: "category", Type: "string", Required: false, Description: "分类"},
					{Name: "description", Type: "string", Required: false, Description: "描述"},
					{Name: "downloadUrl", Type: "string", Required: false, Description: "下载URL"},
					{Name: "keywords", Type: "string", Required: false, Description: "关键词(逗号分隔)"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "提交成功"},
				{StatusCode: 400, Code: "INVALID_PARAM", Description: "参数错误"},
			},
		},
		{
			Method: "PUT", Path: "/api/admin/plugins/{id}/offline", Tag: "管理后台-插件管理",
			Summary: "下架插件", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "插件ID"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "下架成功"},
				{StatusCode: 404, Code: "PLUGIN_NOT_FOUND", Description: "插件不存在"},
			},
		},
		{
			Method: "DELETE", Path: "/api/admin/plugins/{id}", Tag: "管理后台-插件管理",
			Summary: "删除插件", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "插件ID"},
				{Name: "confirm", In: "query", Type: "string", Required: false, Description: "确认为true时执行删除"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "删除成功"},
				{StatusCode: 400, Code: "CONFIRMATION_REQUIRED", Description: "需要确认"},
				{StatusCode: 404, Code: "PLUGIN_NOT_FOUND", Description: "插件不存在"},
			},
		},
		{
			Method: "POST", Path: "/api/admin/plugins/{id}/resubmit", Tag: "管理后台-插件管理",
			Summary: "重新提交插件", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "插件ID"},
			},
			RequestBody: &BodyMeta{
				ContentType: "multipart/form-data", Required: true,
				Fields: []FieldMeta{
					{Name: "package", Type: "file", Required: true, Description: "插件包文件"},
					{Name: "name", Type: "string", Required: false, Description: "插件名称"},
					{Name: "version", Type: "string", Required: false, Description: "版本号"},
					{Name: "author", Type: "string", Required: false, Description: "作者"},
					{Name: "category", Type: "string", Required: false, Description: "分类"},
					{Name: "description", Type: "string", Required: false, Description: "描述"},
					{Name: "downloadUrl", Type: "string", Required: false, Description: "下载URL"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "重新提交成功"},
				{StatusCode: 400, Code: "INVALID_STATUS", Description: "状态不允许重新提交"},
			},
		},

		// 管理后台-审核
		{
			Method: "GET", Path: "/api/admin/plugins/reviewing", Tag: "管理后台-审核",
			Summary: "待审核插件列表", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "page", In: "query", Type: "integer", Required: false, Description: "页码"},
				{Name: "size", In: "query", Type: "integer", Required: false, Description: "每页数量"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "待审核列表"},
			},
		},
		{
			Method: "POST", Path: "/api/admin/plugins/{id}/review", Tag: "管理后台-审核",
			Summary: "审核插件", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "插件ID"},
			},
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "pass", Type: "boolean", Required: true, Description: "是否通过"},
					{Name: "reason", Type: "string", Required: false, Description: "驳回原因"},
					{Name: "comment", Type: "string", Required: false, Description: "审核备注"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "审核完成"},
				{StatusCode: 404, Code: "PLUGIN_NOT_FOUND", Description: "插件不存在"},
			},
		},

		// 管理后台-索引管理
		{
			Method: "POST", Path: "/api/admin/index/rebuild", Tag: "管理后台-索引管理",
			Summary: "重建索引", RequiresAuth: true,
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "重建完成"},
				{StatusCode: 409, Code: "REBUILD_IN_PROGRESS", Description: "重建进行中"},
			},
		},
		{
			Method: "GET", Path: "/api/admin/index/status", Tag: "管理后台-索引管理",
			Summary: "索引状态", RequiresAuth: true,
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "索引状态信息"},
			},
		},

		// 管理后台-统计
		{
			Method: "GET", Path: "/api/admin/metrics", Tag: "管理后台-统计",
			Summary: "统计查询", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "dimension", In: "query", Type: "string", Required: false, Description: "维度"},
				{Name: "granularity", In: "query", Type: "string", Required: false, Description: "粒度"},
				{Name: "subjectKey", In: "query", Type: "string", Required: false, Description: "主体键"},
				{Name: "from", In: "query", Type: "string", Required: false, Description: "开始时间(RFC3339)"},
				{Name: "to", In: "query", Type: "string", Required: false, Description: "结束时间(RFC3339)"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "统计数据"},
				{StatusCode: 400, Code: "QUERY_RANGE_TOO_LARGE", Description: "查询范围过大(最大30天)"},
			},
		},

		// 管理后台-审计
		{
			Method: "GET", Path: "/api/admin/audit", Tag: "管理后台-审计",
			Summary: "审计日志查询", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "from", In: "query", Type: "string", Required: false, Description: "开始时间(RFC3339)"},
				{Name: "to", In: "query", Type: "string", Required: false, Description: "结束时间(RFC3339)"},
				{Name: "action", In: "query", Type: "string", Required: false, Description: "操作类型"},
				{Name: "operator", In: "query", Type: "string", Required: false, Description: "操作人"},
				{Name: "targetType", In: "query", Type: "string", Required: false, Description: "对象类型"},
				{Name: "page", In: "query", Type: "integer", Required: false, Description: "页码"},
				{Name: "size", In: "query", Type: "integer", Required: false, Description: "每页数量"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "审计日志列表"},
				{StatusCode: 400, Code: "QUERY_RANGE_TOO_LARGE", Description: "查询范围过大(最大7天)"},
			},
		},

		// 管理后台-用户管理
		{
			Method: "GET", Path: "/api/admin/users", Tag: "管理后台-用户管理",
			Summary: "用户列表", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "page", In: "query", Type: "integer", Required: false, Description: "页码"},
				{Name: "size", In: "query", Type: "integer", Required: false, Description: "每页数量"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "用户列表"},
			},
		},
		{
			Method: "POST", Path: "/api/admin/users", Tag: "管理后台-用户管理",
			Summary: "创建用户", RequiresAuth: true,
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "username", Type: "string", Required: true, Description: "用户名"},
					{Name: "password", Type: "string", Required: true, Description: "密码"},
					{Name: "roleIds", Type: "array", Required: false, Description: "角色ID列表"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "创建成功"},
				{StatusCode: 409, Code: "USER_ALREADY_EXISTS", Description: "用户已存在"},
			},
		},
		{
			Method: "GET", Path: "/api/admin/users/{id}", Tag: "管理后台-用户管理",
			Summary: "用户详情", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "用户ID"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "用户详情"},
				{StatusCode: 404, Code: "USER_NOT_FOUND", Description: "用户不存在"},
			},
		},
		{
			Method: "PUT", Path: "/api/admin/users/{id}/disable", Tag: "管理后台-用户管理",
			Summary: "禁用用户", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "用户ID"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "禁用成功"},
				{StatusCode: 409, Code: "PROTECTED_USER", Description: "受保护用户不可禁用"},
			},
		},
		{
			Method: "PUT", Path: "/api/admin/users/{id}/enable", Tag: "管理后台-用户管理",
			Summary: "启用用户", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "用户ID"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "启用成功"},
				{StatusCode: 404, Code: "USER_NOT_FOUND", Description: "用户不存在"},
			},
		},
		{
			Method: "DELETE", Path: "/api/admin/users/{id}", Tag: "管理后台-用户管理",
			Summary: "删除用户", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "用户ID"},
				{Name: "confirm", In: "query", Type: "string", Required: false, Description: "确认为true时执行删除"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "删除成功"},
				{StatusCode: 400, Code: "CONFIRMATION_REQUIRED", Description: "需要确认"},
				{StatusCode: 409, Code: "PROTECTED_USER", Description: "受保护用户不可删除"},
			},
		},
		{
			Method: "PUT", Path: "/api/admin/users/{id}/roles", Tag: "管理后台-用户管理",
			Summary: "分配角色", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "用户ID"},
			},
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "roleIds", Type: "array", Required: true, Description: "角色ID列表"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "分配成功"},
			},
		},

		// 管理后台-角色管理
		{
			Method: "GET", Path: "/api/admin/roles", Tag: "管理后台-角色管理",
			Summary: "角色列表", RequiresAuth: true,
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "角色列表"},
			},
		},
		{
			Method: "POST", Path: "/api/admin/roles", Tag: "管理后台-角色管理",
			Summary: "创建角色", RequiresAuth: true,
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "name", Type: "string", Required: true, Description: "角色名称"},
					{Name: "permissions", Type: "array", Required: true, Description: "权限列表"},
					{Name: "description", Type: "string", Required: false, Description: "描述"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "创建成功"},
			},
		},
		{
			Method: "PUT", Path: "/api/admin/roles/{id}", Tag: "管理后台-角色管理",
			Summary: "更新角色", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "角色ID"},
			},
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "permissions", Type: "array", Required: true, Description: "权限列表"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "更新成功"},
			},
		},
		{
			Method: "DELETE", Path: "/api/admin/roles/{id}", Tag: "管理后台-角色管理",
			Summary: "删除角色", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "id", In: "path", Type: "string", Required: true, Description: "角色ID"},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "删除成功"},
				{StatusCode: 409, Code: "INVALID_STATUS", Description: "内置角色不可删除"},
			},
		},

		// 管理后台-配置管理
		{
			Method: "GET", Path: "/api/admin/config", Tag: "管理后台-配置管理",
			Summary: "配置列表", RequiresAuth: true,
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "配置列表"},
			},
		},
		{
			Method: "PUT", Path: "/api/admin/config/{key}", Tag: "管理后台-配置管理",
			Summary: "更新配置", RequiresAuth: true,
			Parameters: []ParamMeta{
				{Name: "key", In: "path", Type: "string", Required: true, Description: "配置键"},
			},
			RequestBody: &BodyMeta{
				ContentType: "application/json", Required: true,
				Fields: []FieldMeta{
					{Name: "value", Type: "string", Required: true, Description: "配置值"},
					{Name: "confirm", Type: "boolean", Required: false, Description: "敏感配置需确认"},
				},
			},
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "更新成功"},
				{StatusCode: 400, Code: "CONFIRMATION_REQUIRED", Description: "敏感配置需确认"},
				{StatusCode: 400, Code: "INVALID_CONFIG_VALUE", Description: "配置值无效"},
				{StatusCode: 404, Code: "CONFIG_NOT_FOUND", Description: "配置不存在"},
			},
		},

		// 系统
		{
			Method: "GET", Path: "/health", Tag: "系统",
			Summary: "健康检查",
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "服务状态"},
			},
		},
		{
			Method: "GET", Path: "/index.json", Tag: "系统",
			Summary: "索引文件",
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "索引JSON"},
			},
		},
		{
			Method: "GET", Path: "/static/index.json", Tag: "系统",
			Summary: "静态索引文件",
			Responses: []ResponseMeta{
				{StatusCode: 200, Code: "OK", Description: "静态索引JSON"},
			},
		},
	}
}

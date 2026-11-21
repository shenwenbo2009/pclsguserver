package init

import (
	"strings"
)

func ValidateStep(step int, data map[string]interface{}) (valid bool, errors []string) {
	switch step {
	case 1:
		return validateStep1(data)
	case 2:
		return validateStep2(data)
	case 3:
		return validateStep3(data)
	case 4:
		return validateStep4(data)
	default:
		return false, []string{"unknown step"}
	}
}

func validateStep1(data map[string]interface{}) (valid bool, errors []string) {
	username, _ := data["adminUsername"].(string)
	password, _ := data["adminPassword"].(string)
	passwordConfirm, _ := data["passwordConfirm"].(string)

	if len(username) < 3 {
		errors = append(errors, "用户名至少 3 字符")
	}
	if len(username) > 50 {
		errors = append(errors, "用户名最多 50 字符")
	}
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			errors = append(errors, "用户名仅允许字母、数字、下划线")
			break
		}
	}
	if username == "root" || username == "system" {
		errors = append(errors, "用户名不能为保留名 root 或 system")
	}
	if len(password) < 8 {
		errors = append(errors, "密码至少 8 字符")
	}
	hasLetter := false
	hasDigit := false
	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		errors = append(errors, "密码需包含至少一个字母和一个数字")
	}
	if password == "password" || password == "12345678" {
		errors = append(errors, "密码过于简单")
	}
	if password != passwordConfirm {
		errors = append(errors, "两次输入的密码不一致")
	}

	return len(errors) == 0, errors
}

func validateStep2(data map[string]interface{}) (valid bool, errors []string) {
	jwtMode, _ := data["jwtMode"].(string)
	jwtSecret, _ := data["jwtSecret"].(string)

	if jwtMode == "auto" {
		return true, nil
	}
	if jwtMode == "manual" {
		if len(jwtSecret) < 16 {
			errors = append(errors, "手动输入的 JWT 密钥至少 16 字符")
		}
		if jwtSecret == "change-me-in-production" {
			errors = append(errors, "不能使用默认占位密钥")
		}
		return len(errors) == 0, errors
	}
	return false, []string{"jwtMode 必须为 auto 或 manual"}
}

func validateStep3(data map[string]interface{}) (valid bool, errors []string) {
	dbDriver, _ := data["dbDriver"].(string)

	if dbDriver != "sqlite" && dbDriver != "postgres" && dbDriver != "mysql" {
		return false, []string{"dbDriver 必须为 sqlite、postgres 或 mysql"}
	}

	if dbDriver == "sqlite" {
		sqlitePath, _ := data["sqlitePath"].(string)
		if strings.TrimSpace(sqlitePath) == "" {
			errors = append(errors, "SQLite 路径不能为空")
		}
	} else if dbDriver == "postgres" || dbDriver == "mysql" {
		dbHost, _ := data["dbHost"].(string)
		dbPort, _ := data["dbPort"].(string)
		dbName, _ := data["dbName"].(string)
		dbUser, _ := data["dbUser"].(string)
		dbPassword, _ := data["dbPassword"].(string)
		if strings.TrimSpace(dbHost) == "" {
			errors = append(errors, "数据库地址不能为空")
		}
		if strings.TrimSpace(dbPort) == "" {
			errors = append(errors, "数据库端口不能为空")
		}
		if strings.TrimSpace(dbName) == "" {
			errors = append(errors, "数据库名不能为空")
		}
		if strings.TrimSpace(dbUser) == "" {
			errors = append(errors, "数据库用户名不能为空")
		}
		if strings.TrimSpace(dbPassword) == "" {
			errors = append(errors, "数据库密码不能为空")
		}
	}

	return len(errors) == 0, errors
}

func validateStep4(data map[string]interface{}) (valid bool, errors []string) {
	storageDriver, _ := data["storageDriver"].(string)

	if storageDriver != "fs" && storageDriver != "s3" {
		return false, []string{"storageDriver 必须为 fs 或 s3"}
	}

	if storageDriver == "fs" {
		storagePath, _ := data["storagePath"].(string)
		if strings.TrimSpace(storagePath) == "" {
			errors = append(errors, "存储路径不能为空")
		}
	} else {
		s3Bucket, _ := data["s3Bucket"].(string)
		s3Region, _ := data["s3Region"].(string)
		s3Endpoint, _ := data["s3Endpoint"].(string)
		if strings.TrimSpace(s3Bucket) == "" {
			errors = append(errors, "S3 Bucket 不能为空")
		}
		if strings.TrimSpace(s3Region) == "" {
			errors = append(errors, "S3 Region 不能为空")
		}
		if strings.TrimSpace(s3Endpoint) == "" {
			errors = append(errors, "S3 Endpoint 不能为空")
		}
	}

	return len(errors) == 0, errors
}

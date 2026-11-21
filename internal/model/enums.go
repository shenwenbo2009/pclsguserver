package model

type SoftwareStatus string

const (
	SoftwareStatusDraft     SoftwareStatus = "draft"
	SoftwareStatusPublished SoftwareStatus = "published"
	SoftwareStatusOffline   SoftwareStatus = "offline"
)

type PluginStatus string

const (
	PluginStatusDraft     PluginStatus = "draft"
	PluginStatusReviewing PluginStatus = "reviewing"
	PluginStatusPublished PluginStatus = "published"
	PluginStatusRejected  PluginStatus = "rejected"
	PluginStatusOffline   PluginStatus = "offline"
)

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusDisabled AccountStatus = "disabled"
	AccountStatusLocked   AccountStatus = "locked"
)

type AuditAction string

const (
	AuditActionLogin            AuditAction = "LOGIN"
	AuditActionPublish          AuditAction = "PUBLISH"
	AuditActionOffline          AuditAction = "OFFLINE"
	AuditActionDelete           AuditAction = "DELETE"
	AuditActionApprove          AuditAction = "APPROVE"
	AuditActionReject           AuditAction = "REJECT"
	AuditActionIndexRebuild     AuditAction = "INDEX_REBUILD"
	AuditActionConfigChanged    AuditAction = "CONFIG_CHANGED"
	AuditActionUserRoleChanged  AuditAction = "USER_ROLE_CHANGED"
	AuditActionPluginSubmitted  AuditAction = "PLUGIN_SUBMITTED"
	AuditActionPluginUploaded   AuditAction = "PLUGIN_UPLOADED"
	AuditActionSoftwareUploaded AuditAction = "SOFTWARE_UPLOADED"
	AuditActionFileUploaded     AuditAction = "FILE_UPLOADED"
	AuditActionTokenCreated     AuditAction = "TOKEN_CREATED"
	AuditActionTokenDisabled    AuditAction = "TOKEN_DISABLED"
)

type AuditTargetType string

const (
	AuditTargetSoftware    AuditTargetType = "software"
	AuditTargetPlugin      AuditTargetType = "plugin"
	AuditTargetIndex       AuditTargetType = "index"
	AuditTargetConfig      AuditTargetType = "config"
	AuditTargetUser        AuditTargetType = "user"
	AuditTargetRole        AuditTargetType = "role"
	AuditTargetUploadToken AuditTargetType = "upload_token"
	AuditTargetFile        AuditTargetType = "file"
)

type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
)

type MetricsDimension string

const (
	MetricsDimensionSoftwareDownload MetricsDimension = "software_download"
	MetricsDimensionPluginDownload   MetricsDimension = "plugin_download"
	MetricsDimensionSoftwareActivity MetricsDimension = "software_activity"
	MetricsDimensionPluginActivity   MetricsDimension = "plugin_activity"
)

type Granularity string

const (
	GranularityHour Granularity = "hour"
	GranularityDay  Granularity = "day"
)

type ReviewConclusion string

const (
	ReviewConclusionPass   ReviewConclusion = "pass"
	ReviewConclusionReject ReviewConclusion = "reject"
)

type UploadTokenStatus string

const (
	UploadTokenStatusActive   UploadTokenStatus = "active"
	UploadTokenStatusDisabled UploadTokenStatus = "disabled"
)

type FilePurpose string

const (
	FilePurposeAvatar     FilePurpose = "avatar"
	FilePurposeAttachment FilePurpose = "attachment"
	FilePurposeScreenshot FilePurpose = "screenshot"
	FilePurposeDocument   FilePurpose = "document"
	FilePurposeOther      FilePurpose = "other"
)

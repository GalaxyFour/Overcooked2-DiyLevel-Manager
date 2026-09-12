package domain

import "time"

type Role string

const (
	RoleAuthor      Role = "author"
	RoleAdmin       Role = "admin"
	RoleSuperAdmin  Role = "super_admin"
)

type SetStatus string

const (
	SetStatusActive  SetStatus = "active"
	SetStatusHidden  SetStatus = "hidden"
	SetStatusInvalid SetStatus = "invalid"
)

type ParseStatus string

const (
	ParseStatusQueued  ParseStatus = "queued"
	ParseStatusRunning ParseStatus = "running"
	ParseStatusDone    ParseStatus = "done"
	ParseStatusFailed  ParseStatus = "failed"
)

type JobStatus string

const (
	JobStatusQueued  JobStatus = "queued"
	JobStatusRunning JobStatus = "running"
	JobStatusDone    JobStatus = "done"
	JobStatusFailed  JobStatus = "failed"
)

type PresignKind string

const (
	PresignKindPackage PresignKind = "package"
	PresignKindImage   PresignKind = "image"
)

type User struct {
	ID                   int64      `json:"id"`
	Username             string     `json:"username"`
	PasswordMD5          string     `json:"-"`
	Role                 Role       `json:"role"`
	MustChangePassword   bool       `json:"mustChangePassword"`
	DisplayName          string     `json:"displayName"`
	AvatarCOSKey         string     `json:"avatarCosKey,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	DisabledAt           *time.Time `json:"disabledAt,omitempty"`
}

type LevelSet struct {
	ID           int64     `json:"id"`
	Slug         string    `json:"slug"`
	AuthorID     int64     `json:"authorId"`
	Name         string    `json:"name"`
	NameZH       string    `json:"nameZh"`
	Status       SetStatus `json:"status"`
	CoverCOSKey  string    `json:"coverCosKey,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	AuthorName   string    `json:"authorName,omitempty"`
	LatestVersion string   `json:"latestVersion,omitempty"`
	LevelCount   int       `json:"levelCount,omitempty"`
	CoverURL     string    `json:"coverUrl,omitempty"`
}

type LevelSetVersion struct {
	ID            int64       `json:"id"`
	SetID         int64       `json:"setId"`
	Version       string      `json:"version"`
	UID           string      `json:"uid"`
	ZipCOSKey     string      `json:"zipCosKey,omitempty"`
	FileCount     int         `json:"fileCount"`
	HasRuntime    bool        `json:"hasRuntime"`
	HasCommonW2   bool        `json:"hasCommonW2"`
	ParseStatus   ParseStatus `json:"parseStatus"`
	ParsePhase    string      `json:"parsePhase"`
	ParseProgress int         `json:"parseProgress"`
	ParseMessage  string      `json:"parseMessage"`
	IsLatest      bool        `json:"isLatest"`
	CreatedAt     time.Time   `json:"createdAt"`
}

type LevelEntry struct {
	ID               int64  `json:"id"`
	VersionID        int64  `json:"versionId"`
	LevelID          string `json:"levelId"`
	LevelName        string `json:"levelName"`
	LevelNameZH      string `json:"levelNameZh"`
	SceneName        string `json:"sceneName"`
	ScreenshotCOSKey string `json:"screenshotCosKey,omitempty"`
	ScreenshotURL    string `json:"screenshotUrl,omitempty"`
	SortOrder        int    `json:"sortOrder"`
}

type ParseJob struct {
	ID        int64      `json:"id"`
	VersionID int64      `json:"versionId"`
	Status    JobStatus  `json:"status"`
	TempDir   string     `json:"tempDir,omitempty"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Error     string     `json:"error,omitempty"`
	Phase     string     `json:"phase,omitempty"`
	Progress  int        `json:"progress"`
	Message   string     `json:"message,omitempty"`
}

type PresignResult struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type BundleManifest struct {
	Slug          string              `json:"slug"`
	Version       string              `json:"version"`
	UID           string              `json:"uid"`
	LevelSetName  string              `json:"levelSetName"`
	LevelSetNameZH string             `json:"levelSetNameZH"`
	Author        string              `json:"author"`
	Levels        []BundleLevelEntry  `json:"levels"`
}

type BundleLevelEntry struct {
	ID          string `json:"id"`
	LevelName   string `json:"levelName"`
	LevelNameZH string `json:"levelNameZh"`
	SceneName   string `json:"sceneName"`
	Screenshot  string `json:"screenshot"`
}

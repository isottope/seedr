package seedrcc

import (
	"time"
)

// BaseModel is a common interface for all models to potentially share common methods.
// In Go, embedding structs is often preferred over deep inheritance like Python's dataclasses.
// For now, it's just a conceptual placeholder.

// Torrent represents a torrent in the user's account.
type Torrent struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Size           int        `json:"size"`
	Hash           string     `json:"hash"`
	Progress       string     `json:"progress"`
	LastUpdate     *time.Time `json:"last_update,omitempty"`
	Folder         string     `json:"folder,omitempty"`
	DownloadRate   int        `json:"download_rate,omitempty"`
	UploadRate     int        `json:"upload_rate,omitempty"`
	TorrentQuality *int       `json:"torrent_quality,omitempty"`
	ConnectedTo    int        `json:"connected_to,omitempty"`
	DownloadingFrom int        `json:"downloading_from,omitempty"`
	UploadingTo    int        `json:"uploading_to,omitempty"`
	Seeders        int        `json:"seeders,omitempty"`
	Leechers       int        `json:"leechers,omitempty"`
	Warnings       *string    `json:"warnings,omitempty"`
	Stopped        int        `json:"stopped,omitempty"`
	ProgressURL    *string    `json:"progress_url,omitempty"`
}

// File represents a file within Seedr.
type File struct {
	FileID      int        `json:"file_id"`
	Name        string     `json:"name"`
	Size        int        `json:"size"`
	FolderID    int        `json:"folder_id"`
	FolderFileID int        `json:"folder_file_id"`
	Hash        string     `json:"hash"`
	LastUpdate  *time.Time `json:"last_update,omitempty"`
	PlayAudio   bool       `json:"play_audio,omitempty"`
	PlayVideo   bool       `json:"play_video,omitempty"`
	VideoProgress *string    `json:"video_progress,omitempty"`
	IsLost      int        `json:"is_lost,omitempty"`
	Thumb       *string    `json:"thumb,omitempty"`
}

// Folder represents a folder, which can contain files, torrents, and other folders.
type Folder struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Fullname   string     `json:"fullname"`
	Size       int        `json:"size"`
	LastUpdate *time.Time `json:"last_update,omitempty"`
	IsShared   bool       `json:"is_shared"`
	PlayAudio  bool       `json:"play_audio"`
	PlayVideo  bool       `json:"play_video"`
	Folders    []Folder   `json:"folders,omitempty"`
	Files      []File     `json:"files,omitempty"`
	Torrents   []Torrent  `json:"torrents,omitempty"`
	Parent     *int       `json:"parent,omitempty"`
	Timestamp  *time.Time `json:"timestamp,omitempty"`
	Indexes    []interface{} `json:"indexes,omitempty"` // Python used List[Any]
}

// AccountSettings represents the nested 'settings' object in the user settings response.
type AccountSettings struct {
	AllowRemoteAccess bool   `json:"allow_remote_access"`
	SiteLanguage      string `json:"site_language"`
	SubtitlesLanguage string `json:"subtitles_language"`
	EmailAnnouncements bool   `json:"email_announcements"`
	EmailNewsletter   bool   `json:"email_newsletter"`
}

// AccountInfo represents the nested 'account' object in the user settings response.
type AccountInfo struct {
	Username      string        `json:"username"`
	UserID        int           `json:"user_id"`
	Premium       int           `json:"premium"`
	PackageID     int           `json:"package_id"`
	PackageName   string        `json:"package_name"`
	SpaceUsed     int           `json:"space_used"`
	SpaceMax      int           `json:"space_max"`
	BandwidthUsed int           `json:"bandwidth_used"`
	Email         string        `json:"email"`
	Wishlist      []interface{} `json:"wishlist"` // Python used list
	Invites       int           `json:"invites"`
	InvitesAccepted int           `json:"invites_accepted"`
	MaxInvites    int           `json:"max_invites"`
}

// UserSettings represents the complete response from the get_settings endpoint.
type UserSettings struct {
	Result   bool            `json:"result"`
	Code     int             `json:"code"`
	Settings AccountSettings `json:"settings"`
	Account  AccountInfo     `json:"account"`
	Country  string          `json:"country"`
}

// MemoryBandwidth represents the user's memory and bandwidth usage details.
type MemoryBandwidth struct {
	BandwidthUsed int `json:"bandwidth_used"`
	BandwidthMax  int `json:"bandwidth_max"`
	SpaceUsed     int `json:"space_used"`
	SpaceMax      int `json:"space_max"`
	IsPremium     int `json:"is_premium"`
}

// Device represents a device connected to the user's account.
type Device struct {
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
	DeviceCode string `json:"device_code"`
	TK         string `json:"tk"`
}

// DeviceCode represents the codes used in the device authentication flow.
type DeviceCode struct {
	ExpiresIn     int    `json:"expires_in"`
	Interval      int    `json:"interval"`
	DeviceCode    string `json:"device_code"`
	UserCode      string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
}

// ScannedTorrent represents a torrent found by the scan_page method.
type ScannedTorrent struct {
	ID        int        `json:"id"`
	Hash      string     `json:"hash"`
	Size      int        `json:"size"`
	Title     string     `json:"title"`
	Magnet    string     `json:"magnet"`
	LastUse   *time.Time `json:"last_use,omitempty"`
	Pct       float64    `json:"pct"`
	Filenames []string   `json:"filenames,omitempty"`
	Filesizes []int      `json:"filesizes,omitempty"`
}

// ListContentsResult represents the result of listing folder contents, including account metadata.
// It embeds Folder to inherit its fields.
type ListContentsResult struct {
	Folder
	SpaceUsed     int          `json:"space_used"`
	SpaceMax      int          `json:"space_max"`
	SawWalkthrough int          `json:"saw_walkthrough"`
	Type          string       `json:"type"`
	T             []*time.Time `json:"t,omitempty"` // List of Optional[datetime]
}

// AddTorrentResult represents the result of adding a torrent.
type AddTorrentResult struct {
	Result        bool    `json:"result"`
	UserTorrentID int     `json:"user_torrent_id"`
	Title         string  `json:"title"`
	TorrentHash   string  `json:"torrent_hash"`
	Code          *int    `json:"code,omitempty"`
}

// CreateArchiveResult represents the result of a request to create an archive.
type CreateArchiveResult struct {
	Result     bool   `json:"result"`
	ArchiveID  int    `json:"archive_id"`
	ArchiveURL string `json:"archive_url"`
	Code       *int   `json:"code,omitempty"`
}

// FetchFileResult represents the result of a request to fetch a file, including the download URL.
type FetchFileResult struct {
	Result bool   `json:"result"`
	URL    string `json:"url"`
	Name   string `json:"name"`
}

// RefreshTokenResult represents the response from a token refresh.
type RefreshTokenResult struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   int     `json:"expires_in"`
	TokenType   string  `json:"token_type"`
	Scope       *string `json:"scope,omitempty"`
}

// ScanPageResult represents the full result of a scan_page request.
type ScanPageResult struct {
	Result  bool             `json:"result"`
	Torrents []ScannedTorrent `json:"torrents"`
}

// APIResult represents a generic API result for operations that return a simple success/failure.
type APIResult struct {
	Result bool `json:"result"`
	Code   *int `json:"code,omitempty"`
}

// Helper functions for FromMap (equivalent to Python's from_dict)
// These are not direct translations of Python's BaseModel.from_dict but manual unmarshaling.

func NewTorrentFromMap(data map[string]interface{}) Torrent {
	t := Torrent{}
	if id, ok := data["id"].(float64); ok {
		t.ID = int(id)
	}
	if name, ok := data["name"].(string); ok {
		t.Name = name
	}
	if size, ok := data["size"].(float64); ok {
		t.Size = int(size)
	}
	if hash, ok := data["hash"].(string); ok {
		t.Hash = hash
	}
	if progress, ok := data["progress"].(string); ok {
		t.Progress = progress
	}
	t.LastUpdate = ParseDateTime(data["last_update"])
	if folder, ok := data["folder"].(string); ok {
		t.Folder = folder
	}
	if dr, ok := data["download_rate"].(float64); ok {
		t.DownloadRate = int(dr)
	}
	if ur, ok := data["upload_rate"].(float64); ok {
		t.UploadRate = int(ur)
	}
	if tq, ok := data["torrent_quality"].(float64); ok {
		val := int(tq)
		t.TorrentQuality = &val
	}
	if ct, ok := data["connected_to"].(float64); ok {
		t.ConnectedTo = int(ct)
	}
	if df, ok := data["downloading_from"].(float64); ok {
		t.DownloadingFrom = int(df)
	}
	if ut, ok := data["uploading_to"].(float64); ok {
		t.UploadingTo = int(ut)
	}
	if seeders, ok := data["seeders"].(float64); ok {
		t.Seeders = int(seeders)
	}
	if leechers, ok := data["leechers"].(float64); ok {
		t.Leechers = int(leechers)
	}
	if warnings, ok := data["warnings"].(string); ok {
		t.Warnings = &warnings
	}
	if stopped, ok := data["stopped"].(float64); ok {
		t.Stopped = int(stopped)
	}
	if pu, ok := data["progress_url"].(string); ok {
		t.ProgressURL = &pu
	}
	return t
}

func NewFileFromMap(data map[string]interface{}) File {
	f := File{}
	if fileID, ok := data["file_id"].(float64); ok {
		f.FileID = int(fileID)
	}
	if name, ok := data["name"].(string); ok {
		f.Name = name
	}
	if size, ok := data["size"].(float64); ok {
		f.Size = int(size)
	}
	if folderID, ok := data["folder_id"].(float64); ok {
		f.FolderID = int(folderID)
	}
	if folderFileID, ok := data["folder_file_id"].(float64); ok {
		f.FolderFileID = int(folderFileID)
	}
	if hash, ok := data["hash"].(string); ok {
		f.Hash = hash
	}
	f.LastUpdate = ParseDateTime(data["last_update"])
	if playAudio, ok := data["play_audio"].(bool); ok {
		f.PlayAudio = playAudio
	}
	if playVideo, ok := data["play_video"].(bool); ok {
		f.PlayVideo = playVideo
	}
	if vp, ok := data["video_progress"].(string); ok {
		f.VideoProgress = &vp
	}
	if isLost, ok := data["is_lost"].(float64); ok {
		f.IsLost = int(isLost)
	}
	if thumb, ok := data["thumb"].(string); ok {
		f.Thumb = &thumb
	}
	return f
}

func NewFolderFromMap(data map[string]interface{}) Folder {
	f := Folder{}
	// Handle multiple possible keys for ID
	if id, ok := data["id"].(float64); ok {
		f.ID = int(id)
	} else if folderID, ok := data["folder_id"].(float64); ok {
		f.ID = int(folderID)
	}

	if name, ok := data["name"].(string); ok {
		f.Name = name
	}
	if fullname, ok := data["fullname"].(string); ok {
		f.Fullname = fullname
	} else if name, ok := data["name"].(string); ok {
		f.Fullname = name // Fallback to name if fullname is missing
	}
	if size, ok := data["size"].(float64); ok {
		f.Size = int(size)
	}
	if lastUpdate := ParseDateTime(data["last_update"]); lastUpdate != nil {
		f.LastUpdate = lastUpdate
	} else if timestamp := ParseDateTime(data["timestamp"]); timestamp != nil {
		f.LastUpdate = timestamp // Fallback to timestamp
	}

	if isShared, ok := data["is_shared"].(bool); ok {
		f.IsShared = isShared
	}
	if playAudio, ok := data["play_audio"].(bool); ok {
		f.PlayAudio = playAudio
	}
	if playVideo, ok := data["play_video"].(bool); ok {
		f.PlayVideo = playVideo
	}

	if foldersData, ok := data["folders"].([]interface{}); ok {
		for _, fd := range foldersData {
			if folderMap, isMap := fd.(map[string]interface{}); isMap {
				f.Folders = append(f.Folders, NewFolderFromMap(folderMap))
			}
		}
	}
	if filesData, ok := data["files"].([]interface{}); ok {
		for _, fData := range filesData {
			if fileMap, isMap := fData.(map[string]interface{}); isMap {
				f.Files = append(f.Files, NewFileFromMap(fileMap))
			}
		}
	}
	if torrentsData, ok := data["torrents"].([]interface{}); ok {
		for _, tData := range torrentsData {
			if torrentMap, isMap := tData.(map[string]interface{}); isMap {
				f.Torrents = append(f.Torrents, NewTorrentFromMap(torrentMap))
			}
		}
	}
	if parent, ok := data["parent"].(float64); ok {
		val := int(parent)
		f.Parent = &val
	}
	f.Timestamp = ParseDateTime(data["timestamp"])
	if indexes, ok := data["indexes"].([]interface{}); ok {
		f.Indexes = indexes
	}
	return f
}

func NewAccountSettingsFromMap(data map[string]interface{}) AccountSettings {
	as := AccountSettings{}
	if v, ok := data["allow_remote_access"].(bool); ok {
		as.AllowRemoteAccess = v
	}
	if v, ok := data["site_language"].(string); ok {
		as.SiteLanguage = v
	}
	if v, ok := data["subtitles_language"].(string); ok {
		as.SubtitlesLanguage = v
	}
	if v, ok := data["email_announcements"].(bool); ok {
		as.EmailAnnouncements = v
	}
	if v, ok := data["email_newsletter"].(bool); ok {
		as.EmailNewsletter = v
	}
	return as
}

func NewAccountInfoFromMap(data map[string]interface{}) AccountInfo {
	ai := AccountInfo{}
	if v, ok := data["username"].(string); ok {
		ai.Username = v
	}
	if v, ok := data["user_id"].(float64); ok {
		ai.UserID = int(v)
	}
	if v, ok := data["premium"].(float64); ok {
		ai.Premium = int(v)
	}
	if v, ok := data["package_id"].(float64); ok {
		ai.PackageID = int(v)
	}
	if v, ok := data["package_name"].(string); ok {
		ai.PackageName = v
	}
	if v, ok := data["space_used"].(float64); ok {
		ai.SpaceUsed = int(v)
	}
	if v, ok := data["space_max"].(float64); ok {
		ai.SpaceMax = int(v)
	}
	if v, ok := data["bandwidth_used"].(float64); ok {
		ai.BandwidthUsed = int(v)
	}
	if v, ok := data["email"].(string); ok {
		ai.Email = v
	}
	if v, ok := data["wishlist"].([]interface{}); ok {
		ai.Wishlist = v
	}
	if v, ok := data["invites"].(float64); ok {
		ai.Invites = int(v)
	}
	if v, ok := data["invites_accepted"].(float64); ok {
		ai.InvitesAccepted = int(v)
	}
	if v, ok := data["max_invites"].(float64); ok {
		ai.MaxInvites = int(v)
	}
	return ai
}

func NewUserSettingsFromMap(data map[string]interface{}) UserSettings {
	us := UserSettings{}
	if v, ok := data["result"].(bool); ok {
		us.Result = v
	}
	if v, ok := data["code"].(float64); ok {
		us.Code = int(v)
	}
	if v, ok := data["settings"].(map[string]interface{}); ok {
		us.Settings = NewAccountSettingsFromMap(v)
	}
	if v, ok := data["account"].(map[string]interface{}); ok {
		us.Account = NewAccountInfoFromMap(v)
	}
	if v, ok := data["country"].(string); ok {
		us.Country = v
	}
	return us
}

func NewMemoryBandwidthFromMap(data map[string]interface{}) MemoryBandwidth {
	mb := MemoryBandwidth{}
	if v, ok := data["bandwidth_used"].(float64); ok {
		mb.BandwidthUsed = int(v)
	}
	if v, ok := data["bandwidth_max"].(float64); ok {
		mb.BandwidthMax = int(v)
	}
	if v, ok := data["space_used"].(float64); ok {
		mb.SpaceUsed = int(v)
	}
	if v, ok := data["space_max"].(float64); ok {
		mb.SpaceMax = int(v)
	}
	if v, ok := data["is_premium"].(float64); ok {
		mb.IsPremium = int(v)
	}
	return mb
}

func NewDeviceFromMap(data map[string]interface{}) Device {
	d := Device{}
	if v, ok := data["client_id"].(string); ok {
		d.ClientID = v
	}
	if v, ok := data["client_name"].(string); ok {
		d.ClientName = v
	}
	if v, ok := data["device_code"].(string); ok {
		d.DeviceCode = v
	}
	if v, ok := data["tk"].(string); ok {
		d.TK = v
	}
	return d
}

func NewDeviceCodeFromMap(data map[string]interface{}) DeviceCode {
	dc := DeviceCode{}
	if v, ok := data["expires_in"].(float64); ok {
		dc.ExpiresIn = int(v)
	}
	if v, ok := data["interval"].(float64); ok {
		dc.Interval = int(v)
	}
	if v, ok := data["device_code"].(string); ok {
		dc.DeviceCode = v
	}
	if v, ok := data["user_code"].(string); ok {
		dc.UserCode = v
	}
	if v, ok := data["verification_url"].(string); ok {
		dc.VerificationURL = v
	}
	return dc
}

func NewScannedTorrentFromMap(data map[string]interface{}) ScannedTorrent {
	st := ScannedTorrent{}
	if v, ok := data["id"].(float64); ok {
		st.ID = int(v)
	}
	if v, ok := data["hash"].(string); ok {
		st.Hash = v
	}
	if v, ok := data["size"].(float64); ok {
		st.Size = int(v)
	}
	if v, ok := data["title"].(string); ok {
		st.Title = v
	}
	if v, ok := data["magnet"].(string); ok {
		st.Magnet = v
	}
	st.LastUse = ParseDateTime(data["last_use"])
	if v, ok := data["pct"].(float64); ok {
		st.Pct = v
	}
	if v, ok := data["filenames"].([]interface{}); ok {
		for _, item := range v {
			if s, isString := item.(string); isString {
				st.Filenames = append(st.Filenames, s)
			}
		}
	}
	if v, ok := data["filesizes"].([]interface{}); ok {
		for _, item := range v {
			if i, isFloat := item.(float64); isFloat {
				st.Filesizes = append(st.Filesizes, int(i))
			}
		}
	}
	return st
}

func NewListContentsResultFromMap(data map[string]interface{}) ListContentsResult {
	lcr := ListContentsResult{
		Folder: NewFolderFromMap(data), // Embed and initialize Folder part
	}
	if v, ok := data["space_used"].(float64); ok {
		lcr.SpaceUsed = int(v)
	}
	if v, ok := data["space_max"].(float64); ok {
		lcr.SpaceMax = int(v)
	}
	if v, ok := data["saw_walkthrough"].(float64); ok {
		lcr.SawWalkthrough = int(v)
	}
	if v, ok := data["type"].(string); ok {
		lcr.Type = v
	}
	if tList, ok := data["t"].([]interface{}); ok {
		for _, item := range tList {
			lcr.T = append(lcr.T, ParseDateTime(item))
		}
	}
	return lcr
}

func NewAddTorrentResultFromMap(data map[string]interface{}) AddTorrentResult {
	atr := AddTorrentResult{}
	if v, ok := data["result"].(bool); ok {
		atr.Result = v
	}
	if v, ok := data["user_torrent_id"].(float64); ok {
		atr.UserTorrentID = int(v)
	}
	if v, ok := data["title"].(string); ok {
		atr.Title = v
	}
	if v, ok := data["torrent_hash"].(string); ok {
		atr.TorrentHash = v
	}
	if v, ok := data["code"].(float64); ok {
		val := int(v)
		atr.Code = &val
	}
	return atr
}

func NewCreateArchiveResultFromMap(data map[string]interface{}) CreateArchiveResult {
	car := CreateArchiveResult{}
	if v, ok := data["result"].(bool); ok {
		car.Result = v
	}
	if v, ok := data["archive_id"].(float64); ok {
		car.ArchiveID = int(v)
	}
	if v, ok := data["archive_url"].(string); ok {
		car.ArchiveURL = v
	}
	if v, ok := data["code"].(float64); ok {
		val := int(v)
		car.Code = &val
	}
	return car
}

func NewFetchFileResultFromMap(data map[string]interface{}) FetchFileResult {
	ffr := FetchFileResult{}
	if v, ok := data["result"].(bool); ok {
		ffr.Result = v
	}
	if v, ok := data["url"].(string); ok {
		ffr.URL = v
	}
	if v, ok := data["name"].(string); ok {
		ffr.Name = v
	}
	return ffr
}

func NewRefreshTokenResultFromMap(data map[string]interface{}) RefreshTokenResult {
	rtr := RefreshTokenResult{}
	if v, ok := data["access_token"].(string); ok {
		rtr.AccessToken = v
	}
	if v, ok := data["expires_in"].(float64); ok {
		rtr.ExpiresIn = int(v)
	}
	if v, ok := data["token_type"].(string); ok {
		rtr.TokenType = v
	}
	if v, ok := data["scope"].(string); ok {
		rtr.Scope = &v
	}
	return rtr
}

func NewScanPageResultFromMap(data map[string]interface{}) ScanPageResult {
	spr := ScanPageResult{}
	if v, ok := data["result"].(bool); ok {
		spr.Result = v
	}
	if v, ok := data["torrents"].([]interface{}); ok {
		for _, item := range v {
			if torrentMap, isMap := item.(map[string]interface{}); isMap {
				spr.Torrents = append(spr.Torrents, NewScannedTorrentFromMap(torrentMap))
			}
		}
	}
	return spr
}

func NewAPIResultFromMap(data map[string]interface{}) APIResult {
	ar := APIResult{}
	if v, ok := data["result"].(bool); ok {
		ar.Result = v
	}
	if v, ok := data["code"].(float64); ok {
		val := int(v)
		ar.Code = &val
	}
	return ar
}

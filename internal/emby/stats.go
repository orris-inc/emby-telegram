package emby

import "time"

type SessionInfo struct {
	ID                   string              `json:"Id"`
	UserID               string              `json:"UserId"`
	UserName             string              `json:"UserName"`
	DeviceID             string              `json:"DeviceId"`
	DeviceName           string              `json:"DeviceName"`
	Client               string              `json:"Client"`
	ApplicationVersion   string              `json:"ApplicationVersion"`
	RemoteEndPoint       string              `json:"RemoteEndPoint"`
	NowPlayingItem       *NowPlayingItem     `json:"NowPlayingItem"`
	PlayState            *PlayState          `json:"PlayState"`
	LastActivityDate     time.Time           `json:"LastActivityDate"`
	SupportsRemoteControl bool               `json:"SupportsRemoteControl"`
	TranscodingInfo      *TranscodingInfo    `json:"TranscodingInfo"`
}

type NowPlayingItem struct {
	ID                string  `json:"Id"`
	Name              string  `json:"Name"`
	Type              string  `json:"Type"`
	MediaType         string  `json:"MediaType"`
	RunTimeTicks      int64   `json:"RunTimeTicks"`
	SeriesName        string  `json:"SeriesName"`
	SeasonName        string  `json:"SeasonName"`
	IndexNumber       int     `json:"IndexNumber"`
	ParentIndexNumber int     `json:"ParentIndexNumber"`
	ProductionYear    int     `json:"ProductionYear"`
}

type PlayState struct {
	PositionTicks      int64  `json:"PositionTicks"`
	CanSeek            bool   `json:"CanSeek"`
	IsPaused           bool   `json:"IsPaused"`
	IsMuted            bool   `json:"IsMuted"`
	VolumeLevel        int    `json:"VolumeLevel"`
	PlayMethod         string `json:"PlayMethod"`
	RepeatMode         string `json:"RepeatMode"`
}

type TranscodingInfo struct {
	IsVideoDirect      bool    `json:"IsVideoDirect"`
	IsAudioDirect      bool    `json:"IsAudioDirect"`
	VideoCodec         string  `json:"VideoCodec"`
	AudioCodec         string  `json:"AudioCodec"`
	Container          string  `json:"Container"`
	Bitrate            int     `json:"Bitrate"`
	Framerate          float64 `json:"Framerate"`
	CompletionPercentage float64 `json:"CompletionPercentage"`
	TranscodeReasons   []string `json:"TranscodeReasons"`
}

func (s *SessionInfo) IsPlaying() bool {
	return s.NowPlayingItem != nil && s.PlayState != nil && !s.PlayState.IsPaused
}

func (s *SessionInfo) GetProgress() float64 {
	if s.NowPlayingItem == nil || s.PlayState == nil {
		return 0
	}
	if s.NowPlayingItem.RunTimeTicks == 0 {
		return 0
	}
	return float64(s.PlayState.PositionTicks) / float64(s.NowPlayingItem.RunTimeTicks) * 100
}

func (n *NowPlayingItem) GetDisplayName() string {
	if n.Type == "Episode" && n.SeriesName != "" {
		if n.ParentIndexNumber > 0 && n.IndexNumber > 0 {
			return n.SeriesName + " S" + formatNumber(n.ParentIndexNumber) + "E" + formatNumber(n.IndexNumber)
		}
		return n.SeriesName + " - " + n.Name
	}
	return n.Name
}

func (n *NowPlayingItem) GetDuration() time.Duration {
	return time.Duration(n.RunTimeTicks * 100)
}

func formatNumber(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0' + n/10)) + string(rune('0' + n%10))
}

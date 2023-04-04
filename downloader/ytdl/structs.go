package ytdl

type ytdlJSON struct {
	Id                 string      `json:"id"`
	Title              string      `json:"title"`
	Formats            []format    `json:"formats"`
	Thumbnails         []thumbnail `json:"thumbnails"`
	Description        string      `json:"description"`
	UploadDate         string      `json:"upload_date"`
	Uploader           string      `json:"uploader"`
	ChannelId          string      `json:"channel_id"`
	Duration           int         `json:"duration"`
	ViewCount          int         `json:"view_count"`
	AgeLimit           int         `json:"age_limit"`
	WebpageUrl         string      `json:"webpage_url"`
	Categories         []string    `json:"categories"`
	Tags               []string    `json:"tags"`
	LikeCount          int         `json:"like_count"`
	Track              string      `json:"track"`
	Artist             string      `json:"artist"`
	License            string      `json:"license"`
	Creator            string      `json:"creator"`
	AltTitle           string      `json:"alt_title"`
	UploaderId         string      `json:"uploader_id"`
	UploaderUrl        string      `json:"uploader_url"`
	ChannelUrl         string      `json:"channel_url"`
	Channel            string      `json:"channel"`
	Extractor          string      `json:"extractor"`
	WebpageUrlBasename string      `json:"webpage_url_basename"`
	ExtractorKey       string      `json:"extractor_key"`
	Playlist           interface{} `json:"playlist"`
	PlaylistIndex      interface{} `json:"playlist_index"`
	Thumbnail          string      `json:"thumbnail"`
	DisplayId          string      `json:"display_id"`
	RequestedSubtitles interface{} `json:"requested_subtitles"`
	RequestedFormats   []format    `json:"requested_formats"`
	Format             string      `json:"format"`
	FormatId           string      `json:"format_id"`
	Width              int         `json:"width"`
	Height             int         `json:"height"`
	Resolution         interface{} `json:"resolution"`
	Fps                int         `json:"fps"`
	Vcodec             string      `json:"vcodec"`
	Vbr                float64     `json:"vbr"`
	StretchedRatio     interface{} `json:"stretched_ratio"`
	Acodec             string      `json:"acodec"`
	Abr                float64     `json:"abr"`
	Ext                string      `json:"ext"`
	Fulltitle          string      `json:"fulltitle"`
	Filename           string      `json:"_filename"`
}

type format struct {
	Asr        *int    `json:"asr"`
	Filesize   int64   `json:"filesize"`
	FormatId   string  `json:"format_id"`
	FormatNote string  `json:"format_note"`
	Fps        *int    `json:"fps"`
	Height     *int    `json:"height"`
	Quality    int     `json:"quality"`
	Tbr        float64 `json:"tbr"`
	Url        string  `json:"url"`
	Width      *int    `json:"width"`
	Ext        string  `json:"ext"`
	Vcodec     string  `json:"vcodec"`
	Acodec     string  `json:"acodec"`
	Abr        float64 `json:"abr,omitempty"`
	Protocol   string  `json:"protocol"`
	Fragments  []struct {
		Url string `json:"url"`
	} `json:"fragments,omitempty"`
	Container   string `json:"container,omitempty"`
	Format      string `json:"format"`
	HttpHeaders struct {
		UserAgent      string `json:"User-Agent"`
		AcceptCharset  string `json:"Accept-Charset"`
		Accept         string `json:"Accept"`
		AcceptEncoding string `json:"Accept-Encoding"`
		AcceptLanguage string `json:"Accept-Language"`
	} `json:"http_headers"`
	Vbr float64 `json:"vbr,omitempty"`
}

type thumbnail struct {
	Height     int    `json:"height"`
	Url        string `json:"url"`
	Width      int    `json:"width"`
	Resolution string `json:"resolution"`
	Id         string `json:"id"`
}

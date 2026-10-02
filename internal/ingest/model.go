package ingest

type FFProbeOutput struct {
	Streams []struct{
			Width     int     `json:"width"`
			Height    int     `json:"height"`
			CodecName string  `json:"codec_name"`
	} `json:"streams"`
	Format struct {
			Duration  string  `json:"duration"`
			Size      string  `json:"size"`
	} `json:"format"`
}

type VideoMetadata struct {
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	Duration  float64  `json:"duration"`
	Codec     string   `json:"codec"`
}
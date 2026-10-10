package transcode

import (
	"context"
	"os/exec"
	"strings"
)

func HasAudioTrack(ctx context.Context, inputPath string) bool {
	cmd := exec.CommandContext(ctx, "ffprobe",
					"-v", "error",
					"-select_streams", "a",
					"-show_entries", "stream=codec_type",
					"-of", "csv=p=0",
					inputPath,
	)

	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) != ""
}
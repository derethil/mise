package video

import (
	"context"

	"github.com/lrstanley/go-ytdlp"
)

const (
	audioFormat         = "mp3"
	audioSourceFormat   = "bestaudio/best"
	audioOutputTemplate = "audio.%(ext)s"
)

func (e *Extraction) extractAudio(ctx context.Context) (string, error) {
	cmd := audioCommand(e.baseCmd, audioFormat)

	if err := e.run(ctx, cmd); err != nil {
		return "", err
	}

	return locateDownload(e.workdir)
}

func audioCommand(cmd *ytdlp.Command, format string) *ytdlp.Command {
	return cmd.Clone().
		Format(audioSourceFormat).
		ExtractAudio().
		AudioFormat(format).
		Output(audioOutputTemplate)
}

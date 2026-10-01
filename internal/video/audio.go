package video

import (
	"context"
	"fmt"

	cwhisper "github.com/ggerganov/whisper.cpp/bindings/go"
	"github.com/lrstanley/go-ytdlp"
)

const (
	audioFormat         = "wav"
	audioSourceFormat   = "bestaudio/best"
	audioOutputTemplate = "audio.%(ext)s"
	audioChannels       = 1
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
		PostProcessorArgs(fmt.Sprintf("ExtractAudio+ffmpeg:-ar %d -ac %d", cwhisper.SampleRate, audioChannels)).
		Output(audioOutputTemplate)
}

package video

import (
	"encoding/json"
	"encoding/json/jsontext"
	"testing"
	"time"

	"github.com/lrstanley/go-ytdlp"
	"github.com/stretchr/testify/suite"
)

func ptr[T any](v T) *T { return &v }

type ExtractSuite struct {
	suite.Suite
}

func TestExtractSuite(t *testing.T) {
	suite.Run(t, new(ExtractSuite))
}

func (s *ExtractSuite) jsonValue(raw string) *jsontext.Value {
	s.T().Helper()

	value := jsontext.Value(raw)
	return &value
}

func (s *ExtractSuite) TestSingleVideo() {
	video := &ytdlp.ExtractedInfo{Type: ytdlp.ExtractedTypeVideo, ID: "abc"}

	tests := []struct {
		name  string
		infos []*ytdlp.ExtractedInfo
		want  error
	}{
		{"no results", nil, ErrNoVideo},
		{"empty slice", []*ytdlp.ExtractedInfo{}, ErrNoVideo},
		{"playlist wrapper", []*ytdlp.ExtractedInfo{{Type: ytdlp.ExtractedTypePlaylist}}, ErrPlaylistURL},
		{"multi_video wrapper", []*ytdlp.ExtractedInfo{{Type: ytdlp.ExtractedTypeMultiVideo}}, ErrPlaylistURL},
		{"flat playlist url stub", []*ytdlp.ExtractedInfo{{Type: ytdlp.ExtractedTypeURL}}, ErrPlaylistURL},
		{"several videos", []*ytdlp.ExtractedInfo{video, video}, ErrPlaylistURL},
		{"single video", []*ytdlp.ExtractedInfo{video}, nil},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			info, err := singleVideo(tt.infos)

			if tt.want != nil {
				s.ErrorIs(err, tt.want)
				return
			}

			s.Require().NoError(err)
			s.Equal("abc", info.ID)
		})
	}
}

func (s *ExtractSuite) TestCheckDuration() {
	tests := []struct {
		name  string
		info  *ytdlp.ExtractedInfo
		limit int
		want  error
	}{
		{"live stream is refused even without a limit", &ytdlp.ExtractedInfo{IsLive: ptr(true)}, 0, ErrLiveVideo},
		{"live stream with a limit", &ytdlp.ExtractedInfo{IsLive: ptr(true), Duration: ptr(60.0)}, 20, ErrLiveVideo},
		{"missing duration with a limit", &ytdlp.ExtractedInfo{}, 20, ErrDurationUnknown},
		{"missing duration without a limit", &ytdlp.ExtractedInfo{}, 0, nil},
		{"under the limit", &ytdlp.ExtractedInfo{Duration: ptr(600.0)}, 20, nil},
		{"exactly at the limit", &ytdlp.ExtractedInfo{Duration: ptr(1200.0)}, 20, nil},
		{"over the limit", &ytdlp.ExtractedInfo{Duration: ptr(1800.0)}, 20, ErrVideoTooLong},
		{"long video with the limit disabled", &ytdlp.ExtractedInfo{Duration: ptr(99999.0)}, 0, nil},
		{"zero duration is reported, not missing", &ytdlp.ExtractedInfo{Duration: ptr(0.0)}, 20, nil},
		{"explicitly not live", &ytdlp.ExtractedInfo{IsLive: ptr(false), Duration: ptr(60.0)}, 20, nil},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := checkDuration(tt.info, tt.limit)

			if tt.want == nil {
				s.NoError(err)
				return
			}

			s.ErrorIs(err, tt.want)
		})
	}
}

func (s *ExtractSuite) TestCheckDurationReportsBothDurations() {
	err := checkDuration(&ytdlp.ExtractedInfo{Duration: ptr(1800.0)}, 20)

	var tooLong *TooLongError
	s.Require().ErrorAs(err, &tooLong)
	s.Equal(30*time.Minute, tooLong.Duration)
	s.Equal(20*time.Minute, tooLong.Limit)
}

func (s *ExtractSuite) TestFlattenSourceHandlesMissingFields() {
	source := flattenSource(&ytdlp.ExtractedInfo{ID: "abc"}, "https://tiktok.com/typed")

	s.Equal("abc", source.ID)
	s.Empty(source.Title)
	s.Empty(source.Uploader)
	s.Zero(source.Duration)
	s.Equal("https://tiktok.com/typed", source.URL, "falls back to the requested url")
}

func (s *ExtractSuite) TestFlattenSourcePrefersCanonicalURL() {
	info := &ytdlp.ExtractedInfo{
		ID:          "abc",
		Title:       ptr("Tacos"),
		Description: ptr("a description"),
		Uploader:    ptr("chef"),
		Duration:    ptr(90.5),
		Thumbnail:   ptr("https://img.example/t.jpg"),
		Extractor:   ptr("tiktok"),
		WebpageURL:  ptr("https://tiktok.com/canonical"),
	}

	source := flattenSource(info, "https://tiktok.com/typed")

	s.Equal("https://tiktok.com/canonical", source.URL)
	s.Equal("Tacos", source.Title)
	s.Equal("chef", source.Uploader)
	s.Equal(90500*time.Millisecond, source.Duration)
	s.Equal("tiktok", source.Extractor)
}

func (s *ExtractSuite) TestInfoDocumentPrefersRawJSON() {
	raw := s.jsonValue(`{"id":"abc","extractor_specific":"KEEPME"}`)
	res := &ytdlp.Result{OutputLogs: []*ytdlp.ResultLog{{JSON: raw}}}

	document, err := infoDocument(res, &ytdlp.ExtractedInfo{ID: "abc"})

	s.Require().NoError(err)
	s.Contains(string(document), "extractor_specific", "fields go-ytdlp does not model must survive")
}

func (s *ExtractSuite) TestInfoDocumentFallsBackToMarshalling() {
	res := &ytdlp.Result{OutputLogs: []*ytdlp.ResultLog{{Line: "not json"}}}

	document, err := infoDocument(res, &ytdlp.ExtractedInfo{ID: "abc"})

	s.Require().NoError(err)

	var decoded map[string]any
	s.Require().NoError(json.Unmarshal(document, &decoded))
	s.Equal("abc", decoded["id"])
}

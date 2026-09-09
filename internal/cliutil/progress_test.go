package cliutil

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProgressSuite struct {
	suite.Suite
}

func TestProgressSuite(t *testing.T) {
	suite.Run(t, new(ProgressSuite))
}

func (s *ProgressSuite) TestPrintProgressWithTotal() {
	onProgress := PrintProgress()

	err := onProgress(Progress{Label: "qwen2.5", Status: "pulling", Total: 200, Completed: 100})

	s.Require().NoError(err)
}

func (s *ProgressSuite) TestPrintProgressWithoutTotal() {
	onProgress := PrintProgress()

	err := onProgress(Progress{Label: "qwen2.5", Status: "verifying"})

	s.Require().NoError(err)
}

func (s *ProgressSuite) TestPrintProgressTracksStatusChanges() {
	onProgress := PrintProgress()

	s.Require().NoError(onProgress(Progress{Label: "qwen2.5", Status: "pulling", Total: 10, Completed: 1}))
	s.Require().NoError(onProgress(Progress{Label: "qwen2.5", Status: "pulling", Total: 10, Completed: 5}))
	s.Require().NoError(onProgress(Progress{Label: "qwen2.5", Status: "verifying"}))
}

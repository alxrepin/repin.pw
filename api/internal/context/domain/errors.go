package domain

import "errors"

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrChannelNotFound = errors.New("channel not found")
	ErrMediaNotFound   = errors.New("media not found")
	ErrNoJobs          = errors.New("no runnable jobs")
)

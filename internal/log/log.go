package log

import (
	"go.uber.org/zap"
)

// Logger has zap's Logger
var Logger *zap.Logger
// Sugar has zap's SugaredLogger
var Sugar *zap.SugaredLogger

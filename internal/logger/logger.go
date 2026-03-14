package logger

import "go.uber.org/zap"

var Log *zap.Logger

func Init(nodeID string) {
	base, _ := zap.NewProduction()
	Log = base.With(zap.String("node_id", nodeID))
}

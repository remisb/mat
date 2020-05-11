package log

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

type Log struct {
}

//func init() {
//	config := zap.NewDevelopmentConfig()
//	logger, err := config.Build() // NewExample, or NewProduction, or NewDevelopment
//	if err != nil {
//		err := fmt.Errorf("error in log init err: %v", err)
//		fmt.Print(err)
//	}
//	Sugar = logger.Sugar()
//	config.Level.SetLevel(zapcore.DebugLevel)
//}

func SetupLogger() {

}

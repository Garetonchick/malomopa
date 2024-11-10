package config

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Level       string   `json:"level"`
	Encoding    string   `json:"encoding"`
	OutputPaths []string `json:"output_paths"`
}

func MakeLogger(loggerCfg *LoggerConfig) (*zap.Logger, error) {
	if loggerCfg == nil {
		return nil, errors.New("no logger config supplied")
	}

	level, err := zap.ParseAtomicLevel(loggerCfg.Level)
	if err != nil {
		return nil, errors.New("failed to parse logging level")
	}

	zapCfg := zap.Config{
		Level:            level,
		Encoding:         loggerCfg.Encoding,
		OutputPaths:      loggerCfg.OutputPaths,
		ErrorOutputPaths: loggerCfg.OutputPaths,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "level",
			TimeKey:        "ts",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
		},
	}
	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

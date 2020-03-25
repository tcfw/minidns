package main

import "github.com/spf13/viper"

const (
	logLevelVerbose     = "verbose"
	logLevelVeryVerbose = "vverbose"
)

func shouldLogVerbose() bool {
	logLevel := viper.GetString("log_level")
	return logLevel == logLevelVerbose || logLevel == logLevelVeryVerbose
}

func shouldLogVeryVerbose() bool {
	logLevel := viper.GetString("log_level")
	return logLevel == logLevelVeryVerbose
}

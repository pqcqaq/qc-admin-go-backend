package main

import "github.com/spf13/viper"

func injectGlobalVars() {
	viper.Set("application.name", "qc-admin-socket")
}

func main() {
	injectGlobalVars()
	Execute()
}

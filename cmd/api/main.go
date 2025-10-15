// @title           Go Backend API
// @version         1.0
// @description     这是一个基于Go和Gin框架的后端API服务
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
package main

import "github.com/spf13/viper"

func injectGlobalVars() {
	viper.Set("application.name", "qc-admin-api")
}

func main() {
	injectGlobalVars()
	Execute()
}

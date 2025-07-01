package main

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/infra"
	"github.com/mobiletoly/gokatana/katapp"
)

//go:generate go tool oapi-codegen -config swagger/cfg-common.yaml swagger/common.yaml
//go:generate go tool oapi-codegen -config swagger/cfg-auth.yaml swagger/auth.yaml
//go:generate go tool oapi-codegen -config swagger/cfg-tenant.yaml swagger/tenant.yaml
//go:generate go tool oapi-codegen -config swagger/cfg-user.yaml swagger/user.yaml

//go:generate go tool gobetter -input=./internal/core/swagger/auth.gen.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/common.gen.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/tenant.gen.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/user.gen.go -generate-for=exported -receiver=pointer

//go:generate go tool templ generate

func main() {
	katapp.CmdlineExecute(
		"iamservice",
		"IAMService API server",
		"IAMService API server",
		&katapp.CmdlineHandler{
			Run: func(deployment string) {
				infra.Start(deployment, nil, nil)
			},
		})
}

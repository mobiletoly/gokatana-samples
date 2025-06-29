package main

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/infra"
	"github.com/mobiletoly/gokatana/katapp"
)

//go:generate go tool swagger mixin swagger/common.yaml swagger/auth.yaml swagger/tenant.yaml swagger/user.yaml --format=yaml --output=swagger/merged.yaml
//go:generate go tool swagger generate model --spec=swagger/merged.yaml --target=internal/core --model-package=swagger --keep-spec-order

//go:generate go tool gobetter -input=./internal/core/swagger/auth_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/email_confirmation_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/refresh_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/signin_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/signup_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/signup_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/auth_user_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/user_profile.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/user_profile_update_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/assign_role_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/pagination_info.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/user_list_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/user_roles_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/tenant_create_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/tenant_update_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/tenant_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/tenant_list_response.go -generate-for=exported -receiver=pointer

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

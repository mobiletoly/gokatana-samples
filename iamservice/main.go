package main

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/infra"
	"github.com/mobiletoly/gokatana/katapp"
)

//go:generate go tool swagger generate model --spec=swagger/contact.yaml --target=internal/core --model-package=swagger --keep-spec-order
//go:generate go tool swagger generate model --spec=swagger/auth.yaml --target=internal/core --model-package=swagger --keep-spec-order
//go:generate go tool swagger generate model --spec=swagger/user.yaml --target=internal/core --model-package=swagger --keep-spec-order
//go:generate go tool gobetter -input=./internal/core/swagger/add_contact.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/contact.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/auth_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/message_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/refresh_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/signin_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/signup_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/user_profile.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/assign_role_request.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/pagination_info.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/user_list_response.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/swagger/user_roles_response.go -generate-for=exported -receiver=pointer

func main() {
	katapp.CmdlineExecute(
		"hexagonal",
		"IAM Service API server",
		"IAM Service API server",
		&katapp.CmdlineHandler{
			Run: func(deployment string) {
				infra.Start(deployment, nil, nil)
			},
		})
}

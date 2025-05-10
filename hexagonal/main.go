package main

import (
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/infra"
	"github.com/mobiletoly/gokatana/katapp"
)

//go:generate go tool swagger generate model --spec=swagger/contact.yaml --target=internal/core --model-package=model --keep-spec-order
//go:generate go tool gobetter -input=./internal/core/model/add_contact.go -generate-for=exported -receiver=pointer
//go:generate go tool gobetter -input=./internal/core/model/contact.go -generate-for=exported -receiver=pointer

func main() {
	katapp.CmdlineExecute(
		"hexagonal",
		"GoKatana/Hexagonal Echo API server",
		"GoKatana/Hexagonal API server - sample GoKatana service based on Hexagonal architecture with Echo",
		&katapp.CmdlineHandler{
			Run: func(deployment string) {
				infra.Start(deployment, nil, nil)
			},
		})
}

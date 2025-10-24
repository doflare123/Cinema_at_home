package main

import (
	"cinema/internal/app/utils"
	"cinema/internal/container"
)

func main() {
	utils.RegisterPasswordValidator()
	cont, err := container.NewContainer()
	if err != nil {
		panic(err)
	}

	serv, err := initServer(cont)
	if err != nil {
		cont.Logger.Error("Error with server initialization: %s", err)
	}
	if err := serv.run(); err != nil {
		panic(err)
	}
}

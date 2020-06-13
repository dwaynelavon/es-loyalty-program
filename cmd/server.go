package main

import (
	"github.com/dwaynelavon/es-loyalty-program/cmd/dependency"
	"go.uber.org/fx"
)

func main() {
	dependencies := fx.Provide(
		dependency.NewLogger,
		dependency.NewFirebaseApp,
		dependency.NewFirebaseClient,
		dependency.NewUserReadRepo,
		dependency.NewUserReadModel,
		dependency.NewDispatcher,
		dependency.NewEventBus,
		dependency.NewPointsMappingService,
	)

	modules := fx.Options()

	invocations := fx.Invoke(
		dependency.LoadEnv,
		dependency.RegisterEventHandlers,
		dependency.RegisterDispatchHandlers,
		dependency.RegisterRoutes,
	)

	app := fx.New(
		dependencies,
		modules,
		invocations,
	)

	app.Run()
}

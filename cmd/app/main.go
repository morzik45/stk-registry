package main

import (
	"context"
	"github.com/morzik45/stk-registry/pkg/config"
)

func main() {
	ctx := context.Background()
	cfg := config.GetConfig()
	app, err := NewApp(ctx, cfg)
	if err != nil {
		panic(err)
	}
	app.Run(ctx)
}

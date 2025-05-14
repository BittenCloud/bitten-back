package main

import (
	"bitback/internal/app"
	"context"
)

func main() {
	ctx := context.Background()
	application := app.NewApplication(ctx)
	application.Start()
}

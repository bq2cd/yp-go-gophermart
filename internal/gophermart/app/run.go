package app

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
)

// Run launches the application.
func Run() {
	var cli CLI

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	kctx := kong.Parse(
		&cli,
		kong.Vars(CLIVars()),
		kong.BindTo(ctx, (*context.Context)(nil)),
	)

	err := kctx.Run(ctx)

	kctx.FatalIfErrorf(err)
}

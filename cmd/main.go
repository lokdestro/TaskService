package main

import (
	"TaskService/config"
	"TaskService/pkg/logger"
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
	"syscall"

	"TaskService/internal/app"
	exit "TaskService/pkg/context"
	"github.com/joho/godotenv"
)

// @title Task Service API
// @version 1.0
// @description API для управления задачами
// @host localhost:3000
// @BasePath /
func init() {
	_ = godotenv.Load()

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()

	if err := logger.Init(config.Log()); err != nil {
		panic(err)
	}
}

func main() {
	app, err := app.New()
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx, cancel := exit.WithSignal(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := app.Run(ctx); err != nil {
		fmt.Println(err)
		return
	}

	return
}

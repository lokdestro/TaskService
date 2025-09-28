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

//export GOOSE_DBSTRING="user=dima password=1 dbname=networks sslmode=disable
//echo $GOOSE_DBSTRING

//export GOOSE_DRIVER=postgres
//echo $GOOSE_DRIVER

//sudo -u postgres psql
//\c networks

//swag init --generalInfo cmd/Cringe-Networks/main.go --output docs

func init() {
	_ = godotenv.Load()

	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()
	//
	//viper.SetConfigType("yaml")
	//if err := viper.ReadConfig(bytes.NewBuffer(config.Data)); err != nil {
	//	panic(err)
	//}

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

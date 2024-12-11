package main

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"project/cmd/database/environment"
	"project/cmd/database/gorm"
	"project/cmd/webapi"
	"project/internal/api"
	repository "project/internal/repositories"
	"project/internal/service"
	"syscall"
)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))

	err := godotenv.Load()
	if err != nil {
		logrus.Fatalf("error loading .env file: %v", err.Error())
	}

	dbConfig, err := environment.GetDbConfigWithEnvs()
	if err != nil {
		logrus.Fatalf("error database connection: %v", err.Error())
	}

	db, err := gormSql.NewGormSqlDB(dbConfig)
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	router := api.NewHandler(services)

	srv := new(webapi.Server)

	go func() {
		if err := srv.Run(os.Getenv("SERVER_PORT"), router.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("Project is Started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("Project is Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	if err := db.Close(); err != nil {
		logrus.Errorf("error occured on db connection close: %s", err.Error())
	}
}

package main

import (
	"L0/pkg/database"
	"L0/pkg/handler"
	"L0/pkg/models"
	"L0/pkg/mynats"
	"L0/pkg/server"
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cache = make(map[int]models.Order)

func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	err := initConfig()
	if err != nil {
		logrus.Fatalf("initializing config error: %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		logrus.Fatalf("error loading env variables: %s", err.Error())
	}

	//Connect to db
	db, err := database.NewPostgresDB(database.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("DB_PASSWORD"),
	})
	if err != nil {
		logrus.Fatalf("failed to initialize db: %s", err.Error())
	}

	//Get All data from bd
	orders, err := database.GetOrders(db)
	if err != nil {
		logrus.Fatalf("failed to save in cache: %s", err.Error())
	}

	//Save orders in cache
	for _, order := range orders {
		cache[order.Id] = order
	}
	mynats.Cache = cache

	//Connect to NATS
	js, nc, err := mynats.ConnectToNATS()
	if err != nil {
		logrus.Fatalf("failed connect to nats streaming : %s", err.Error())
	}
	defer nc.Close()

	//Create a stream
	err = mynats.CreateStream(js, "Test")
	if err != nil {
		logrus.Fatalf("failed create a stream : %s", err.Error())
	}

	//Subscribe to NATS
	sub, err := mynats.SubscribeToJetStream(js, "test", db)
	if err != nil {
		logrus.Fatalf("failed subscribe to jetstream: %s", err.Error())
	}

	defer sub.Unsubscribe()

	//start server
	handlers := handler.NewHandler()
	srv := new(server.Server)

	go func() {
		if err := srv.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
			logrus.Fatalf("error occured while running http server: %s", err.Error())
		}
	}()

	logrus.Print("Site Started")

	//Off server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("Site Shutting Down")

	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	// if err := db.Close(); err != nil {
	// 	logrus.Errorf("error occured on db connection close: %s", err.Error())
	// }

}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

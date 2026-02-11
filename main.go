package main

import (
	"daka/functions"
	"daka/utils"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/consul/api"
	"github.com/nats-io/nats.go"
)

func main() {
	cfg, err := utils.LoadConfig("config.yml")
	fmt.Println("::: 开始服务 :::")
	if err != nil {
		log.Fatal("读起config.yml失败:", err)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.Mysql.Username, cfg.Mysql.Password, cfg.Mysql.Ip, cfg.Mysql.Port, cfg.Mysql.Database)
	fmt.Println(dsn)
	sb2, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("打开DB失败:", err)
	}
	defer sb2.Close()

	if err := sb2.Ping(); err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	_, err = nc.Subscribe("phone.tasks", func(msg *nats.Msg) {
		fmt.Printf("Received message: %s\n", string(msg.Data))

		processed := functions.HandleTask(sb2, string(msg.Data))
		if processed == false {
			fmt.Printf("Error processing task: %s\n", string(msg.Data))
			return
		}
		msg.Ack()
	})

	if err != nil {
		log.Fatalf("Error subscribing to phone.tasks: %v", err)
	}
	go startGinServer(nc, cfg)
	RegisterConsulService(cfg)
	fmt.Println("Subscriber is running... Waiting for messages.")
	select {}
}
func startGinServer(nc *nats.Conn, cfg *utils.Config) {
	r := gin.Default()
	r.GET("/status", func(c *gin.Context) {
		status := "Subscriber is not running"
		if nc.IsConnected() {
			status = "Subscriber is running"
		}

		c.JSON(http.StatusOK, gin.H{
			"status": status,
		})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	port := strconv.Itoa(cfg.Server.Port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error starting Gin server: %v", err)
	}
}

func RegisterConsulService(cfg *utils.Config) {
	consulCfg := api.DefaultConfig()
	consulCfg.Address = fmt.Sprintf("%s:%d", cfg.Consul.Server, cfg.Consul.Port) //"consul:8500"
	client, err := api.NewClient(consulCfg)
	if err != nil {
		log.Fatal("Failed to create Consul client:", err)
	}

	hostIP := utils.GetLocalIP()
	if cfg.Consul.LocalIp {
		hostIP = "127.0.0.1"
	}

	registration := &api.AgentServiceRegistration{
		ID:      cfg.Server.ID,
		Name:    cfg.Server.Name,
		Address: hostIP,
		Port:    cfg.Server.Port,
		Check: &api.AgentServiceCheck{
			HTTP:     fmt.Sprintf("http://%s:%d/health", hostIP, cfg.Server.Port),
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatal("Failed to register service with Consul:", err)
	}

	log.Println("Registered service with Consul:", cfg.Server.Name, "on port", cfg.Server.Port)
}

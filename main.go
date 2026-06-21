package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := flag.Int("port", 6284, "服务监听端口")
	dbPath := flag.String("db", "ocarina.db", "SQLite数据库文件路径")
	flag.Parse()

	db, err := InitDB(*dbPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()

	srv := &Server{DB: db}

	addr := fmt.Sprintf(":%d", *port)
	fmt.Fprintf(os.Stderr, "陶笛工作室管理系统启动中...\n")
	fmt.Fprintf(os.Stderr, "访问 http://localhost:%d\n", *port)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

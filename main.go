package main

import (
	"fmt"
	"log"
	"strings"

	"market-pulse/internal/config"
	"market-pulse/internal/scraper"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("โหลด config ไม่สำเร็จ: %v", err)
	}

	fmt.Println("กำลังดึงข้อมูล...")

	posts, err := scraper.FetchGroupPosts(cfg.GroupURL(), cfg.FacebookCUser, cfg.FacebookXS, 6)
	if err != nil {
		log.Fatalf("เกิดข้อผิดพลาด: %v", err)
	}

	if len(posts) == 0 {
		fmt.Println("ไม่พบโพสต์ในกลุ่ม")
		return
	}

	formatted := make([]string, len(posts))
	for i, p := range posts {
		formatted[i] = p.String()
	}

	finalText := strings.Join(formatted, "\n-------------------\n")

	fmt.Println("=== ข้อมูลที่พร้อมส่งเข้า OpenRouter ===")
	fmt.Println(finalText)
}

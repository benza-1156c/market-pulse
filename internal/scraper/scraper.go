package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Comment struct {
	Post
}

type Post struct {
	Author   string
	Time     string
	Content  string
	Comments []Comment
}

func (p Post) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ผู้โพสต์: %s\n", p.Author))
	if p.Time != "" {
		sb.WriteString(fmt.Sprintf("เวลา: %s\n", p.Time))
	}
	sb.WriteString(fmt.Sprintf("รายละเอียด: %s", p.Content))

	if len(p.Comments) > 0 {
		sb.WriteString("\nความคิดเห็น:")
		for i, c := range p.Comments {
			timeStr := ""
			if c.Time != "" {
				timeStr = fmt.Sprintf(" (%s)", c.Time)
			}
			sb.WriteString(fmt.Sprintf("\n  [%d] %s%s: %s", i+1, c.Author, timeStr, c.Content))
		}
	}
	return sb.String()
}

func FetchGroupPosts(targetURL, cUser, xs string, maxScrolls int) ([]Post, error) {
	u, err := launcher.New().
		Headless(false).
		Launch()
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเปิดเบราว์เซอร์ได้: %w", err)
	}

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: ""})
	if err != nil {
		return nil, fmt.Errorf("สร้างแท็บใหม่ไม่สำเร็จ: %w", err)
	}

	err = page.SetCookies([]*proto.NetworkCookieParam{
		{
			Name:   "c_user",
			Value:  cUser,
			Domain: ".facebook.com",
			Path:   "/",
		},
		{
			Name:   "xs",
			Value:  xs,
			Domain: ".facebook.com",
			Path:   "/",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ตั้งค่า Cookie ไม่สำเร็จ: %w", err)
	}

	if err := page.Navigate(targetURL); err != nil {
		return nil, fmt.Errorf("เข้าหน้าเว็บไม่สำเร็จ: %w", err)
	}

	time.Sleep(3 * time.Second)

	collector := NewCollector()

	for i := range maxScrolls {
		fmt.Printf("กำลังเลื่อนหน้าจอ ครั้งที่ %d/%d (พบสะสม %d โพสต์)...\n", i+1, maxScrolls, len(collector.Posts))

		_ = page.Mouse.Scroll(0, 2500, 10)
		_, _ = page.Eval(`() => window.scrollBy(0, 2000)`)
		time.Sleep(2 * time.Second)

		expandSeeMore(page)

		html, err := page.HTML()
		if err != nil {
			continue
		}

		collector.Collect(html)
	}

	return collector.Posts, nil
}

func expandSeeMore(page *rod.Page) {
	_, _ = page.Eval(`() => {
		document.querySelectorAll('div[role="button"], span[role="button"]').forEach(btn => {
			const txt = (btn.innerText || "").trim();
			if (txt.includes("ดูเพิ่มเติม") || txt.includes("See more")) {
				btn.click();
			}
		});
	}`)
	time.Sleep(500 * time.Millisecond)
}

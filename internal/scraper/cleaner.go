package scraper

import (
	"regexp"
	"strings"
)

var garbageWords = []string{
	"ถูกใจ", "แสดงความคิดเห็น", "แชร์", "ส่ง", "ดูเพิ่มเติม", "ดูน้อยลง",
	"ความคิดเห็น", "Like", "Comment", "Share", "Send", "See more", "See less",
	"เข้าสู่ระบบ", "สร้างบัญชีใหม่", "ลืมรหัสผ่านใช่หรือไม่",
	"Facebookเปิดแอพ", "กลุ่มสาธารณะ", "สมาชิกเข้าร่วมกลุ่ม",
	"กรุงเทพมหานคร ประเทศไทย", "กำลังโหลด…", "ลองอีกครั้ง",
	"| FacebookFacebook", "เข้าร่วมกลุ่ม",
}

var (
	reSpace     = regexp.MustCompile(`[ \t]+`)
	reLine      = regexp.MustCompile(`[\r\n]+`)
	reTimestamp = regexp.MustCompile(`(?i)^\s*(\d+\s*(นาที|วินาที|ชม\.|ชั่วโมง|วัน|สัปดาห์|เดือน|ปี|m|min|mins|h|hr|hrs|d|w)|เมื่อสักครู่|เมื่อครู่|just now)\s*[·•\s]*`)
)

func cleanText(text string) string {
	for _, word := range garbageWords {
		text = strings.ReplaceAll(text, word, "")
	}

	text = reTimestamp.ReplaceAllString(text, "")
	text = reSpace.ReplaceAllString(text, " ")
	text = reLine.ReplaceAllString(text, "\n")

	return strings.TrimSpace(text)
}

func cleanTime(t string) string {
	t = strings.ReplaceAll(t, "ประมาณ", "")
	t = strings.ReplaceAll(t, "ที่แล้ว", "")
	t = strings.ReplaceAll(t, "about", "")
	t = strings.ReplaceAll(t, "ago", "")
	return strings.TrimSpace(t)
}

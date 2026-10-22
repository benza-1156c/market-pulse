package scraper

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PostCollector struct {
	seen  map[string]bool
	Posts []Post
}

func NewCollector() *PostCollector {
	return &PostCollector{
		seen: make(map[string]bool),
	}
}

func (c *PostCollector) Collect(html string) []Post {
	newPosts := extractPostsFromHTML(html)
	var added []Post

	for _, p := range newPosts {
		key := p.Author + "::" + truncateSnippet(p.Content, 50)
		if !c.seen[key] {
			c.seen[key] = true
			c.Posts = append(c.Posts, p)
			added = append(added, p)

			timeTag := ""
			if p.Time != "" {
				timeTag = fmt.Sprintf(" [%s]", p.Time)
			}
			commentTag := ""
			if len(p.Comments) > 0 {
				commentTag = fmt.Sprintf(" (%d คอมเมนต์)", len(p.Comments))
			}
			fmt.Printf("    + ได้โพสต์ใหม่: %s%s%s (%s)\n", p.Author, timeTag, commentTag, truncateSnippet(p.Content, 40))
		}
	}

	return added
}

func extractPostsFromHTML(html string) []Post {
	var posts []Post

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return posts
	}

	doc.Find(`[role="article"]`).Each(func(_ int, s *goquery.Selection) {
		h2 := s.Find("h2")
		if h2.Length() == 0 {
			return
		}

		author := strings.TrimSpace(h2.First().Text())
		if author == "" || strings.Contains(author, "ความคิดเห็น") || strings.Contains(author, "แชร์") || strings.Contains(author, "เกี่ยวกับ") {
			return
		}

		author = strings.TrimSpace(strings.ReplaceAll(author, "· ติดตาม", ""))
		author = strings.TrimSpace(strings.ReplaceAll(author, "· Follow", ""))

		timeStr := cleanTime(extractPostTime(s))

		var content string
		msg := s.Find(`div[data-ad-rendering-role="story_message"], div[data-ad-preview="message"], div[data-ad-comet-preview="message"]`)
		if msg.Length() > 0 {
			tabBlocks := msg.Find(`div[tabindex="-1"]`)
			if tabBlocks.Length() > 0 {
				var lines []string
				tabBlocks.Each(func(_ int, d *goquery.Selection) {
					t := strings.TrimSpace(d.Text())
					if t != "" && t != "ดูเพิ่มเติม" && t != "See more" {
						lines = append(lines, t)
					}
				})
				content = strings.Join(lines, "\n")
			} else {
				content = msg.Text()
			}
		} else {
			clone := s.Clone()
			clone.Find(`[role="article"], h2, svg, img, script, style, [role=button], form`).Remove()
			content = clone.Text()
		}

		comments := extractCommentsFromArticle(s)

		content = cleanText(content)
		if len(content) > 1 {
			posts = append(posts, Post{
				Author:   author,
				Time:     timeStr,
				Content:  content,
				Comments: comments,
			})
		}
	})

	return posts
}

func extractCommentsFromArticle(s *goquery.Selection) []Comment {
	var comments []Comment

	s.Find(`div[role="article"]`).Each(func(_ int, nested *goquery.Selection) {
		ariaLabel, _ := nested.Attr("aria-label")
		var author, postTime string

		if strings.HasPrefix(ariaLabel, "ความคิดเห็นจาก ") {
			parts := strings.Split(strings.TrimPrefix(ariaLabel, "ความคิดเห็นจาก "), " เมื่อ ")
			if len(parts) == 2 {
				author = strings.TrimSpace(parts[0])
				postTime = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(ariaLabel, "Comment by ") {
			parts := strings.Split(strings.TrimPrefix(ariaLabel, "Comment by "), " ")
			if len(parts) >= 2 {
				author = strings.TrimSpace(parts[0])
			}
		}

		if author == "" {
			author = strings.TrimSpace(nested.Find(`a[role="link"]`).First().Text())
		}

		if author == "" {
			return
		}

		clone := nested.Clone()
		clone.Find("a, [role=button], form, svg, img").Remove()
		raw := cleanText(clone.Text())

		content := strings.TrimSpace(strings.TrimPrefix(raw, "·"))
		content = strings.TrimSpace(strings.ReplaceAll(content, "ผู้เขียน", ""))
		content = strings.TrimLeft(content, "· \t\n")

		if content != "" {
			comments = append(comments, Comment{
				Post: Post{
					Author:  author,
					Time:    cleanTime(postTime),
					Content: content,
				},
			})
		}
	})

	return comments
}

func extractPostTime(s *goquery.Selection) string {
	var postTime string

	s.Find(`a[href*="/posts/"], a[href*="/permalink/"], a[href*="story_fbid"]`).EachWithBreak(func(_ int, a *goquery.Selection) bool {
		if parent := a.Parent(); parent.Length() > 0 {
			if hiddenDiv := parent.Find(`div[aria-hidden="true"]`); hiddenDiv.Length() > 0 {
				t := strings.TrimSpace(hiddenDiv.First().Text())
				if t != "" && len(t) < 50 {
					postTime = t
					return false
				}
			}
		}

		if label, exists := a.Attr("aria-label"); exists && strings.TrimSpace(label) != "" {
			t := strings.TrimSpace(label)
			if len(t) < 50 {
				postTime = t
				return false
			}
		}

		t := strings.TrimSpace(a.Text())
		if t != "" && len(t) < 50 {
			postTime = t
			return false
		}

		return true
	})

	if postTime != "" {
		return postTime
	}

	s.Find(`a[aria-label], span[aria-label]`).EachWithBreak(func(_ int, el *goquery.Selection) bool {
		label, _ := el.Attr("aria-label")
		label = strings.TrimSpace(label)
		if len(label) < 50 && (strings.Contains(label, "นาที") || strings.Contains(label, "ชั่วโมง") ||
			strings.Contains(label, "วัน") || strings.Contains(label, "เมื่อสักครู่") ||
			strings.Contains(label, "min") || strings.Contains(label, "hr")) {
			postTime = label
			return false
		}
		return true
	})

	return postTime
}

func truncateSnippet(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

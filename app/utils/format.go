package utils

import (
	"Planning_poker/app/models"
	"bytes"
	"fmt"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"html/template"
	"net/url"
	"regexp"
	"strings"
)

var emoticons = map[string]string{
	":)":        "smile",
	":(":        "sad",
	":P":        "tongue",
	":D":        "biggrin",
	";)":        "wink",
	"(y)":       "thumbsup",
	"(n)":       "thumbsdown",
	"(i)":       "information",
	"(/)":       "check",
	"(x)":       "error",
	"(!)":       "warning",
	"(+)":       "add",
	"(-)":       "forbidden",
	"(?)":       "help16",
	"(on)":      "lightbulbon",
	"(off)":     "lightbulb",
	"(*)":       "staryellow",
	"(*r)":      "starred",
	"(*g)":      "stargreen",
	"(*b)":      "starblue",
	"(flag)":    "flag",
	"(flagoff)": "flaggrey",
}

var validMimeTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
}

func Format(content string, attachments map[string]models.AttachmentInfo) string {
	content = processNestedBlocks(content, attachments)
	content = processContent(content, attachments)
	content = strings.ReplaceAll(content, "\r\n", "<br>")
	content = strings.ReplaceAll(content, `\"`, `"`)

	return content
}

func processNestedBlocks(content string, attachments map[string]models.AttachmentInfo) string {
	panelRegex := regexp.MustCompile(`(?s)\{panel:title=([^}]*)\}(.*?)\{panel\}`)
	noformatRegex := regexp.MustCompile(`(?s)\{noformat\}(.*?)\{noformat\}`)

	content = panelRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := panelRegex.FindStringSubmatch(match)
		if len(matches) > 2 {
			title := matches[1]
			content := processContent(matches[2], attachments)
			return fmt.Sprintf(`<div class="panel"><div class="panel-title">%s</div><div class="panel-content">%s</div></div>`, title, content)
		}
		return match
	})

	content = noformatRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := noformatRegex.FindStringSubmatch(match)
		if len(matches) > 1 {
			innerContent := template.HTMLEscapeString(matches[1])
			return fmt.Sprintf(`<pre class="no-format">%s</pre>`, innerContent)
		}
		return match
	})

	return content
}

func processContent(content string, attachments map[string]models.AttachmentInfo) string {
	content = processCodeBlocks(content)
	content = processLists(content)
	content = processLinksAndEmoticons(content)
	content = processStyles(content)
	content = processImagesByContext(content, attachments)
	content = processTables(content)

	horizontalLineRegex := regexp.MustCompile(`----`)
	content = horizontalLineRegex.ReplaceAllString(content, `<hr>`)

	return content
}

func processTables(description string) string {
	tableHeaderRegex := regexp.MustCompile(`^\|\|(.+)$`)
	tableRowRegex := regexp.MustCompile(`^\|(.+)$`)
	tableBlockRegex := regexp.MustCompile(`(?m)(\|\|.+\n(?:\|.+\n?)*)`)

	placeholders := []string{}
	description = tableBlockRegex.ReplaceAllStringFunc(description, func(match string) string {
		placeholders = append(placeholders, match)
		return "{table}"
	})

	for _, tableBlock := range placeholders {
		var tableHeaders []string
		var tableRows []string

		lines := strings.Split(tableBlock, "\n")
		for _, line := range lines {
			if tableHeaderRegex.MatchString(line) {
				matches := tableHeaderRegex.FindStringSubmatch(line)
				if len(matches) > 1 {
					headers := strings.Split(matches[1], "||")
					row := "<tr>" + strings.Join(mapFunc(headers, func(header string) string {
						if strings.TrimSpace(header) == "" {
							return ""
						}
						return "<th>" + strings.TrimSpace(header) + "</th>"
					}), "") + "</tr>"
					tableHeaders = append(tableHeaders, row)
				}
			} else if tableRowRegex.MatchString(line) {
				matches := tableRowRegex.FindStringSubmatch(line)
				if len(matches) > 1 {
					columns := strings.Split(matches[1], "|")
					row := "<tr>" + strings.Join(mapFunc(columns, func(column string) string {
						if strings.TrimSpace(column) == "" {
							return ""
						}
						return "<td>" + strings.TrimSpace(column) + "</td>"
					}), "") + "</tr>"
					tableRows = append(tableRows, row)
				}
			}
		}

		table := `<table class="jira-table">`
		if len(tableHeaders) > 0 {
			table += "<thead>" + strings.Join(tableHeaders, "") + "</thead>"
		}
		if len(tableRows) > 0 {
			table += "<tbody>" + strings.Join(tableRows, "") + "</tbody>"
		}
		table += "</table>"

		description = strings.Replace(description, "{table}", table, 1)
	}

	return description
}

func mapFunc(slice []string, f func(string) string) []string {
	var newSlice []string
	for _, item := range slice {
		newSlice = append(newSlice, f(item))
	}
	return newSlice
}

func processCodeBlocks(description string) string {
	codeRegex := regexp.MustCompile(`(?s)\{code:(\w+)\}(.*?)\{code\}`)
	return codeRegex.ReplaceAllStringFunc(description, func(match string) string {
		matches := codeRegex.FindStringSubmatch(match)
		if len(matches) > 2 {
			language := matches[1]
			code := matches[2]
			return highlightCode(code, language)
		}
		return match
	})
}

func processLists(description string) string {
	listBulletRegex := regexp.MustCompile(`(?m)^ \* (.*?)$`)
	listNumberRegex := regexp.MustCompile(`(?m)^ # (.*?)$`)
	description = listBulletRegex.ReplaceAllString(description, `<ul><li>$1</li></ul>`)
	description = listNumberRegex.ReplaceAllString(description, `<ol><li>$1</li></ol>`)
	description = strings.Replace(description, "</ul><ul>", "", -1)
	description = strings.Replace(description, "</ol><ol>", "", -1)
	return description
}

func processStyles(description string) string {
	boldRegex := regexp.MustCompile(`\*(.*?)\*`)
	italicRegex := regexp.MustCompile(`_(.*?)_`)
	underlineRegex := regexp.MustCompile(`\+(.*?)\+`)
	colorRegex := regexp.MustCompile(`\{color:(#[0-9a-fA-F]{6})\}(.*?)\{color\}`)
	description = boldRegex.ReplaceAllString(description, `<b>$1</b>`)
	description = italicRegex.ReplaceAllString(description, `<i>$1</i>`)
	description = underlineRegex.ReplaceAllString(description, `<u>$1</u>`)
	description = colorRegex.ReplaceAllString(description, `<span style="color:$1;">$2</span>`)
	return description
}

func processLinksAndEmoticons(description string) string {
	linkWithTextRegex := regexp.MustCompile(`\[(.+?)\|((https?|ftp):\/\/[^\s\]]+)\]`)
	description = linkWithTextRegex.ReplaceAllString(description, `<a href="$2" target="_blank">$1</a>`)

	linkRegex := regexp.MustCompile(`\[(https?://[^\s]+)\]`)
	description = linkRegex.ReplaceAllString(description, `<a href="$1" target="_blank">$1</a>`)

	for k, v := range emoticons {
		escaped := regexp.QuoteMeta(k)
		regex := regexp.MustCompile(escaped)
		spanTag := fmt.Sprintf(`<span class="%s"></span>`, v)
		description = regex.ReplaceAllString(description, spanTag)
	}
	return description
}

func processImagesByContext(description string, attachments map[string]models.AttachmentInfo) string {
	imageRegex := regexp.MustCompile(`!(image-\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-\d{3}\.(png|jpg|jpeg|gif))(?:\|width=(\d+),height=(\d+))?!`)

	return imageRegex.ReplaceAllStringFunc(description, func(match string) string {
		matches := imageRegex.FindStringSubmatch(match)
		if len(matches) > 1 {
			filename := matches[1]
			width := matches[2]
			height := matches[3]

			if attachment, ok := attachments[filename]; ok {
				if _, isValid := validMimeTypes[attachment.MimeType]; isValid {
					proxyUrl := fmt.Sprintf("/image-proxy?url=%s", url.QueryEscape(attachment.Content))
					if width != "" && height != "" {
						return fmt.Sprintf(`<img src="%s" alt="%s" width="%s" height="%s" />`, proxyUrl, filename, width, height)
					}
					return fmt.Sprintf(`<img src="%s" alt="%s" />`, proxyUrl, filename)
				}
			}
		}
		return match
	})
}

func highlightCode(code, language string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := html.New(html.WithClasses(true))

	var buf bytes.Buffer
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return fmt.Sprintf("<pre>%s</pre>", template.HTMLEscapeString(code))
	}

	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return fmt.Sprintf("<pre>%s</pre>", template.HTMLEscapeString(code))
	}

	return buf.String()
}

func GetElapsedTime(elapsedTime int64) string {
	hours := elapsedTime / 3600
	minutes := (elapsedTime % 3600) / 60
	seconds := elapsedTime % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

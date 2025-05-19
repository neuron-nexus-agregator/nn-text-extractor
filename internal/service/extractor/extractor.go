package extractor

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"agregator/text-extractor/internal/model/configs"
)

type Extractor struct {
	cfg *configs.Configs
}

type Article struct {
	Content string
}

func New(configFileParh string) (*Extractor, error) {
	cfg := &configs.Configs{}
	err := cfg.Init(configFileParh)
	if err != nil {
		return nil, err
	}
	return &Extractor{
		cfg: cfg,
	}, nil
}

func (e *Extractor) Extract(site string) (string, error) {
	site = strings.ReplaceAll(site, "https://www.", "https://")
	site = strings.ReplaceAll(site, "http://www.", "http://")
	parsedURL, err := url.Parse(site)
	if err != nil {
		return "", fmt.Errorf("не удалось спарсить URL: %w", err)
	}

	config, ok := e.cfg.SiteConfigs[parsedURL.Hostname()]
	if !ok {
		return "", fmt.Errorf("конфигурация для домена %s не найдена", parsedURL.Hostname())
	}

	req, err := http.NewRequest("GET", site, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; NewsAggregator/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("неуспешный статус код: %d — %s", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении тела ответа: %w", err)
	}

	// Правильная обработка возвращаемых значений
	_, encodingName, _ := charset.DetermineEncoding(body, res.Header.Get("Content-Type"))
	if encodingName == "" {
		encodingName = "utf-8"
		log.Println("Не удалось определить кодировку, используем UTF-8 по умолчанию")
	}

	utf8Body, err := convertEncoding(body, encodingName, "utf-8")
	if err != nil {
		log.Printf("Ошибка преобразования кодировки: %v, продолжаем с исходным текстом", err)
		utf8Body = body
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(utf8Body))
	if err != nil {
		return "", fmt.Errorf("ошибка при парсинге HTML: %w", err)
	}

	content := extractContent(doc, &config)
	if content == "" {
		log.Default().Println("Контент не найден в основных контейнерах, поиск в body...")
		content = extractFallbackContent(doc)
	}

	return strings.TrimSpace(content), nil
}

func extractContent(doc *goquery.Document, config *configs.Config) string {
	var contentBuilder strings.Builder

	doc.Find(config.TextDivSelector).Each(func(i int, s *goquery.Selection) {
		appendText(&contentBuilder, s.Text(), "p")
	})

	for _, containerSel := range config.ContentSelectors {
		container := doc.Find(containerSel)
		if container.Length() > 0 {
			container.Find(config.ParagraphSelector).Each(func(i int, s *goquery.Selection) {
				appendText(&contentBuilder, s.Text(), "p")
			})
			container.Find(config.ListItemSelector).Each(func(i int, s *goquery.Selection) {
				appendText(&contentBuilder, cleanListItem(s), "li")
			})
			container.Find(config.TableSelector).Each(func(i int, s *goquery.Selection) {
				appendHTML(&contentBuilder, s)
			})
			break
		}
	}

	return contentBuilder.String()
}

func extractFallbackContent(doc *goquery.Document) string {
	var contentBuilder strings.Builder

	doc.Find("body").Children().Each(func(i int, child *goquery.Selection) {
		if child.Is("p") {
			appendText(&contentBuilder, child.Text(), "p")
		} else if child.Is("table") {
			appendHTML(&contentBuilder, child)
		}
	})

	return contentBuilder.String()
}

func appendText(builder *strings.Builder, text, tag string) {
	text = strings.TrimSpace(text)
	if len(text) > 0 {
		builder.WriteString(fmt.Sprintf("<%s>%s</%s>\n", tag, text, tag))
	}
}

func appendHTML(builder *strings.Builder, selection *goquery.Selection) {
	html, _ := goquery.OuterHtml(selection)
	builder.WriteString(html + "\n")
}

func cleanListItem(selection *goquery.Selection) string {
	text := strings.TrimSpace(selection.Text())
	selection.Find(".article__list-label").Each(func(_ int, label *goquery.Selection) {
		labelText := strings.TrimSpace(label.Text())
		text = strings.Replace(text, labelText, "", 1)
	})
	return text
}

func convertEncoding(body []byte, fromEncodingName, toEncodingName string) ([]byte, error) {
	if toEncodingName != "utf-8" {
		return body, fmt.Errorf("целевая кодировка должна быть UTF-8")
	}

	var fromEncoding encoding.Encoding
	switch strings.ToLower(fromEncodingName) {
	case "windows-1251":
		fromEncoding = charmap.Windows1251
	case "iso-8859-5":
		fromEncoding = charmap.ISO8859_5
	case "koi8-r":
		fromEncoding = charmap.KOI8R
	case "utf-8", "":
		return body, nil
	default:
		return nil, fmt.Errorf("неподдерживаемая кодировка: %s", fromEncodingName)
	}

	reader := transform.NewReader(bytes.NewReader(body), fromEncoding.NewDecoder())
	utf8Bytes, err := io.ReadAll(reader)
	if err != nil {
		return body, fmt.Errorf("ошибка преобразования кодировки из %s в UTF-8: %w", fromEncodingName, err)
	}

	return utf8Bytes, nil
}

package service

import (
	"encoding/json"
	"errors"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tigerowo/infinite-canvas/model"
	"github.com/tigerowo/infinite-canvas/repository"
)

const (
	gptImage2RawBase             = "https://raw.githubusercontent.com/tigerowo/awesome-gpt-image-2-prompts/main"
	awesomeGptImageRawBase       = "https://raw.githubusercontent.com/ZeroLu/awesome-gpt-image/main"
	awesomeGpt4oImagePromptsBase = "https://raw.githubusercontent.com/ImgEdify/Awesome-GPT4o-Image-Prompts/main"
	youMindGptImage2RawBase      = "https://raw.githubusercontent.com/YouMind-OpenLab/awesome-gpt-image-2/main"
	youMindNanoBananaProRawBase  = "https://raw.githubusercontent.com/YouMind-OpenLab/awesome-nano-banana-pro-prompts/main"
	xianyuAwesomeGptImage2RawBase = "https://raw.githubusercontent.com/xianyu110/awesome-gptimage2/main"
	davidWuGptImage2RawBase      = "https://raw.githubusercontent.com/davidwuw0811-boop/awesome-gpt-image2-prompts/main"
	museVideoPromptsRawBase      = "https://raw.githubusercontent.com/apimodels/muse-video-prompts/main"
	museVideoPromptsRepository   = "https://github.com/apimodels/muse-video-prompts"
	museVideoPromptsMediaPage    = "https://apimodels.app/zh/muse-video-prompts"
)

var gptImage2CaseFiles = []string{"README.md", "cases/ad-creative.md", "cases/character.md", "cases/comparison.md", "cases/ecommerce.md", "cases/portrait.md", "cases/poster.md", "cases/ui.md"}

var (
	museVideoThumbnailURLPattern = regexp.MustCompile(`https://pbs\.twimg\.com/amplify_video_thumb/[^\s"'<>\\]+?\.(?:jpg|jpeg|png)(?:\?[^\s"'<>\\]+)?`)
	museVideoMP4URLPattern       = regexp.MustCompile(`https://video\.twimg\.com/[^\s"'<>\\]+?\.mp4(?:\?[^\s"'<>\\]+)?`)
	museVideoResolutionPattern   = regexp.MustCompile(`/([0-9]+)x([0-9]+)/`)
)

type gptImage2Data struct {
	Records []struct {
		Title    string `json:"title"`
		TweetURL string `json:"tweet_url"`
		ImageDir string `json:"image_dir"`
		Category string `json:"category"`
		AddedAt  string `json:"added_at"`
	} `json:"records"`
}

type xianyuLatestPromptData struct {
	Dates []struct {
		Items []xianyuLatestPrompt `json:"items"`
	} `json:"dates"`
	Items []xianyuLatestPrompt `json:"items"`
}

type xianyuLatestPrompt struct {
	XURL            string   `json:"x_url"`
	URL             string   `json:"url"`
	Author          string   `json:"author"`
	CreatedAt       string   `json:"created_at"`
	Text            string   `json:"text"`
	Prompt          string   `json:"prompt"`
	Reason          string   `json:"reason"`
	ImageURLs       []string `json:"image_urls"`
	PrimaryImageURL string   `json:"primary_image_url"`
}
type davidWuGptImage2Prompt struct {
	ID         int    `json:"id"`
	TitleEN    string `json:"title_en"`
	TitleCN    string `json:"title_cn"`
	Category   string `json:"category"`
	CategoryCN string `json:"category_cn"`
	Prompt     string `json:"prompt"`
	Note       string `json:"note"`
	Author     string `json:"author"`
	Source     string `json:"source"`
	NeedsRef   bool   `json:"needs_ref"`
	Image      string `json:"image"`
}

type museVideoPrompt struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Category       string  `json:"category"`
	Language       string  `json:"language"`
	Aspect         string  `json:"aspect"`
	Seconds        float64 `json:"seconds"`
	NeedsReference bool    `json:"needs_reference"`
	Author         string  `json:"author"`
	Date           string  `json:"date"`
	URL            string  `json:"url"`
	TextStatus     string  `json:"text_status"`
	Text           string  `json:"text"`
	UsageNote      string  `json:"usage_note"`
	Page           string  `json:"page"`
}

type museVideoPromptMedia struct {
	CoverURL string
	VideoURL string
}

type museVideoPromptMediaIndex struct {
	BySource map[string]museVideoPromptMedia
	ByText   map[string]museVideoPromptMedia
}

type museVideoPromptJSONLD struct {
	Type            string `json:"@type"`
	URL             string `json:"url"`
	IsBasedOn       string `json:"isBasedOn"`
	ContentURL      string `json:"contentUrl"`
	ThumbnailURL    string `json:"thumbnailUrl"`
	Description     string `json:"description"`
	ItemListElement []struct {
		Item struct {
			URL       string `json:"url"`
			IsBasedOn string `json:"isBasedOn"`
		} `json:"item"`
	} `json:"itemListElement"`
}

func SyncPromptCategory(category string) ([]model.PromptCategory, error) {
	for _, item := range repository.PromptCategories() {
		if item.Category != category {
			continue
		}
		items, err := buildPromptCategory(item.Category)
		if err != nil {
			return nil, err
		}
		if err := repository.ReplacePromptCategory(item, items); err != nil {
			return nil, err
		}
		return repository.ListPromptCategories()
	}
	return nil, errors.New("未知提示词分类")
}

func buildPromptCategory(category string) ([]model.Prompt, error) {
	switch category {
	case "gpt-image-2-prompts":
		return buildGptImage2Prompts()
	case "awesome-gpt-image":
		return buildAwesomeGptImagePrompts()
	case "awesome-gpt4o-image-prompts":
		return buildAwesomeGpt4oImagePrompts()
	case "xianyu-awesome-gptimage2":
		return buildXianyuAwesomeGptImage2Prompts()
	case "youmind-gpt-image-2":
		return buildYouMindGptImage2Prompts()
	case "youmind-nano-banana-pro":
		return buildYouMindNanoBananaProPrompts()
	case "davidwu-gpt-image2-prompts":
		return buildDavidWuGptImage2Prompts()
	case "muse-video-prompts":
		return buildMuseVideoPrompts()
	}
	return nil, errors.New("未知提示词分类")
}

func buildMuseVideoPrompts() ([]model.Prompt, error) {
	raw, err := fetchText(museVideoPromptsRawBase, "data/prompts.json")
	if err != nil {
		return nil, err
	}
	media := museVideoPromptMediaIndex{BySource: map[string]museVideoPromptMedia{}, ByText: map[string]museVideoPromptMedia{}}
	if mediaPage, err := fetchURL(museVideoPromptsMediaPage); err == nil {
		media = museVideoPromptMediaIndexFromHTML(mediaPage)
	}
	entries := []museVideoPrompt{}
	if err := json.Unmarshal([]byte(raw), &entries); err == nil {
		enrichMuseVideoPromptMediaFromSourcePages(entries, &media)
	}
	return buildMuseVideoPromptsFromJSONWithMediaIndex([]byte(raw), media)
}

func buildMuseVideoPromptsFromJSON(raw []byte) ([]model.Prompt, error) {
	return buildMuseVideoPromptsFromJSONWithMedia(raw, nil)
}

func buildMuseVideoPromptsFromJSONWithMedia(raw []byte, mediaBySource map[string]museVideoPromptMedia) ([]model.Prompt, error) {
	return buildMuseVideoPromptsFromJSONWithMediaIndex(raw, museVideoPromptMediaIndex{BySource: mediaBySource, ByText: map[string]museVideoPromptMedia{}})
}

func buildMuseVideoPromptsFromJSONWithMediaIndex(raw []byte, mediaIndex museVideoPromptMediaIndex) ([]model.Prompt, error) {
	data := []museVideoPrompt{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	items := make([]model.Prompt, 0, len(data))
	for _, entry := range data {
		id := strings.TrimSpace(entry.ID)
		title := strings.TrimSpace(entry.Title)
		prompt := strings.TrimSpace(entry.Text)
		if id == "" || title == "" || prompt == "" {
			continue
		}
		tags := []string{"muse-video"}
		for _, tag := range []string{strings.TrimSpace(entry.Category), strings.TrimSpace(entry.Language), strings.TrimSpace(entry.Aspect)} {
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		if entry.NeedsReference {
			tags = append(tags, "需要参考图")
		}
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(entry.TextStatus)), "full") {
			tags = append(tags, "完整文本")
		} else {
			tags = append(tags, "摘录")
		}
		if entry.Seconds > 0 {
			tags = append(tags, museVideoDuration(entry.Seconds)+"秒")
		}
		media := mediaIndex.BySource[normalizeMuseVideoSourceURL(entry.URL)]
		if media.VideoURL == "" {
			media = museVideoPromptMediaByText(entry.Text, mediaIndex.ByText)
		}
		items = append(items, model.Prompt{
			ID:        "muse-video-prompts-" + id,
			Title:     id + " · " + title,
			CoverURL:  media.CoverURL,
			VideoURL:  media.VideoURL,
			Prompt:    prompt,
			Tags:      tags,
			Preview:   museVideoPromptPreview(entry),
			CreatedAt: normalizePromptTime(entry.Date),
			UpdatedAt: normalizePromptTime(entry.Date),
		})
	}
	return items, nil
}

func museVideoPromptMediaFromHTML(html string) map[string]museVideoPromptMedia {
	return museVideoPromptMediaIndexFromHTML(html).BySource
}

func museVideoPromptMediaFromXHTML(source string) museVideoPromptMedia {
	content := html.UnescapeString(strings.ReplaceAll(strings.ReplaceAll(source, `\/`, "/"), `\u0026`, "&"))
	thumbnail := strings.TrimSpace(museVideoThumbnailURLPattern.FindString(content))
	videoURL := ""
	videoPixels := 0
	for _, candidate := range museVideoMP4URLPattern.FindAllString(content, -1) {
		candidate = strings.TrimSpace(candidate)
		match := museVideoResolutionPattern.FindStringSubmatch(candidate)
		if len(match) != 3 {
			continue
		}
		width, _ := strconv.Atoi(match[1])
		height, _ := strconv.Atoi(match[2])
		if width*height <= videoPixels {
			continue
		}
		videoPixels = width * height
		videoURL = candidate
	}
	return museVideoPromptMedia{CoverURL: thumbnail, VideoURL: videoURL}
}

func enrichMuseVideoPromptMediaFromSourcePages(entries []museVideoPrompt, mediaIndex *museVideoPromptMediaIndex) {
	type sourcePageResult struct {
		sourceURL string
		media     museVideoPromptMedia
	}
	missing := make([]museVideoPrompt, 0)
	for _, entry := range entries {
		media := mediaIndex.BySource[normalizeMuseVideoSourceURL(entry.URL)]
		if media.VideoURL == "" {
			media = museVideoPromptMediaByText(entry.Text, mediaIndex.ByText)
		}
		if media.VideoURL == "" && isMuseVideoSourceURL(entry.URL) {
			missing = append(missing, entry)
		}
	}
	if len(missing) == 0 {
		return
	}

	jobs := make(chan museVideoPrompt)
	results := make(chan sourcePageResult)
	workers := 4
	if len(missing) < workers {
		workers = len(missing)
	}
	var workersDone sync.WaitGroup
	for range workers {
		workersDone.Add(1)
		go func() {
			defer workersDone.Done()
			for entry := range jobs {
				page, err := fetchMuseVideoSourcePage(entry.URL)
				if err != nil {
					continue
				}
				media := museVideoPromptMediaFromXHTML(page)
				if media.VideoURL != "" && media.CoverURL != "" {
					results <- sourcePageResult{sourceURL: normalizeMuseVideoSourceURL(entry.URL), media: media}
				}
			}
		}()
	}
	go func() {
		for _, entry := range missing {
			jobs <- entry
		}
		close(jobs)
		workersDone.Wait()
		close(results)
	}()
	for result := range results {
		mediaIndex.BySource[result.sourceURL] = result.media
	}
}

func isMuseVideoSourceURL(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "https://x.com/") || strings.HasPrefix(value, "https://twitter.com/")
}

func fetchMuseVideoSourcePage(url string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	return fetchRequest(request)
}

func museVideoPromptMediaIndexFromHTML(html string) museVideoPromptMediaIndex {
	scripts := regexp.MustCompile(`(?is)<script[^>]+type=["']application/ld\+json["'][^>]*>(.*?)</script>`).FindAllStringSubmatch(html, -1)
	sourceByOnlineURL := map[string]string{}
	documents := make([]museVideoPromptJSONLD, 0, len(scripts))
	for _, script := range scripts {
		var document museVideoPromptJSONLD
		if err := json.Unmarshal([]byte(strings.TrimSpace(script[1])), &document); err != nil {
			continue
		}
		documents = append(documents, document)
		for _, listItem := range document.ItemListElement {
			onlineURL := strings.TrimSpace(listItem.Item.URL)
			sourceURL := normalizeMuseVideoSourceURL(listItem.Item.IsBasedOn)
			if onlineURL != "" && sourceURL != "" {
				sourceByOnlineURL[onlineURL] = sourceURL
			}
		}
	}

	mediaIndex := museVideoPromptMediaIndex{BySource: map[string]museVideoPromptMedia{}, ByText: map[string]museVideoPromptMedia{}}
	for _, document := range documents {
		if !strings.EqualFold(strings.TrimSpace(document.Type), "VideoObject") {
			continue
		}
		if strings.TrimSpace(document.ContentURL) == "" || strings.TrimSpace(document.ThumbnailURL) == "" {
			continue
		}
		media := museVideoPromptMedia{
			CoverURL: strings.TrimSpace(document.ThumbnailURL),
			VideoURL: strings.TrimSpace(document.ContentURL),
		}
		if sourceURL := sourceByOnlineURL[strings.TrimSpace(document.URL)]; sourceURL != "" {
			mediaIndex.BySource[sourceURL] = media
		}
		if textKey := museVideoPromptTextKey(document.Description); textKey != "" {
			mediaIndex.ByText[textKey] = media
		}
	}
	return mediaIndex
}

func normalizeMuseVideoSourceURL(value string) string {
	return strings.TrimSpace(value)
}

func museVideoPromptMediaByText(prompt string, mediaByText map[string]museVideoPromptMedia) museVideoPromptMedia {
	promptKey := museVideoPromptTextKey(prompt)
	if len([]rune(promptKey)) < 16 {
		return museVideoPromptMedia{}
	}
	for mediaText, media := range mediaByText {
		if strings.HasPrefix(mediaText, promptKey) || strings.HasPrefix(promptKey, mediaText) {
			return media
		}
	}
	return museVideoPromptMedia{}
}

func museVideoPromptTextKey(value string) string {
	value = strings.TrimSpace(strings.TrimRight(strings.TrimSpace(value), "…"))
	return strings.Join(strings.Fields(value), " ")
}

func museVideoPromptPreview(entry museVideoPrompt) string {
	lines := []string{museVideoPromptsRepository}
	if author := strings.TrimSpace(entry.Author); author != "" {
		lines = append(lines, "作者：@"+author)
	}
	if source := strings.TrimSpace(entry.URL); source != "" {
		lines = append(lines, "原帖："+source)
	}
	if page := strings.TrimSpace(entry.Page); page != "" {
		lines = append(lines, "条目："+museVideoPromptsRepository+"/blob/main/"+strings.TrimLeft(page, "/"))
	}
	metadata := []string{}
	if entry.Seconds > 0 {
		metadata = append(metadata, museVideoDuration(entry.Seconds)+" 秒")
	}
	if aspect := strings.TrimSpace(entry.Aspect); aspect != "" {
		metadata = append(metadata, aspect)
	}
	if entry.NeedsReference {
		metadata = append(metadata, "需要参考图")
	}
	if len(metadata) > 0 {
		lines = append(lines, "参数："+strings.Join(metadata, " · "))
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(entry.TextStatus)), "full") {
		lines = append(lines, "文本：完整文本（按仓库来源标注）")
	} else {
		lines = append(lines, "文本：摘录；完整内容请查看原帖")
	}
	if note := strings.TrimSpace(entry.UsageNote); note != "" {
		lines = append(lines, "用法："+note)
	}
	return strings.Join(lines, "\n\n")
}

func museVideoDuration(seconds float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(seconds, 'f', 2, 64), "0"), ".")
}

func fetchText(baseURL, file string) (string, error) {
	return fetchURL(baseURL + "/" + file)
}

func fetchURL(url string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	return fetchRequest(request)
}

func fetchRequest(request *http.Request) (string, error) {
	client := http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", errors.New(request.URL.String() + " 拉取失败")
	}
	data, err := io.ReadAll(response.Body)
	return string(data), err
}

type gptImage2Case struct {
	prompt string
	image  string
}

func buildGptImage2Prompts() ([]model.Prompt, error) {
	cases := map[string]gptImage2Case{}
	raw, err := fetchText(gptImage2RawBase, "data/ingested_tweets.json")
	if err != nil {
		return nil, err
	}
	data := gptImage2Data{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}
	for _, file := range gptImage2CaseFiles {
		markdown, err := fetchText(gptImage2RawBase, file)
		if err != nil {
			return nil, err
		}
		collectGptImage2Cases(cases, markdown)
	}
	items := []model.Prompt{}
	for _, item := range data.Records {
		c := cases[item.TweetURL]
		if c.prompt == "" {
			c = cases[item.ImageDir]
		}
		if c.prompt == "" {
			continue
		}
		image := c.image
		date := normalizePromptTime(item.AddedAt)
		items = append(items, model.Prompt{ID: "gpt-image-2-prompts-" + leftPad(len(items)+1), Title: item.Title, CoverURL: image, Prompt: c.prompt, Tags: tagsFromCategory(item.Category), CreatedAt: date, UpdatedAt: date, Preview: markdownPreview([]string{image})})
	}
	return items, nil
}

func collectGptImage2Cases(cases map[string]gptImage2Case, markdown string) {
	re := regexp.MustCompile("(?s)### Case \\d+: \\[[^\\]]+\\]\\(([^)]+)\\).*?\\*\\*Prompt:\\*\\*\\s*\\r?\\n\\s*```[\\w-]*\\r?\\n(.*?)\\r?\\n```")
	reImageDir := regexp.MustCompile(`images/\w+_case\d+`)
	reImage := regexp.MustCompile(`<img[^>]+src="([^"]+)"|!\[[^\]]*\]\(([^)]+)\)`)
	for _, match := range re.FindAllStringSubmatch(markdown, -1) {
		prompt := strings.TrimSpace(match[2])
		image := ""
		if imgMatch := reImage.FindStringSubmatch(match[0]); imgMatch != nil {
			if imgMatch[1] != "" {
				image = absoluteImage(gptImage2RawBase, imgMatch[1])
			} else {
				image = absoluteImage(gptImage2RawBase, imgMatch[2])
			}
		}
		item := gptImage2Case{prompt: prompt, image: image}
		cases[match[1]] = item
		if dir := reImageDir.FindString(match[0]); dir != "" {
			cases[dir] = item
		}
	}
}

func buildAwesomeGptImagePrompts() ([]model.Prompt, error) {
	markdown, err := fetchText(awesomeGptImageRawBase, "README.md")
	if err != nil {
		return nil, err
	}
	items := []model.Prompt{}
	for _, section := range splitBeforeHeading(markdown, "## ") {
		tags := tagsFromHeading(firstMatch(section, `(?m)^##\s+(.+)$`))
		for _, block := range splitBeforeHeading(section, "### ") {
			title := strings.TrimSpace(regexp.MustCompile(`\[([^\]]+)]\([^)]+\)`).ReplaceAllString(firstMatch(block, `(?m)^###\s+(.+)$`), "$1"))
			prompt := strings.TrimSpace(firstMatch(block, "(?s)\\*\\*Prompt:\\*\\*\\s*\\r?\\n\\s*```[\\w-]*\\r?\\n(.*?)\\r?\\n```"))
			if title == "" || prompt == "" {
				continue
			}
			images := extractMarkdownImages(awesomeGptImageRawBase, block)
			cover := ""
			if len(images) > 0 {
				cover = images[0]
			}
			items = append(items, model.Prompt{ID: "awesome-gpt-image-" + leftPad(len(items)+1), Title: title, CoverURL: cover, Prompt: prompt, Tags: tags, Preview: markdownPreview(images)})
		}
	}
	return items, nil
}

func buildAwesomeGpt4oImagePrompts() ([]model.Prompt, error) {
	markdown, err := fetchText(awesomeGpt4oImagePromptsBase, "README.zh-CN.md")
	if err != nil {
		return nil, err
	}
	items := []model.Prompt{}
	for _, block := range splitBeforeHeading(markdown, "### ") {
		title := strings.TrimSpace(firstMatch(block, `(?m)^###\s+(.+)$`))
		prompt := strings.TrimSpace(firstMatch(block, "(?s)- \\*\\*提示词文本：\\*\\*\\s*`(.*?)`"))
		if title == "" || prompt == "" {
			continue
		}
		images := extractMarkdownImages(awesomeGpt4oImagePromptsBase, block)
		cover := ""
		if len(images) > 0 {
			cover = images[0]
		}
		items = append(items, model.Prompt{ID: "awesome-gpt4o-image-prompts-" + leftPad(len(items)+1), Title: title, CoverURL: cover, Prompt: prompt, Tags: []string{"gpt4o"}, Preview: markdownPreview(images)})
	}
	return items, nil
}

func buildXianyuAwesomeGptImage2Prompts() ([]model.Prompt, error) {
	markdown, err := fetchText(xianyuAwesomeGptImage2RawBase, "README.md")
	if err != nil {
		return nil, err
	}
	items := parseXianyuPromptCollection(markdown)
	latest, err := buildXianyuLatestXPrompts(len(items))
	if err != nil {
		return nil, err
	}
	items = append(items, latest...)
	return items, nil
}

func parseXianyuPromptCollection(markdown string) []model.Prompt {
	section := markdownSection(markdown, "## 提示词合集", "## 高级技巧")
	items := []model.Prompt{}
	currentCategory := ""
	currentTitle := ""
	currentLines := []string{}
	finish := func() {
		if currentTitle == "" || currentCategory == "补充案例提示词" {
			return
		}
		block := strings.Join(currentLines, "\n")
		prompt := xianyuCodeBlockText(block)
		if prompt == "" {
			prompt = xianyuFallbackPromptText(block)
		}
		if prompt == "" {
			return
		}
		images := extractMarkdownImages(xianyuAwesomeGptImage2RawBase, block)
		cover := ""
		if len(images) > 0 {
			cover = images[0]
		}
		items = append(items, model.Prompt{ID: "xianyu-awesome-gptimage2-" + leftPad(len(items)+1), Title: currentTitle, CoverURL: cover, Prompt: prompt, Tags: xianyuPromptTags(currentCategory), Preview: markdownPreview(images)})
	}
	for _, line := range strings.Split(section, "\n") {
		if strings.HasPrefix(line, "### ") && !strings.HasPrefix(line, "#### ") {
			finish()
			currentTitle = ""
			currentLines = []string{}
			currentCategory = cleanXianyuCategory(strings.TrimSpace(strings.TrimPrefix(line, "### ")))
			continue
		}
		if strings.HasPrefix(line, "#### ") {
			finish()
			currentTitle = cleanXianyuPromptTitle(strings.TrimSpace(strings.TrimPrefix(line, "#### ")))
			currentLines = []string{}
			continue
		}
		if currentTitle != "" {
			currentLines = append(currentLines, line)
		}
	}
	finish()
	return items
}

func markdownSection(markdown, startHeading, endHeading string) string {
	start := strings.Index(markdown, startHeading)
	if start < 0 {
		return ""
	}
	rest := markdown[start+len(startHeading):]
	end := strings.Index(rest, endHeading)
	if end < 0 {
		return markdown[start:]
	}
	return markdown[start : start+len(startHeading)+end]
}

func cleanXianyuCategory(value string) string {
	value = strings.TrimSpace(value)
	for _, sep := range []string{"、", ".", "．", " "} {
		if index := strings.Index(value, sep); index >= 0 {
			prefix := strings.TrimSpace(value[:index])
			if prefix != "" && len([]rune(prefix)) <= 4 {
				value = strings.TrimSpace(value[index+len(sep):])
			}
			break
		}
	}
	return value
}

func cleanXianyuPromptTitle(value string) string {
	value = strings.TrimSpace(value)
	if index := strings.Index(value, " "); index > 0 {
		prefix := value[:index]
		if strings.Contains(prefix, ".") || strings.Contains(prefix, "．") {
			value = strings.TrimSpace(value[index+1:])
		}
	}
	return value
}

func xianyuCodeBlockText(block string) string {
	lines := []string{}
	inCode := false
	for _, line := range strings.Split(block, "\n") {
		text := strings.TrimSpace(line)
		if strings.HasPrefix(text, "```") {
			if inCode {
				break
			}
			inCode = true
			continue
		}
		if inCode {
			lines = append(lines, line)
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func xianyuFallbackPromptText(block string) string {
	lines := []string{}
	for _, line := range strings.Split(block, "\n") {
		text := strings.TrimSpace(line)
		if text == "" || strings.HasPrefix(text, "#") || strings.HasPrefix(text, "---") || strings.HasPrefix(text, "![") || strings.HasPrefix(text, "|") || strings.HasPrefix(text, ">") || strings.HasPrefix(text, "```") {
			continue
		}
		if strings.HasPrefix(text, "- 原文链接") || strings.HasPrefix(text, "- 公众号") || strings.HasPrefix(text, "- 作者") || strings.HasPrefix(text, "- 本次补充") || strings.HasPrefix(text, "- 说明") {
			continue
		}
		text = strings.TrimSpace(strings.TrimPrefix(text, "-"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "*"))
		if strings.HasPrefix(text, "提示词：") {
			text = strings.TrimSpace(strings.TrimPrefix(text, "提示词："))
		}
		if text != "" && !strings.HasPrefix(text, "http") {
			lines = append(lines, text)
		}
	}
	return strings.Join(lines, "\n")
}

func xianyuPromptTags(category string) []string {
	tags := []string{"gpt-image-2"}
	if category != "" {
		tags = append(tags, splitTags(category, "\\s*(/|&|、|与)\\s*")...)
	}
	return tags
}

func buildXianyuLatestXPrompts(offset int) ([]model.Prompt, error) {
	raw, err := fetchText(xianyuAwesomeGptImage2RawBase, "data/latest-prompts.json")
	if err != nil {
		return nil, err
	}
	data := xianyuLatestPromptData{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}
	items := []model.Prompt{}
	seen := map[string]bool{}
	appendOne := func(item xianyuLatestPrompt) {
		prompt := strings.TrimSpace(item.Prompt)
		if prompt == "" {
			return
		}
		key := firstXianyuNonEmpty(item.XURL, item.URL, item.Author+item.CreatedAt+prompt)
		if seen[key] {
			return
		}
		seen[key] = true
		image := firstXianyuNonEmpty(item.PrimaryImageURL, firstString(item.ImageURLs))
		title := firstXianyuNonEmpty(item.Reason, item.Author, "X Prompt")
		preview := xianyuLatestXPreview(item, image)
		date := normalizePromptTime(item.CreatedAt)
		items = append(items, model.Prompt{ID: "xianyu-awesome-gptimage2-" + leftPad(offset+len(items)+1), Title: title, CoverURL: image, Prompt: prompt, Tags: []string{"x"}, CreatedAt: date, UpdatedAt: date, Preview: preview})
	}
	for _, group := range data.Dates {
		for _, item := range group.Items {
			appendOne(item)
		}
	}
	for _, item := range data.Items {
		appendOne(item)
	}
	return items, nil
}

func xianyuLatestXPreview(item xianyuLatestPrompt, image string) string {
	lines := []string{}
	link := firstXianyuNonEmpty(item.XURL, item.URL)
	if link != "" {
		lines = append(lines, link)
	}
	for _, url := range item.ImageURLs {
		if url = strings.TrimSpace(url); url != "" {
			lines = append(lines, url)
		}
	}
	if len(lines) == 1 && image != "" {
		lines = append(lines, image)
	}
	return strings.Join(lines, "\n")
}

func normalizePromptTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, time.RFC1123, time.RFC1123Z, "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format(time.RFC3339)
		}
	}
	return value
}

func firstXianyuNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func firstString(values []string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func buildYouMindGptImage2Prompts() ([]model.Prompt, error) {
	return buildYouMindPrompts(youMindGptImage2RawBase, "youmind-gpt-image-2", "gpt-image-2")
}

func buildYouMindNanoBananaProPrompts() ([]model.Prompt, error) {
	return buildYouMindPrompts(youMindNanoBananaProRawBase, "youmind-nano-banana-pro", "nano-banana-pro")
}

func buildDavidWuGptImage2Prompts() ([]model.Prompt, error) {
	raw, err := fetchText(davidWuGptImage2RawBase, "prompts.json")
	if err != nil {
		return nil, err
	}
	data := []davidWuGptImage2Prompt{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}
	items := []model.Prompt{}
	for _, item := range data {
		title := strings.TrimSpace(item.TitleCN)
		if title == "" {
			title = strings.TrimSpace(item.TitleEN)
		}
		prompt := strings.TrimSpace(item.Prompt)
		if title == "" || prompt == "" {
			continue
		}
		image := absoluteImage(davidWuGptImage2RawBase, item.Image)
		items = append(items, model.Prompt{ID: "davidwu-gpt-image2-prompts-" + leftPad(item.ID), Title: title, CoverURL: image, Prompt: prompt, Tags: davidWuGptImage2Tags(item), Preview: davidWuGptImage2Preview(item, image)})
	}
	return items, nil
}

func buildYouMindPrompts(baseURL, idPrefix, modelTag string) ([]model.Prompt, error) {
	markdown, err := fetchText(baseURL, "README_zh.md")
	if err != nil {
		return nil, err
	}
	items := []model.Prompt{}
	for _, block := range splitBeforeHeading(markdown, "### ") {
		title := strings.TrimSpace(firstMatch(block, `(?m)^###\s+No\.\s*\d+:\s*(.+)$`))
		prompt := strings.TrimSpace(firstMatch(block, "(?s)#### .*?提示词\\s*\\r?\\n\\s*```[\\w-]*\\r?\\n(.*?)\\r?\\n```"))
		if title == "" || prompt == "" {
			continue
		}
		images := extractMarkdownImages(baseURL, block)
		cover := ""
		if len(images) > 0 {
			cover = images[0]
		}
		items = append(items, model.Prompt{ID: idPrefix + "-" + leftPad(len(items)+1), Title: title, CoverURL: cover, Prompt: prompt, Tags: youMindTags(title, modelTag), Preview: markdownPreview(images)})
	}
	return items, nil
}

func splitBeforeHeading(markdown string, prefix string) []string {
	blocks := []string{}
	lines := strings.Split(markdown, "\n")
	current := []string{}
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) && len(current) > 0 {
			blocks = append(blocks, strings.Join(current, "\n"))
			current = []string{}
		}
		current = append(current, line)
	}
	return append(blocks, strings.Join(current, "\n"))
}

func firstMatch(value string, pattern string) string {
	match := regexp.MustCompile(pattern).FindStringSubmatch(value)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func tagsFromCategory(category string) []string {
	return splitTags(regexp.MustCompile(`(?i)\s+Cases$`).ReplaceAllString(category, ""), `\s*(&|and)\s*`)
}

func tagsFromHeading(heading string) []string {
	return splitTags(regexp.MustCompile(`[^\p{L}\p{N}/&、与 ]`).ReplaceAllString(heading, ""), `\s*(/|&|、|与)\s*`)
}

func youMindTags(title, modelTag string) []string {
	tags := []string{modelTag}
	parts := strings.SplitN(title, " - ", 2)
	if len(parts) > 1 {
		tags = append(tags, tagsFromHeading(parts[0])...)
	}
	return tags
}

func davidWuGptImage2Tags(item davidWuGptImage2Prompt) []string {
	tags := splitTags(strings.Join([]string{item.CategoryCN, item.Category, item.Author, item.Source}, "/"), `/`)
	if item.NeedsRef {
		tags = append(tags, "需要参考图")
	}
	return tags
}

func davidWuGptImage2Preview(item davidWuGptImage2Prompt, image string) string {
	lines := []string{}
	if item.TitleEN != "" {
		lines = append(lines, item.TitleEN)
	}
	if item.Note != "" {
		lines = append(lines, item.Note)
	}
	if image != "" {
		lines = append(lines, "![]("+image+")")
	}
	return strings.Join(lines, "\n\n")
}

func splitTags(value string, pattern string) []string {
	tags := []string{}
	for _, tag := range regexp.MustCompile(pattern).Split(value, -1) {
		if tag = strings.ToLower(strings.TrimSpace(tag)); tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func markdownPreview(images []string) string {
	lines := []string{}
	for _, image := range images {
		if image != "" {
			lines = append(lines, "![]("+image+")")
		}
	}
	return strings.Join(lines, "\n\n")
}

func extractMarkdownImages(baseURL string, block string) []string {
	seen := map[string]bool{}
	images := []string{}
	for _, pattern := range []string{`<img[^>]+src="([^"]+)"`, `!\[[^\]]*]\(([^)]+)\)`} {
		for _, match := range regexp.MustCompile(pattern).FindAllStringSubmatch(block, -1) {
			image := absoluteImage(baseURL, match[1])
			if image != "" && !seen[image] {
				seen[image] = true
				images = append(images, image)
			}
		}
	}
	return images
}

func absoluteImage(baseURL, image string) string {
	if image == "" || strings.HasPrefix(image, "http://") || strings.HasPrefix(image, "https://") {
		return image
	}
	return baseURL + "/" + strings.TrimLeft(strings.TrimPrefix(image, "."), "/")
}

func leftPad(value int) string {
	if value >= 1000 {
		return strconv.Itoa(value)
	}
	text := "000" + strconv.Itoa(value)
	return text[len(text)-3:]
}

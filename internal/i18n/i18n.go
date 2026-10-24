// Package i18n localizes human-facing CLI text; protocol/JSON payloads bypass it.
package i18n

import (
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

var formatPattern = regexp.MustCompile(`%[sdvw]`)

func Normalize(language string) string {
	language = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(language), "_", "-"))
	if language == "zh" || strings.HasPrefix(language, "zh-") {
		return "zh"
	}
	return "en"
}

func Language() string {
	if override := os.Getenv("REMOTE_AGENT_LANGUAGE"); override != "" {
		return Normalize(override)
	}
	return installedLanguage()
}

func Text(english string) string {
	if Language() == "zh" {
		if translated, ok := messages[english]; ok {
			return translated
		}
	}
	return english
}

type pattern struct {
	re               *regexp.Regexp
	english, chinese string
}

var patterns = func() []pattern {
	keys := make([]string, 0, len(messages))
	for english := range messages {
		keys = append(keys, english)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	var result []pattern
	for _, english := range keys {
		if !formatPattern.MatchString(english) {
			continue
		}
		parts := formatPattern.Split(english, -1)
		expression := "(?s)^" + regexp.QuoteMeta(parts[0])
		for _, part := range parts[1:] {
			expression += "(.*?)" + regexp.QuoteMeta(part)
		}
		result = append(result, pattern{regexp.MustCompile(expression + "$"), english, messages[english]})
	}
	return result
}()

func translate(message string, depth int) string {
	if translated, ok := messages[message]; ok {
		return translated
	}
	if depth > 8 {
		return message
	}
	for _, candidate := range patterns {
		match := candidate.re.FindStringSubmatch(message)
		if match == nil {
			continue
		}
		formats := formatPattern.FindAllString(candidate.english, -1)
		index := 0
		return formatPattern.ReplaceAllStringFunc(candidate.chinese, func(string) string {
			index++
			value := match[index]
			if formats[index-1] == "%w" || formats[index-1] == "%v" {
				value = translate(value, depth+1)
			} else if translated, ok := messages[value]; ok {
				value = translated
			}
			return value
		})
	}
	return message
}

func Error(err error) string {
	if err == nil {
		return ""
	}
	if Language() != "zh" {
		return err.Error()
	}
	return translate(err.Error(), 0)
}

type writer struct{ output io.Writer }

func Writer(output io.Writer) io.Writer { return writer{output} }
func (w writer) Write(data []byte) (int, error) {
	message := string(data)
	if Language() == "zh" {
		message = translate(message, 0)
		if message == string(data) {
			var lines []string
			for _, line := range strings.SplitAfter(message, "\n") {
				translated := translate(line, 0)
				if translated == line {
					translated = translate(strings.TrimSuffix(line, "\n"), 0)
					if strings.HasSuffix(line, "\n") {
						translated += "\n"
					}
				}
				lines = append(lines, translated)
			}
			message = strings.Join(lines, "")
		}
	}
	_, err := io.WriteString(w.output, message)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

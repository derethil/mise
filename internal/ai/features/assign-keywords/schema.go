package assignkeywords

import "strings"

const sectionMarker = "## "

type section struct {
	Title string
	Body  string
}

func (s section) label() string {
	if s.Title == "" {
		return "keywords"
	}

	return s.Title
}

func splitSchema(schema string) []section {
	lines := strings.Split(schema, "\n")

	var preamble []string
	var sections []section

	for _, line := range lines {
		title, isHeading := strings.CutPrefix(line, sectionMarker)
		if isHeading {
			sections = append(sections, section{Title: strings.TrimSpace(title)})
			continue
		}

		if len(sections) == 0 {
			preamble = append(preamble, line)
			continue
		}

		last := &sections[len(sections)-1]
		last.Body += line + "\n"
	}

	shared := strings.TrimSpace(strings.Join(preamble, "\n"))

	if len(sections) == 0 {
		if shared == "" {
			return nil
		}
		return []section{{Body: shared}}
	}

	for i := range sections {
		sections[i].Body = strings.TrimSpace(sections[i].Body)
		if shared != "" {
			sections[i].Body = shared + "\n\n" + sections[i].Body
		}
	}

	return sections
}

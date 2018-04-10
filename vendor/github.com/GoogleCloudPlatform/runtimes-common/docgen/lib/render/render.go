package render

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/proto"
	"github.com/GoogleCloudPlatform/runtimes-common/docgen/lib/render/templates"
)

var predefinedAnchorIds = [...]string{
	"about",
	"other-versions",
	"references",
	"references-environment-variables",
	"references-ports",
	"references-volumes",
	"table-of-contents",
	"using-docker",
	"using-kubernetes",
}

func Render(document *proto.Document) ([]byte, error) {
	document = generateAnchorIds(document)
	tmpl, err := template.New("README").Parse(templates.Readme)
	check(err)
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, NewDocument(document))
	return buf.Bytes(), err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func generateAnchorIds(doc *proto.Document) *proto.Document {
	// Maps from a base anchor ID to the number of anchors with the same base ID.
	// Anchors sharing a base ID will differ only in the suffixes, e.g. "-1".
	anchorCounts := make(map[string]int)
	for _, id := range predefinedAnchorIds {
		anchorCounts[id] = 1
	}

	result := &proto.Document{}
	*result = *doc
	for _, taskGroup := range result.TaskGroups {
		if len(taskGroup.AnchorId) > 0 {
			// Handles a user-defined anchor ID. This ID must be unique.
			if _, found := anchorCounts[taskGroup.AnchorId]; found {
				panic(fmt.Sprintf("Duplicate anchor: %s", taskGroup.AnchorId))
			}
		} else {
			// Generates an anchor ID from the title.
			desiredId := makeAnchorId(taskGroup.Title)
			taskGroup.AnchorId = updateAnchorId(anchorCounts, desiredId)
		}

		for _, task := range taskGroup.Tasks {
			if len(task.AnchorId) > 0 {
				// Handles a user-defined anchor ID. This ID must be unique.
				if _, found := anchorCounts[task.AnchorId]; found {
					panic(fmt.Sprintf("Duplicate anchor: %s", task.AnchorId))
				}
			} else {
				// Generates an anchor ID from the title.
				desiredId := makeAnchorId(task.Title)
				task.AnchorId = updateAnchorId(anchorCounts, desiredId)
			}
		}
	}
	return result
}

func updateAnchorId(anchorCounts map[string]int, desiredId string) string {
	result := desiredId
	if count, ok := anchorCounts[desiredId]; ok {
		result = fmt.Sprintf("%s-%d", desiredId, count)
		anchorCounts[desiredId] = count + 1
	} else {
		anchorCounts[desiredId] = 1
	}
	return result
}

func makeAnchorId(title string) string {
	if len(title) == 0 {
		panic("Title cannot be empty")
	}
	r := regexp.MustCompile("[^A-Za-z0-9]+")
	safe := r.ReplaceAllString(title, "-")
	safe = strings.ToLower(strings.Trim(safe, "-"))
	return safe
}

// Package scripts Backlink is a small script which is meant to insert backlinks into our notes
// You can run it manually on your knowledge base, or you can run it periodically
// Should be run with working directory set to your root knowledge base
// WARNING! Cases with "|" in urls aren't handled yet, so duplicate urls possible
package scripts

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/exp/slices"
	"golang.org/x/text/unicode/norm"

	"zakirullin/stuffbot/internal/fs"
)

func main() {
	// [dir][note] => links referring to our note (backlinks)
	backlinks := make(map[string]map[string][]string)

	userFS, err := fs.NewFS(".", afero.NewOsFs())
	if err != nil {
		fmt.Printf("Can't create FS: %s", err)
		return
	}

	files, err := userFS.FilesAndDirs("")
	if err != nil {
		fmt.Printf("Can't get files and dirs: %s", err)
		return
	}
	dirs := fs.OnlyNoteDirs(fs.OnlyDirs(files))
	for _, dir := range dirs {
		notes, err := userFS.FilesAndDirs(dir.Name)
		if err != nil {
			fmt.Printf("Can't get notes: %s", err)
		}

		notes = fs.OnlyMDFiles(notes)
		for _, note := range notes {
			if filepath.Ext(note.Name) != ".md" {
				continue
			}

			content, err := userFS.Read(dir.Name, note.Name)
			if err != nil {
				fmt.Printf("Can't get content: %s", err)
				return
			}

			links := regexp.MustCompile(`\[\[(.*?)\]\]`)
			matches := links.FindAllStringSubmatch(content, -1)
			for _, match := range matches {
				if len(match) < 2 {
					continue
				}

				link := match[1]
				if strings.Contains(link, "/img/") {
					continue
				}

				parts := strings.Split(link, "/")
				isInAnotherDir := len(parts) > 2

				targetDir := dir.Name
				targetNote := parts[0]
				link = strings.TrimSuffix(note.Name, ".md")
				// There are issues with "й" letter. Probably it has non-canonical encoding in mac FS
				link = string(norm.NFC.Bytes([]byte(link)))
				if isInAnotherDir {
					targetDir = parts[1]
					targetNote = parts[2]

					link = fmt.Sprintf("../%s/%s", dir.Name, link)
				}
				targetNote = strings.Split(targetNote, "|")[0]

				if _, ok := backlinks[targetDir]; !ok {
					backlinks[targetDir] = make(map[string][]string)
				}

				backlinks[targetDir][targetNote] = append(backlinks[targetDir][targetNote], link)
			}
		}
	}

	for dir, notes := range backlinks {
		for note, links := range notes {
			for _, link := range links {
				content, err := userFS.Read(dir, note+".md")
				if err != nil {
					fmt.Printf("Can't get target note content: %s, backlinks: %v", err, links)
					return
				}
				existingLinksRx := regexp.MustCompile(`\[\[(.*)\]\]`)
				matches := existingLinksRx.FindAllStringSubmatch(content, -1)
				var existingLinks []string
				for _, match := range matches {
					if len(match) < 2 {
						continue
					}
					existingLinks = append(existingLinks, string(match[1]))
				}

				if slices.Contains(existingLinks, link) {
					continue
				}

				err = userFS.Write(dir, note+".md", fmt.Sprintf("%s\n[[%s]]", strings.TrimSpace(content), link))
				if err != nil {
					fmt.Printf("Can't put to file: %s", err)
					return
				}

				fmt.Printf("Appending %s to %s/%s\n", link, dir, note)
			}
		}
	}
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	blackfriday "v/github.com/russross/blackfriday@v1.5.1"
)

type titles struct {
	t1 string
	t2 string
}

type siteData struct {
	dir         string
	sections    []titles
	subSections []titles

	titleLeads    []string
	titlePrefixes []string
	titleSuffixes []string
}

type testpageBuilder struct {
	dir       string
	startDate time.Time
	seen      map[string]bool
	titleFn   func(s string) string
}

func newTestpageBuilder(dir string) *testpageBuilder {
	startDate, _ := time.Parse("2006-01-02", "2017-01-01")
	return &testpageBuilder{dir: dir, startDate: startDate, seen: make(map[string]bool)}
}

func (t *testpageBuilder) createPage(title, linkTitle string, i int) string {
	title = t.titleFn(title)
	linkTitle = t.titleFn(linkTitle)
	return fmt.Sprintf(pageTemplate, title, linkTitle, t.startDate.Add(time.Duration(i*24)*time.Hour).Format("2006-01-02"))
}

func (t *testpageBuilder) createSectionPage(title, linkTitle string, i int) string {
	title = t.titleFn(title)
	linkTitle = t.titleFn(linkTitle)
	return fmt.Sprintf(sectionTemplate, title, linkTitle, t.startDate.Add(time.Duration(i*24)*time.Hour).Format("2006-01-02"))
}

func (t *testpageBuilder) createMainSectionPage(title, linkTitle string, i int) string {
	title = t.titleFn(title)
	linkTitle = t.titleFn(linkTitle)
	return fmt.Sprintf(mainSectionTemplate, title, linkTitle, i, i)
}

func (t *testpageBuilder) buildSections() error {
	var enData = siteData{
		dir:           "content/en/docs",
		sections:      []titles{titles{"Big Data", "Big Data in Small Disks"}, titles{"API Reference", "Full API Reference"}, titles{"Cloud Computing", "Computing in the cloud"}, titles{"Content Management", "Content Management"}, titles{"Cross-Platform", "Cross-Platform"}},
		subSections:   []titles{titles{"Examples", "Examples in Code"}, titles{"Tutorials", "Step by step tutorials"}},
		titleLeads:    []string{"In depth", "The Math of", "The inside of"},
		titlePrefixes: []string{"Recursion", "Cryptography", "Monographs", "Java", "Go", "Monoliths", "Microservices"},
		titleSuffixes: []string{"The Core Concepts", "How does it work?", "The Inner Workings", "Detailed Spec"},
	}

	var noData = siteData{
		dir:           "content/no/docs",
		sections:      []titles{titles{"Big Data", "Store mengder data"}, titles{"API-referanse", "Komplett API-referance"}, titles{"Sky-data", "Data i skyen"}, titles{"Innholdshåndtering", "Håndtering av innhold"}, titles{"Flerplattform", "Flere plattformer"}},
		subSections:   []titles{titles{"Eksempler", "Praktiske eksempler"}, titles{"Hjelpeartikler", "Steg for steg hjelpeartikler"}},
		titleLeads:    []string{"I dybden på", "Matematikken til", "På innsiden av"},
		titlePrefixes: []string{"rekursjon", "kryptografi", "monografer", "Java", "Go", "monolitter", "mikroservicer"},
		titleSuffixes: []string{"De grunnleggende konsepten", "Hvordan virker det?", "Drivverket", "Detaljert spefisikasjon"},
	}

	t.titleFn = strings.Title

	docSectionContent := t.createMainSectionPage("TechOS Documentation", "Documentation", 20)
	if err := t.writeContent(docSectionContent, filepath.Join(enData.dir, "_index.md")); err != nil {
		return err
	}

	docSectionContent = t.createMainSectionPage("TechOS-dokumentasjon", "Dokumentasjon", 20)
	if err := t.writeContent(docSectionContent, filepath.Join(noData.dir, "_index.md")); err != nil {
		return err
	}

	if err := t.buildSectionsFor(enData, enData); err != nil {
		return err
	}

	t.titleFn = func(s string) string {
		return s

	}
	return t.buildSectionsFor(enData, noData)
}

func (t *testpageBuilder) buildSectionsFor(master, data siteData) error {
	rand.Seed(int64(32))
	for i, sect := range data.sections {
		masterSect := master.sections[i]
		dir := blackfriday.SanitizedAnchorName(masterSect.t1)
		dir = filepath.Join(data.dir, dir)
		pageContent := t.createSectionPage(sect.t2, sect.t1, i)
		if err := t.writeContent(pageContent, filepath.Join(dir, "_index.md")); err != nil {
			return err
		}

		must(t.buildPages(master, data, dir, sect, i))

		numSubSections := rand.Intn(len(data.subSections))
		for j := 0; j <= numSubSections; j++ {
			si := rand.Intn(len(data.subSections))
			masterSubSect := master.subSections[si]
			subsect := data.subSections[si]
			subdir := blackfriday.SanitizedAnchorName(masterSubSect.t1)
			subdir = filepath.Join(dir, subdir)
			pageContent := t.createSectionPage(subsect.t2, subsect.t1, i+j)
			must(t.buildPages(master, data, subdir, subsect, i+j))
			indexFilename := filepath.Join(subdir, "_index.md")
			if err := t.writeContent(pageContent, indexFilename); err != nil {
				return err
			}

		}
	}

	return nil
}

func (t *testpageBuilder) buildPages(master, data siteData, dir string, sect titles, i int) error {
	if t.seen[dir] {
		return nil
	}
	t.seen[dir] = true

	numPagesInSection := rand.Intn(len(data.titlePrefixes))
	for j := 0; j <= numPagesInSection; j++ {
		i1, i2, i3 := rand.Intn(len(data.titleLeads)), rand.Intn(len(data.titlePrefixes)), rand.Intn(len(data.titleSuffixes))
		linkTitle := data.titleLeads[i1] + " " + data.titlePrefixes[i2]
		title := linkTitle + ": " + data.titleSuffixes[i3]
		name := blackfriday.SanitizedAnchorName(master.titleLeads[i1] + " " + master.titlePrefixes[i2] + ":" + master.titleSuffixes[i3])
		pageContent := t.createPage(title, linkTitle, i+j)
		if err := t.writeContent(pageContent, filepath.Join(dir, fmt.Sprintf("%s.md", name))); err != nil {
			return err
		}
	}

	return nil

}

func (t *testpageBuilder) writeContent(content, name string) error {
	filename := filepath.Join(t.dir, name)
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(filename, []byte(content), 0755)
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Building test content in", dir)
	//must(os.RemoveAll(filepath.Join(dir, docsDir)))
	builder := newTestpageBuilder(dir)
	must(builder.buildSections())

}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

const (
	sectionTemplate = `
---
title: %q
linkTitle: %q
date: %s
description: >
  A short lead descripton about this section page. Text here can also be **bold** or _italic_ and can even be split over multiple paragraphs.
---

This is the section landing page.

* Summarize
* Your section
* Here

`

	mainSectionTemplate = `
---
title: %q
linkTitle: %q
weight: %d
menu:
  main:
    weight: %d
---

This is a landing page for a top level section.

* Summarize
* Your section
* Here


`
	pageTemplate = `
---
title: %q
linkTitle: %q
date: %s
description: >
  A short lead descripton about this content page. Text here can also be **bold** or _italic_ and can even be split over multiple paragraphs.
---

Text can be **bold**, _italic_, or ~~strikethrough~~. [Links](https://github.com) should be blue with no underlines (unless hovered over).

There should be whitespace between paragraphs. There should be whitespace between paragraphs. There should be whitespace between paragraphs. There should be whitespace between paragraphs.

There should be whitespace between paragraphs. There should be whitespace between paragraphs. There should be whitespace between paragraphs. There should be whitespace between paragraphs.

> There should be no margin above this first sentence.
>
> Blockquotes should be a lighter gray with a border along the left side in the secondary color.
>
> There should be no margin below this final sentence.

## First Header

This is a normal paragraph following a header. Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.  Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.  Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.



Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

On big screens, paragraphs and headings should not take up the full container width, but we want tables, code blocks and similar to take the full width.

Lorem markdownum tuta hospes stabat; idem saxum facit quaterque repetito
occumbere, oves novem gestit haerebat frena; qui. Respicit recurvam erat:
pignora hinc reppulit nos **aut**, aptos, ipsa.

Meae optatos *passa est* Epiros utiliter *Talibus niveis*, hoc lata, edidit.
Dixi ad aestum.

## Header 2

> This is a blockquote following a header. Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### Header 3

` + "```" + `
This is a code block following a header.
` + "```" + `

#### Header 4

* This is an unordered list following a header.
* This is an unordered list following a header.
* This is an unordered list following a header.

##### Header 5

1. This is an ordered list following a header.
2. This is an ordered list following a header.
3. This is an ordered list following a header.

###### Header 6

| What      | Follows         |
|-----------|-----------------|
| A table   | A header        |
| A table   | A header        |
| A table   | A header        |

----------------

There's a horizontal rule above and below this.

----------------

Here is an unordered list:

* Salt-n-Pepa
* Bel Biv DeVoe
* Kid 'N Play

And an ordered list:

1. Michael Jackson
2. Michael Bolton
3. Michael Bublé

And an unordered task list:

- [x] Create a sample markdown document
- [x] Add task lists to it
- [ ] Take a vacation

And a "mixed" task list:

- [ ] Steal underpants
- ?
- [ ] Profit!

And a nested list:

* Jackson 5
  * Michael
  * Tito
  * Jackie
  * Marlon
  * Jermaine
* TMNT
  * Leonardo
  * Michelangelo
  * Donatello
  * Raphael

Definition lists can be used with Markdown syntax. Definition terms are bold.

Name
: Godzilla

Born
: 1952

Birthplace
: Japan

Color
: Green


----------------

Tables should have bold headings and alternating shaded rows.

| Artist            | Album           | Year |
|-------------------|-----------------|------|
| Michael Jackson   | Thriller        | 1982 |
| Prince            | Purple Rain     | 1984 |
| Beastie Boys      | License to Ill  | 1986 |

If a table is too wide, it should scroll horizontally.

| Artist            | Album           | Year | Label       | Awards   | Songs     |
|-------------------|-----------------|------|-------------|----------|-----------|
| Michael Jackson   | Thriller        | 1982 | Epic Records | Grammy Award for Album of the Year, American Music Award for Favorite Pop/Rock Album, American Music Award for Favorite Soul/R&B Album, Brit Award for Best Selling Album, Grammy Award for Best Engineered Album, Non-Classical | Wanna Be Startin' Somethin', Baby Be Mine, The Girl Is Mine, Thriller, Beat It, Billie Jean, Human Nature, P.Y.T. (Pretty Young Thing), The Lady in My Life |
| Prince            | Purple Rain     | 1984 | Warner Brothers Records | Grammy Award for Best Score Soundtrack for Visual Media, American Music Award for Favorite Pop/Rock Album, American Music Award for Favorite Soul/R&B Album, Brit Award for Best Soundtrack/Cast Recording, Grammy Award for Best Rock Performance by a Duo or Group with Vocal | Let's Go Crazy, Take Me With U, The Beautiful Ones, Computer Blue, Darling Nikki, When Doves Cry, I Would Die 4 U, Baby I'm a Star, Purple Rain |
| Beastie Boys      | License to Ill  | 1986 | Mercury Records | noawardsbutthistablecelliswide | Rhymin & Stealin, The New Style, She's Crafty, Posse in Effect, Slow Ride, Girls, (You Gotta) Fight for Your Right, No Sleep Till Brooklyn, Paul Revere, Hold It Now, Hit It, Brass Monkey, Slow and Low, Time to Get Ill |

----------------

Code snippets like ` + "`" + `var foo = "bar";` + "`" + ` can be shown inline.

Also, ` + "`" + `this should vertically align` + "`" + ` ~~` + "`" + `with this` + "`" + `~~ ~~and this~~.

Code can also be shown in a block element.

` + "```" + `
foo := "bar";
bar := "foo";
` + "```" + `

Code can also use syntax highlighting.

` + "```" + `go
func main() {
  input := ` + "`" + `var foo = "bar";` + "`" + `

  lexer := lexers.Get("javascript")
  iterator, _ := lexer.Tokenise(nil, input)
  style := styles.Get("github")
  formatter := html.New(html.WithLineNumbers())

  var buff bytes.Buffer
  formatter.Format(&buff, style, iterator)

  fmt.Println(buff.String())
}
` + "```" + `

` + "```" + `
Long, single-line code blocks should not wrap. They should horizontally scroll if they are too long. This line should be long enough to demonstrate this.
` + "```" + `

Inline code inside table cells should still be distinguishable.

| Language    | Code               |
|-------------|--------------------|
| Javascript  | ` + "`" + `var foo = "bar";` + "`" + ` |
| Ruby        | ` + "`" + `foo = "bar"{` + "`" + `      |

----------------

Small images should be shown at their actual size.

![](http://placekitten.com/g/300/200/)

Large images should always scale down and fit in the content container.

![](http://placekitten.com/g/1200/800/)

## Components

### Alerts

{{< alert >}}This is an alert.{{< /alert >}}
{{< alert title="Note:" >}}This is an alert with a title.{{< /alert >}}
{{< alert type="success" >}}This is a successful alert.{{< /alert >}}
{{< alert type="warning" >}}This is a warning!{{< /alert >}}
{{< alert type="warning" title="Warning!" >}}This is a warning with a title!{{< /alert >}}


## Sizing

Add some sections here to see how the ToC looks like. Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### Parameters available

Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### Using pixels

Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### Using rem

Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

## Memory

Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### RAM to use

Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### More is better

Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### Used RAM

Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.



` + "```" + `
This is the final element on the page and there should be no margin below this.
` + "```" + `
`
)

package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/jay9596/goRant"
	"github.com/rivo/tview"
	"strings"
)

var c *goRant.Client
var skip int
var mode string
var rants *[]goRant.Rant

func main() {
	c := goRant.New()
	app := tview.NewApplication()
	mode = "algo"
	skip = 0
	r := LoadRants(mode, 5, skip)
	rants = &r

	pages := tview.NewPages()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	textView.SetTitle("[yellow] NAVIGATION: Next Rant(TAB) ; Previous Rant(SHIFT+TAB) ; Enter Rant (ENTER) ; Quit (ESC) [yellow]")
	textView.Highlight()

	rantView := 0
	rantView = DrawTextView(textView, rantView)

	textView.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			// Redraw rant to avoid comment duplication
			rantView = DrawTextView(textView, rantView)
			r := *rants
			_, comments, _ := c.GetRant(r[rantView].ID)

			fmt.Fprint(textView, " \n\n")

			chs := make([]chan string, len(comments))
			for i, _ := range chs {
				chs[i] = make(chan string)
			}

			for index, comment := range comments {
				go func(index int, comment goRant.Comment) {
					user, _ := c.Profile(comment.Username)

					com := fmt.Sprintf("[red]%d++ ", comment.Score)
					com += fmt.Sprintf("[yellow]%s[yellow] [white]+%d[white] ", comment.Username, user.Score)

					if user.DPP == 1 {
						com += fmt.Sprintln("[red]++[red] [white]:")
					} else {
						com += fmt.Sprintln("[white]:")
					}
					com += fmt.Sprintf("%s \n\n", comment.Body)

					chs[index] <- com

				}(index, comment)
			}

			for _, ch := range chs {
				fmt.Fprintf(textView, "%s", <-ch)
			}

			textView.ScrollToBeginning()

		} else if key == tcell.KeyTab {
			rantView += 1
			rantView = DrawTextView(textView, rantView)
		} else if key == tcell.KeyBacktab {
			rantView -= 1
			rantView = DrawTextView(textView, rantView)
		} else if key == tcell.KeyEscape {
			app.Stop()
		}
	})

	textView.SetBorder(true)

	pages.AddPage("name", textView, true, true)

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}

func DrawTextView(textView *tview.TextView, rantView int) int {
	textView.Clear()
	r := *rants
	if rantView < 0 {
		rantView = 0
	} else if rantView > len(r)-1 {
		skip += 1
		rArr := LoadRants(mode, 5, skip)
		r = append(r, rArr...)
		rants = &r
	}
	rant := r[rantView]

	fmt.Fprintf(textView, "Score: %d \n", rant.Score)
	if len(rant.AttachedImage.URL) > 0 {
		fmt.Fprintf(textView, "Image: %s \n", rant.AttachedImage.URL)
	}

	fmt.Fprintf(textView, "[yellow]Ranter: %s [yellow] \n\n", rant.Username)

	contWords := 1
	for _, word := range strings.Split(rant.Text, " ") {
		if contWords%10 == 0 {
			contWords += 1
			word = fmt.Sprintf("%s \n", word)
		} else {
			contWords += 1
			word = fmt.Sprintf("%s ", word)
		}

		fmt.Fprintf(textView, "%s", word)
	}
	fmt.Fprint(textView, "\n\n")

	var tags string
	for _, tag := range rant.Tags {
		tags += tag + ", "
	}
	tags = strings.TrimSuffix(tags, ", ")
	fmt.Fprintf(textView, "[yellow]Tags: [red]%s[red] \n", tags)
	fmt.Fprintf(textView, "Comments: %d \n", rant.NumComments)

	return rantView
}

func LoadRants(mode string, page, skip int) []goRant.Rant {
	r, err := c.Rants(mode, page, skip)
	if err != nil {
		panic(err)
	}
	return r
}

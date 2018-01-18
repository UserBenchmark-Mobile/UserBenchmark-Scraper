package main

import (
	"encoding/json"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

func (c CPU) Get(ctx context.Context, cdp *chromedp.CDP, url string) (Component, error) {
	var html string

	if err := GetOuterHTML(ctx, c, cdp, &html); err != nil {
		fmt.Println(err)
		return c, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		fmt.Println(err)
		return c, err
	}

	c.GetCores(*doc, ctx, cdp)
	c.GetAverages(*doc)
	c.GetSubResults(*doc, ctx, cdp)
	c.GetRelPerf(*doc, ctx, cdp)

	b, _ := json.MarshalIndent(c, "", "  ")

	fmt.Println(string(b))

	if old, ok := cpus.Get(c.Model); ok {
		if !c.IsValid(old) {
			return c, ErrNotValid
		}
	}

	return c, nil
}

func (c CPU) GetURL() string {
	return c.URL
}

func (c *CPU) GetCores(doc goquery.Document, ctx context.Context, cdp *chromedp.CDP) {
	err := cdp.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible(`.cmp-cpt.tallp.cmp-cpt-l`),
	})
	if err != nil {
		fmt.Println(err)
	}
	c.Cores = doc.Find(`.cmp-cpt.tallp.cmp-cpt-l`).Text()
}

func (c *CPU) GetAverages(doc goquery.Document) {
	for i := 0; i < 3; i++ {
		c.Scores[i] = doc.Find(fmt.Sprintf(`.para-m-t.uc-table.table-no-border > thead > tr > td:nth-child(%d) .mcs-caption.pgbg`, i+3)).Text()
		if c.Scores[i] == "" {
			c.Scores[i] = doc.Find(fmt.Sprintf(`.para-m-t.uc-table.table-no-border > thead > tr > td:nth-child(%d) .mcs-caption.pybg`, i+3)).Text()
			if c.Scores[i] == "" {
				c.Scores[i] = doc.Find(fmt.Sprintf(`.para-m-t.uc-table.table-no-border > thead > tr > td:nth-child(%d) .mcs-caption.prbg`, i+3)).Text()
			}
		}
	}
}

func (c *CPU) GetSubResults(doc goquery.Document, ctx context.Context, cdp *chromedp.CDP) {
	err := cdp.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible(`.mcs-hl-col`),
	})
	if err != nil {
		fmt.Println(err)
	}

	doc.Find(`.mcs-hl-col`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i == 8 {
			return false
		}
		c.SubResults[i] = s.Text()
		return true
	})
}

func (c *CPU) GetRelPerf(doc goquery.Document, ctx context.Context, cdp *chromedp.CDP) {
	err := cdp.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible(`.bsc-w.text-left.semi-strong`),
	})
	if err != nil {
		fmt.Println(err)
	}

	doc.Find(`.bsc-w.text-left.semi-strong div:first-child`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i == 3 {
			return false
		}
		c.SegmentPerf[i] = strings.TrimSpace(s.Text())
		return true
	})

	doc.Find(`.bsc-w.text-left.semi-strong div:nth-child(3)`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		if i == 3 {
			return false
		}
		c.SegmentPerf[i] += " "+s.Text()	
		return true
	})

	
}

func (c CPU) IsValid(old CPU) bool {
	switch {
	case c.Cores == "" && old.Cores != "":
		return false
	case !equallyEmpty(c.Scores[:], old.Scores[:]):
		return false
	case !equallyEmpty(c.SegmentPerf[:], old.SegmentPerf[:]):
		return false
	case !equallyEmpty(c.SubResults[:], old.SubResults[:]):
		return false
	default:
		return true
	}
}

func equallyEmpty(new, old []string) bool {
	for i, v := range new {
		if v == "" && old[i] != "" {
			return false
		}
	}
	return true
}

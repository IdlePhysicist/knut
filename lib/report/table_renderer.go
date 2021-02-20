// Copyright 2021 Silvio Böhler
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language govetring permissions and
// limitations under the License.

package report

import (
	"github.com/sboehler/knut/lib/model/accounts"
	"github.com/sboehler/knut/lib/model/commodities"
	"github.com/sboehler/knut/lib/table"
	"github.com/sboehler/knut/lib/vector"
)

// TableRenderer renders a report.
type TableRenderer struct {
	Commodities bool
	negate      bool
	report      *Report
	table       *table.Table
	indent      int
}

const indent = 2

// Render renders a report.
func (tr *TableRenderer) Render(r *Report) {

	tr.table = table.New(1, len(r.Dates))
	tr.indent = 0
	tr.report = r

	var render func(s *Segment)
	if tr.Commodities {
		render = tr.renderSegmentWithCommodities
	} else {
		render = tr.renderSegment
	}

	// sep
	tr.table.AddSeparatorRow()

	// header
	var header = tr.table.AddRow().AddText("Account", table.Center)
	for _, d := range r.Dates {
		header.AddText(d.Format("2006-01-02"), table.Center)
	}
	// sep
	tr.table.AddSeparatorRow()

	var g1, g2 []*Segment

	for _, at := range accounts.AccountTypes {
		s, ok := tr.report.Segments[at]
		if !ok {
			continue
		}
		if at == accounts.ASSETS || at == accounts.LIABILITIES {
			g1 = append(g1, s)
		} else {
			g2 = append(g2, s)
		}
	}

	// values
	if len(g1) > 0 {
		for _, s := range g1 {
			render(s)
			tr.table.AddEmptyRow()
		}

		var totals = make(map[*commodities.Commodity]vector.Vector)
		for _, s := range g1 {
			s.sum(totals)
		}
		render(&Segment{Key: "Total", Positions: totals})
		tr.table.AddSeparatorRow()

	}
	if len(g2) > 0 {
		tr.negate = true
		for _, s := range g2 {
			render(s)
			tr.table.AddEmptyRow()
		}
		var totals = make(map[*commodities.Commodity]vector.Vector)
		for _, s := range g2 {
			s.sum(totals)
		}
		render(&Segment{Key: "Total", Positions: totals})

		tr.negate = false
		tr.table.AddSeparatorRow()
	}

	// totals
	render(&Segment{
		Key:       "Delta",
		Positions: r.Positions,
	})
	tr.table.AddSeparatorRow()
}

func (tr *TableRenderer) GetTable() *table.Table {
	return tr.table
}

func (tr *TableRenderer) renderSegment(s *Segment) {
	// compute total value
	var total = vector.New(len(tr.report.Dates))
	for _, amounts := range s.Positions {
		total.Add(amounts)
	}

	// fill header cells with total values
	var header = tr.table.AddRow().AddIndented(s.Key, tr.indent)
	for _, amount := range total.Values {
		if amount.IsZero() {
			header.AddEmpty()
		} else {
			if tr.negate {
				amount = amount.Neg()
			}
			header.AddNumber(amount)
		}
	}

	// render subsegments
	tr.indent += indent
	for _, ss := range s.Subsegments {
		tr.renderSegment(ss)
	}
	tr.indent -= indent
}

func (tr *TableRenderer) renderSegmentWithCommodities(segment *Segment) {
	var header = tr.table.AddRow().AddIndented(segment.Key, tr.indent)
	for range tr.report.Dates {
		header.AddEmpty()
	}

	// add one row per commodity in this position
	tr.indent += indent
	for _, commodity := range tr.report.Commodities {
		if amounts, ok := segment.Positions[commodity]; ok {
			var row = tr.table.AddRow().AddIndented(commodity.String(), tr.indent)
			for _, amount := range amounts.Values {
				if amount.IsZero() {
					row.AddEmpty()
				} else {
					if tr.negate {
						amount = amount.Neg()
					}
					row.AddNumber(amount)
				}
			}
		}
	}

	// render subsegments
	for _, ss := range segment.Subsegments {
		tr.renderSegmentWithCommodities(ss)
	}
	tr.indent -= indent
}

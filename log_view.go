package main

import (
	"fmt"
	"strconv"
	"strings"

	ui "github.com/gizak/termui"
)

// LogViews is a list of viewnames
type LogViews []LogView

// LogView represents view into logs
type LogView struct {
	FileName string
	offSet   int
}
type initLogViews struct{}

// Display returns a Row object representing all of the logViews
func (logViews LogViews) display(state AppState) []*ui.Row {
	listBlocks := []*ui.List{}
	for i, view := range logViews {
		if i >= state.layout {
			break
		}
		logView := view.display(state)
		logView.BorderFg = ui.ColorWhite
		if state.wrap {
			logView.Overflow = "wrap"
		}
		listBlocks = append(listBlocks, logView)
	}
	if len(listBlocks) > 0 && (state.CurrentMode == normal || state.CurrentMode == selectCategory) {
		listBlocks[state.selected].BorderFg = ui.ColorMagenta
	}
	logViewColumns := []*ui.Row{}

	filterSize := 0
	if state.showMods {
		filterSize = getModifierSpan(state.termWidth)
	}
	numColumnsEach := (12 - filterSize) / state.layout
	leftOver := (12 - filterSize) - (numColumnsEach * state.layout)
	for _, logViewBlock := range listBlocks {
		extra := 0
		if leftOver > 0 {
			extra = 1
			leftOver--
		}
		logViewColumns = append(logViewColumns, ui.NewCol(numColumnsEach+extra, 0, logViewBlock))
	}
	return logViewColumns
}

func (view LogView) display(state AppState) *ui.List {
	list := ui.NewList()
	list.Height = logViewHeight(state.termHeight)
	list.BorderLabelFg = ui.ColorCyan
	active := state.getSelectedFileName() == view.FileName
	if active {
		list.BorderFg = ui.ColorWhite
	} else {
		list.BorderFg = ui.ColorYellow
	}
	list.BorderLabel = view.FileName
	file := state.Files[view.FileName]
	filteredLines := file.hlAndFiltered(state)
	height := view.numVisibleLines(state)
	visibleLines := filteredLines.getVisibleSlice(view, height)
	list.Items = visibleLines
	return list
}

func (view LogView) scrollToSearch(state AppState) LogView {
	file := state.Files[view.FileName]
	searchResultOffset := file.lines.getSearchResultLine(state.searchBuffer.text, state.searchIndex)
	if searchResultOffset >= 0 {
		view.offSet = searchResultOffset - (logViewHeight(state.termHeight) / 2)
		if view.offSet < 0 {
			view.offSet = 0
		}
	}
	return view
}

func (view LogView) numVisibleLines(state AppState) int {
	return logViewHeight(state.termHeight) - 2
}

func (lines lines) getVisibleSlice(view LogView, height int) []string {
	start := (len(lines) - height) - view.offSet
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > len(lines) {
		end = len(lines)
	}
	return lines[start:end]
}

func (lines lines) getSearchResultLine(term string, searchIndex int) int {
	for i := range lines {
		line := lines[len(lines)-i-1]
		if strings.Contains(line, term) {
			if searchIndex <= 0 {
				return i
			}
			searchIndex--
		}
	}
	return -1
}

func (lines lines) highlightMatches(term string) lines {
	if term == "" {
		return lines
	}
	var highlightedLines = make([]string, len(lines))
	for i, line := range lines {
		hlTerm := fmt.Sprintf("[%s](bg-yellow,fg-black)", term)
		highlightedLines[i] = strings.Replace(line, term, hlTerm, -1) //hlTerm, -1)
	}
	return highlightedLines
}

// Scroll
func (state AppState) scroll(direction direction, amount int) AppState {
	view := state.getSelectedView()
	file := state.getSelectedFile()
	switch direction {
	case up:
		view.offSet += amount
	case down:
		view.offSet -= amount
	case bottom:
		view.offSet = 0
	}
	if view.offSet > len(file.lines)-view.numVisibleLines(state) {
		view.offSet = len(file.lines) - view.numVisibleLines(state)
	}
	if view.offSet < 0 {
		view.offSet = 0
	}
	state.LogViews[state.selected] = view
	state.StatusBar.Text = strconv.Itoa(state.getSelectedView().offSet)
	return state
}

func anyActiveModifiers(modifiers modifiers, kind modifierType) bool {
	for _, m := range modifiers {
		if m.active && m.kind == kind {
			return true
		}
	}
	return false
}

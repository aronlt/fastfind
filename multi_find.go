package main

type MultiFind interface {
	Match(entries []*Entry, text string)[]*Entry
}

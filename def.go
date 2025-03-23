package kio

type Action int

const (
	ActionTraffic Action = iota
	ActionData
	ActionClose
	ActionShutdown
)

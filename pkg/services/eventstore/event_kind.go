package eventstore

type eventKind uint8

const (
	pageviewEventKind eventKind = iota
	customEventKind
	fileDownloadEventKind
	outboundLinkClickEventKind
	maxEventKind
)

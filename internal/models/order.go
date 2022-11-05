package models

var OrderStatus = struct {
	NEW        string
	PROCESSING string
	INVALID    string
	PROCESSED  string
}{
	NEW:        "NEW",
	PROCESSING: "PROCESSING",
	INVALID:    "INVALID",
	PROCESSED:  "PROCESSED",
}

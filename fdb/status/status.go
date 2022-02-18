package status

import "github.com/digisan/go-generics/str"

const (
	All      = "all"
	Received = "received"
	Applying = "applying"
	Approved = "approved"
	Rejected = "rejected"
	Deleted  = "deleted"
	Unknown  = "unknown"
)

func AllStatus() []string {
	return []string{Received, Applying, Approved, Rejected, Deleted, Unknown}
}

func StatusOK(status string) bool {
	return str.In(status, AllStatus()...)
}

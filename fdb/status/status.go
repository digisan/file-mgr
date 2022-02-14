package status

import "github.com/digisan/go-generics/str"

const (	
	Received = "received"
	Applying = "applying"
	Approved = "approved"
	Rejected = "rejected"
	Deleted  = "deleted"
	Unknown  = "unknown"
)

func AllStatus() []string {
	return []string{Unknown, Received, Applying, Approved, Rejected, Deleted}
}

func StatusOK(status string) bool {
	return str.In(status, AllStatus()...)
}

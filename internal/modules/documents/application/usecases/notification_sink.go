package usecases

import "context"

// ShareNotifier is the documents-module narrow port для emitting a
// notification when a document is shared с another user. The adapter
// lives in cmd/server/main.go (DI seam) so this package stays free of
// cross-module Go imports — satisfying the CLAUDE.md gate "Cross-module
// imports запрещены, только через адаптеры в main.go / DI-точке".
//
// v0.156.0 ADR-5 (#266): pre-fix sharing_usecase.go imported
// notifications/application/usecases directly + fire-and-forget
// goroutine с context.Background() + Russian UI strings — all три
// violations closed by routing через this port.
type ShareNotifier interface {
	NotifyDocumentShared(ctx context.Context, recipientID int64, documentID int64, documentTitle string) error
}

package archive

import "context"

// ArchiveRepository defines the persistence interface for archives.
type ArchiveRepository interface {
	// Save persists an archive to storage.
	Save(ctx context.Context, archive *Archive) error

	// FindByID retrieves an archive by its ID.
	FindByID(ctx context.Context, id ArchiveID) (*Archive, error)

	// FindAll retrieves all archives.
	FindAll(ctx context.Context) ([]*Archive, error)

	// Delete removes an archive from storage.
	Delete(ctx context.Context, id ArchiveID) error

	// SaveEntry persists an archive entry.
	SaveEntry(ctx context.Context, entry *ArchiveEntry) error

	// FindEntryByID retrieves an entry by its ID.
	FindEntryByID(ctx context.Context, id EntryID) (*ArchiveEntry, error)

	// FindEntriesByArchive retrieves all entries for an archive.
	FindEntriesByArchive(ctx context.Context, archiveID ArchiveID) ([]*ArchiveEntry, error)

	// DeleteEntry removes an entry from storage.
	DeleteEntry(ctx context.Context, id EntryID) error
}

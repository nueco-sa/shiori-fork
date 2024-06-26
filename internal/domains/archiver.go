package domains

import (
	"fmt"
	"io"

	"github.com/go-shiori/shiori/internal/archiver"
	"github.com/go-shiori/shiori/internal/core"
	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

type ArchiverDomain struct {
	deps      *dependencies.Dependencies
	archivers map[string]model.Archiver
}

func (d *ArchiverDomain) GenerateBookmarkArchive(book model.BookmarkDTO) (*model.BookmarkDTO, error) {
	content, contentType, err := core.DownloadBookmark(book.URL)
	if err != nil {
		return nil, fmt.Errorf("error downloading url: %s", err)
	}

	return d.ProcessBookmarkArchive(content, contentType, book)
}

func (d *ArchiverDomain) GenerateBookmarkEbook(request model.EbookProcessRequest) error {
	_, err := core.GenerateEbook(d.deps, request)
	if err != nil {
		return fmt.Errorf("error generating ebook: %s", err)
	}

	return nil
}

func (d *ArchiverDomain) ProcessBookmarkArchive(content io.ReadCloser, contentType string, book model.BookmarkDTO) (*model.BookmarkDTO, error) {
	for _, archiver := range d.archivers {
		if archiver.Matches(contentType) {
			_, err := archiver.Archive(content, contentType, book)
			if err != nil {
				d.deps.Log.Errorf("Error archiving bookmark with archviver: %s", err)
				continue
			}
			return &book, nil
		}
	}

	return nil, fmt.Errorf("no archiver found for content type: %s", contentType)
}

func (d *ArchiverDomain) GetBookmarkArchiveFile(book *model.BookmarkDTO, resourcePath string) (*model.ArchiveFile, error) {
	archiver, err := d.GetArchiver(book.Archiver)
	if err != nil {
		return nil, err
	}

	archiveFile, err := archiver.GetArchiveFile(*book, resourcePath)
	if err != nil {
		return nil, fmt.Errorf("error getting archive file: %w", err)
	}

	return archiveFile, nil
}

func (d *ArchiverDomain) GetArchiver(name string) (model.Archiver, error) {
	archiver, ok := d.archivers[name]
	if !ok {
		return nil, fmt.Errorf("archiver %s not found", name)
	}
	return archiver, nil
}

func NewArchiverDomain(deps *dependencies.Dependencies) *ArchiverDomain {
	archivers := map[string]model.Archiver{
		model.ArchiverPDF:  archiver.NewPDFArchiver(deps),
		model.ArchiverWARC: archiver.NewWARCArchiver(deps),
	}
	return &ArchiverDomain{
		deps:      deps,
		archivers: archivers,
	}
}

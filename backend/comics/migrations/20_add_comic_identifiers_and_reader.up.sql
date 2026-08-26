-- Comic identifiers (single-issue vs graphic-novel vs periodical) and
-- production-reader fields. A single comic carries exactly one identifier
-- type; the others are NULL.
ALTER TABLE comics
    ADD COLUMN isbn TEXT,
    ADD COLUMN upc  TEXT,
    ADD COLUMN issn TEXT;

-- Reader support: paging direction, denormalized page count, per-page
-- dimensions (parallel to page_keys, aligned by index), and archive MIME type.
ALTER TABLE comics
    ADD COLUMN reading_direction TEXT NOT NULL DEFAULT 'ltr'
        CHECK (reading_direction IN ('ltr', 'rtl')),
    ADD COLUMN page_count INT NOT NULL DEFAULT 0,
    ADD COLUMN page_dimensions JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN archive_mimetype TEXT;

COMMENT ON COLUMN comics.isbn IS 'ISBN-13 for graphic novels only (NULL for single issues).';
COMMENT ON COLUMN comics.upc  IS 'UPC barcode for single-issue comics only (NULL for graphic novels).';
COMMENT ON COLUMN comics.issn IS 'ISSN for periodical/series distribution, when applicable.';
COMMENT ON COLUMN comics.reading_direction IS 'Reader paging order: ltr (western) or rtl (manga).';
COMMENT ON COLUMN comics.page_count IS 'Denormalized number of pages (length of page_keys) for list/header rendering.';
COMMENT ON COLUMN comics.page_dimensions IS 'Parallel array [{width,height}] aligned by index with page_keys, for layout-shift-free rendering.';
COMMENT ON COLUMN comics.archive_mimetype IS 'MIME type of file_key (e.g. application/vnd.comicbook+zip, application/pdf) for download headers.';

CREATE INDEX idx_comics_isbn ON comics(isbn);
CREATE INDEX idx_comics_upc  ON comics(upc);
CREATE INDEX idx_comics_issn ON comics(issn);
CREATE INDEX idx_comics_page_count ON comics(page_count);

package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ExpectedMimetype = "application/epub+zip"
	MimetypeFilename = "mimetype"
	ContainerXMLPath = "META-INF/container.xml"
)

func createMinimalValidEPUB() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(containerXML))

	contentOPF := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:uuid:12345678-1234-1234-1234-123456789012</dc:identifier>
    <dc:title>Minimal Test EPUB</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="content" href="content.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="content"/>
  </spine>
</package>`

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(contentOPF))

	navXHTML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="content.xhtml">Chapter 1</a></li>
    </ol>
  </nav>
</body>
</html>`

	navWriter, _ := zipWriter.Create("OEBPS/nav.xhtml")
	navWriter.Write([]byte(navXHTML))

	contentXHTML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Chapter 1</title>
</head>
<body>
  <h1>Chapter 1</h1>
  <p>This is the content.</p>
</body>
</html>`

	contentWriter, _ := zipWriter.Create("OEBPS/content.xhtml")
	contentWriter.Write([]byte(contentXHTML))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBWithInvalidMimetype() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte("application/zip"))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBMimetypeNotFirst() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	dummyWriter, _ := zipWriter.Create("dummy.txt")
	dummyWriter.Write([]byte("dummy"))

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBMimetypeCompressed() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.Create(MimetypeFilename)
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBNoContainer() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBInvalidContainerXML() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<invalid xml`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBNoRootfile() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
  </rootfiles>
</container>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBInvalidOPF() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(`<invalid opf xml`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBMissingTitle() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:uuid:12345678-1234-1234-1234-123456789012</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBMissingIdentifier() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBMissingLanguage() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:uuid:12345678-1234-1234-1234-123456789012</dc:identifier>
    <dc:title>Test</dc:title>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBMissingModified() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:uuid:12345678-1234-1234-1234-123456789012</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBMissingNavDocument() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:uuid:12345678-1234-1234-1234-123456789012</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="content" href="content.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="content"/>
  </spine>
</package>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBInvalidNavDocument() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:uuid:12345678-1234-1234-1234-123456789012</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`))

	navWriter, _ := zipWriter.Create("OEBPS/nav.xhtml")
	navWriter.Write([]byte(`<?xml version="1.0"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Nav</title></head>
<body>
  <p>No nav element</p>
</body>
</html>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createEPUBInvalidContentDocument() []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:uuid:12345678-1234-1234-1234-123456789012</dc:identifier>
    <dc:title>Test</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="content" href="content.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="content"/>
  </spine>
</package>`))

	navWriter, _ := zipWriter.Create("OEBPS/nav.xhtml")
	navWriter.Write([]byte(`<?xml version="1.0"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head><title>Nav</title></head>
<body>
  <nav epub:type="toc">
    <ol><li><a href="content.xhtml">Content</a></li></ol>
  </nav>
</body>
</html>`))

	contentWriter, _ := zipWriter.Create("OEBPS/content.xhtml")
	contentWriter.Write([]byte(`<html><body><p>Missing DOCTYPE and namespace</p></body></html>`))

	zipWriter.Close()
	return buf.Bytes()
}

func createNotZipFile() []byte {
	return []byte("This is not a ZIP file at all")
}

func createCorruptZip() []byte {
	validZip := createMinimalValidEPUB()
	if len(validZip) > 100 {
		return validZip[:len(validZip)-50]
	}
	return validZip
}

func createLargeEPUB(numChapters int) []byte {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, _ := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	mimetypeWriter.Write([]byte(ExpectedMimetype))

	containerWriter, _ := zipWriter.Create(ContainerXMLPath)
	containerWriter.Write([]byte(`<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	var manifestItems string
	var spineItems string
	var navLinks string

	for i := 1; i <= numChapters; i++ {
		manifestItems += fmt.Sprintf(`    <item id="chapter%d" href="chapter%d.xhtml" media-type="application/xhtml+xml"/>
`, i, i)
		spineItems += fmt.Sprintf(`    <itemref idref="chapter%d"/>
`, i)
		navLinks += fmt.Sprintf(`      <li><a href="chapter%d.xhtml">Chapter %d</a></li>
`, i, i)
	}

	opfContent := fmt.Sprintf(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="uid">urn:uuid:large-test-epub</dc:identifier>
    <dc:title>Large Test EPUB</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
%s  </manifest>
  <spine>
%s  </spine>
</package>`, manifestItems, spineItems)

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(opfContent))

	navContent := fmt.Sprintf(`<?xml version="1.0"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head><title>Navigation</title></head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
%s    </ol>
  </nav>
</body>
</html>`, navLinks)

	navWriter, _ := zipWriter.Create("OEBPS/nav.xhtml")
	navWriter.Write([]byte(navContent))

	for i := 1; i <= numChapters; i++ {
		chapterContent := fmt.Sprintf(`<?xml version="1.0"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter %d</title></head>
<body>
  <h1>Chapter %d</h1>
  <p>%s</p>
</body>
</html>`, i, i, generateLongText(1000))

		chapterWriter, _ := zipWriter.Create(fmt.Sprintf("OEBPS/chapter%d.xhtml", i))
		chapterWriter.Write([]byte(chapterContent))
	}

	zipWriter.Close()
	return buf.Bytes()
}

func generateLongText(words int) string {
	text := ""
	for i := 0; i < words; i++ {
		text += "Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua "
	}
	return text
}

func main() {
	fixtures := map[string][]byte{
		"valid/minimal.epub":                    createMinimalValidEPUB(),
		"valid/large_100_chapters.epub":         createLargeEPUB(100),
		"valid/large_500_chapters.epub":         createLargeEPUB(500),
		"invalid/not_zip.epub":                  createNotZipFile(),
		"invalid/corrupt_zip.epub":              createCorruptZip(),
		"invalid/wrong_mimetype.epub":           createEPUBWithInvalidMimetype(),
		"invalid/mimetype_not_first.epub":       createEPUBMimetypeNotFirst(),
		"invalid/mimetype_compressed.epub":      createEPUBMimetypeCompressed(),
		"invalid/no_container.epub":             createEPUBNoContainer(),
		"invalid/invalid_container_xml.epub":    createEPUBInvalidContainerXML(),
		"invalid/no_rootfile.epub":              createEPUBNoRootfile(),
		"invalid/invalid_opf.epub":              createEPUBInvalidOPF(),
		"invalid/missing_title.epub":            createEPUBMissingTitle(),
		"invalid/missing_identifier.epub":       createEPUBMissingIdentifier(),
		"invalid/missing_language.epub":         createEPUBMissingLanguage(),
		"invalid/missing_modified.epub":         createEPUBMissingModified(),
		"invalid/missing_nav_document.epub":     createEPUBMissingNavDocument(),
		"invalid/invalid_nav_document.epub":     createEPUBInvalidNavDocument(),
		"invalid/invalid_content_document.epub": createEPUBInvalidContentDocument(),
	}

	baseDir := "."
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}

	for filename, data := range fixtures {
		filePath := filepath.Join(baseDir, filename)
		dir := filepath.Dir(filePath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create directory %s: %v\n", dir, err)
			continue
		}

		if err := os.WriteFile(filePath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", filename, err)
			continue
		}
		fmt.Printf("Created %s (%d bytes)\n", filePath, len(data))
	}

	fmt.Println("\nAll EPUB fixtures generated successfully")
}

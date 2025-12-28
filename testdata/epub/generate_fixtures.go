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

func createValidEPUB() []byte {
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
    <dc:title>Test EPUB</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`

	opfWriter, _ := zipWriter.Create("OEBPS/content.opf")
	opfWriter.Write([]byte(contentOPF))

	navXHTML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Test EPUB</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="nav.xhtml">Start</a></li>
    </ol>
  </nav>
</body>
</html>`

	navWriter, _ := zipWriter.Create("OEBPS/nav.xhtml")
	navWriter.Write([]byte(navXHTML))

	zipWriter.Close()
	return buf.Bytes()
}

func createInvalidEPUB() []byte {
	return []byte("This is not a valid EPUB file")
}

func main() {
	outputDir := "."
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	fixtures := map[string][]byte{
		"valid.epub":   createValidEPUB(),
		"invalid.epub": createInvalidEPUB(),
	}

	for filename, data := range fixtures {
		filePath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", filename, err)
			os.Exit(1)
		}
		fmt.Printf("Created %s\n", filePath)
	}

	fmt.Println("Fixtures generated successfully")
}

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func createMinimalValidPDF() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/Resources <<
/Font <<
/F1 <<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
>>
>>
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Hello World) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000317 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
410
%%EOF
`
	return []byte(pdf)
}

func createPDFNoHeader() []byte {
	return []byte(`This is not a PDF file
1 0 obj
<< /Type /Catalog >>
endobj
xref
0 1
0000000000 65535 f 
trailer
<< /Size 1 >>
startxref
50
%%EOF
`)
}

func createPDFInvalidVersion() []byte {
	return []byte(`%PDF-2.0
1 0 obj
<< /Type /Catalog >>
endobj
xref
0 1
0000000000 65535 f 
trailer
<< /Size 1 >>
startxref
50
%%EOF
`)
}

func createPDFNoEOF() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids []
/Count 0
>>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<<
/Size 3
/Root 1 0 R
>>
startxref
115
`
	return []byte(pdf)
}

func createPDFNoStartxref() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
%%EOF
`
	return []byte(pdf)
}

func createPDFCorruptXref() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids []
/Count 0
>>
endobj
xref
this is not a valid xref table
trailer
<<
/Size 3
/Root 1 0 R
>>
startxref
115
%%EOF
`
	return []byte(pdf)
}

func createPDFNoCatalog() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Info
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
>>
startxref
50
%%EOF
`
	return []byte(pdf)
}

func createPDFInvalidCatalog() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
50
%%EOF
`
	return []byte(pdf)
}

func createPDFTruncatedStream() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj
4 0 obj
<<
/Length 100
>>
stream
BT
/F1 12 Tf
`
	return []byte(pdf)
}

func createPDFMalformedObjects() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R
/Count 1
>>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<<
/Size 3
/Root 1 0 R
>>
startxref
150
%%EOF
`
	return []byte(pdf)
}

func createPDFWithEncryption() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
>>
endobj
4 0 obj
<<
/Filter /Standard
/V 1
/R 2
/O <encrypted>
/U <encrypted>
/P -64
>>
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000200 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
/Encrypt 4 0 R
>>
startxref
300
%%EOF
`
	return []byte(pdf)
}

func createCorruptPDF() []byte {
	validPDF := createMinimalValidPDF()
	if len(validPDF) > 100 {
		return validPDF[:len(validPDF)-100]
	}
	return validPDF
}

func createLargePDF(numPages int) []byte {
	var buf bytes.Buffer

	buf.WriteString("%PDF-1.4\n")

	objectNum := 1
	catalogObj := objectNum
	objectNum++

	pagesObj := objectNum
	objectNum++

	buf.WriteString(fmt.Sprintf("%d 0 obj\n<<\n/Type /Catalog\n/Pages %d 0 R\n>>\nendobj\n", catalogObj, pagesObj))

	pageObjects := make([]int, numPages)
	for i := 0; i < numPages; i++ {
		pageObjects[i] = objectNum
		objectNum++
	}

	buf.WriteString(fmt.Sprintf("%d 0 obj\n<<\n/Type /Pages\n/Kids [", pagesObj))
	for i, pageObj := range pageObjects {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(fmt.Sprintf("%d 0 R", pageObj))
	}
	buf.WriteString(fmt.Sprintf("]\n/Count %d\n>>\nendobj\n", numPages))

	contentObjects := make([]int, numPages)
	for i := 0; i < numPages; i++ {
		contentObjects[i] = objectNum
		objectNum++
	}

	for i := 0; i < numPages; i++ {
		pageContent := fmt.Sprintf(`%d 0 obj
<<
/Type /Page
/Parent %d 0 R
/Resources <<
/Font <<
/F1 <<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
>>
>>
/MediaBox [0 0 612 792]
/Contents %d 0 R
>>
endobj
`, pageObjects[i], pagesObj, contentObjects[i])
		buf.WriteString(pageContent)
	}

	for i := 0; i < numPages; i++ {
		stream := fmt.Sprintf("BT\n/F1 12 Tf\n100 700 Td\n(Page %d of %d) Tj\nET\n", i+1, numPages)
		contentObj := fmt.Sprintf(`%d 0 obj
<<
/Length %d
>>
stream
%s
endstream
endobj
`, contentObjects[i], len(stream), stream)
		buf.WriteString(contentObj)
	}

	xrefStart := buf.Len()
	buf.WriteString(fmt.Sprintf("xref\n0 %d\n", objectNum))
	buf.WriteString("0000000000 65535 f \n")

	buf.WriteString(fmt.Sprintf("trailer\n<<\n/Size %d\n/Root %d 0 R\n>>\n", objectNum, catalogObj))
	buf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefStart))

	return buf.Bytes()
}

func createPDFWithImages() []byte {
	pdf := `%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
/Resources <<
/XObject <<
/Im1 5 0 R
>>
>>
>>
endobj
4 0 obj
<<
/Length 50
>>
stream
q
200 0 0 200 100 400 cm
/Im1 Do
Q
endstream
endobj
5 0 obj
<<
/Type /XObject
/Subtype /Image
/Width 10
/Height 10
/ColorSpace /DeviceRGB
/BitsPerComponent 8
/Length 300
>>
stream
` + strings.Repeat("\x00", 300) + `
endstream
endobj
xref
0 6
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000274 00000 n 
0000000373 00000 n 
trailer
<<
/Size 6
/Root 1 0 R
>>
startxref
600
%%EOF
`
	return []byte(pdf)
}

func createNotPDFFile() []byte {
	return []byte("This is not a PDF file at all. Just plain text.")
}

func main() {
	fixtures := map[string][]byte{
		"valid/minimal.pdf":              createMinimalValidPDF(),
		"valid/with_images.pdf":          createPDFWithImages(),
		"valid/large_100_pages.pdf":      createLargePDF(100),
		"valid/large_1000_pages.pdf":     createLargePDF(1000),
		"edge_cases/large_10mb_plus.pdf": createLargePDF(5000),
		"edge_cases/encrypted.pdf":       createPDFWithEncryption(),
		"invalid/not_pdf.pdf":            createNotPDFFile(),
		"invalid/no_header.pdf":          createPDFNoHeader(),
		"invalid/invalid_version.pdf":    createPDFInvalidVersion(),
		"invalid/no_eof.pdf":             createPDFNoEOF(),
		"invalid/no_startxref.pdf":       createPDFNoStartxref(),
		"invalid/corrupt_xref.pdf":       createPDFCorruptXref(),
		"invalid/no_catalog.pdf":         createPDFNoCatalog(),
		"invalid/invalid_catalog.pdf":    createPDFInvalidCatalog(),
		"invalid/corrupt.pdf":            createCorruptPDF(),
		"invalid/truncated_stream.pdf":   createPDFTruncatedStream(),
		"invalid/malformed_objects.pdf":  createPDFMalformedObjects(),
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

		if err := os.WriteFile(filePath, data, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", filename, err)
			continue
		}
		fmt.Printf("Created %s (%d bytes)\n", filePath, len(data))
	}

	fmt.Println("\nAll PDF fixtures generated successfully")
}
